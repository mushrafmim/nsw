package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	workflowmanager "github.com/OpenNSW/go-temporal-workflow"
	"github.com/OpenNSW/nsw/internal/task/plugin"
	"github.com/OpenNSW/nsw/internal/taskworkflow/persistence"
	"github.com/OpenNSW/nsw/internal/workflow/service"
	"go.temporal.io/sdk/client"
)

// TaskWorkflowManager orchestrates task-level micro-workflows using go-temporal-workflow.
type TaskWorkflowManager interface {
	InitTask(ctx context.Context, input InitTaskInput) error
	ExecuteAction(ctx context.Context, req ExecutionRequest) error
}

type taskWorkflowManager struct {
	temporalManager workflowmanager.TemporalManager
	store           persistence.Store
	commandStore    persistence.CommandStore
	templateService service.TemplateProvider
	executors       map[string]ActionExecutor
	serviceBaseURL  string
}

// NewTaskWorkflowManager creates a new TaskWorkflowManager.
func NewTaskWorkflowManager(
	tc client.Client,
	taskQueue string,
	store persistence.Store,
	commandStore persistence.CommandStore,
	templateService service.TemplateProvider,
	activationHandler workflowmanager.TaskActivationHandler,
	completionHandler workflowmanager.WorkflowCompletionHandler,
	serviceBaseURL string,
) TaskWorkflowManager {
	tm := workflowmanager.NewTemporalManager(tc, taskQueue, activationHandler, completionHandler)
	return &taskWorkflowManager{
		temporalManager: tm,
		store:           store,
		commandStore:    commandStore,
		templateService: templateService,
		executors:       make(map[string]ActionExecutor),
		serviceBaseURL:  serviceBaseURL,
	}
}

func (m *taskWorkflowManager) RegisterExecutor(action string, executor ActionExecutor) {
	m.executors[action] = executor
}

func (m *taskWorkflowManager) InitTask(ctx context.Context, input InitTaskInput) error {
	// 1. Resolve the Micro-Workflow Definition from the database
	if input.Config.MicroWorkflowId == "" {
		return fmt.Errorf("microWorkflowId is required in task configuration")
	}

	wt, err := m.templateService.GetWorkflowTemplateByIDV2(ctx, input.Config.MicroWorkflowId)
	if err != nil {
		return fmt.Errorf("failed to fetch micro-workflow template %s: %w", input.Config.MicroWorkflowId, err)
	}
	if wt == nil {
		return fmt.Errorf("micro-workflow template %s not found", input.Config.MicroWorkflowId)
	}

	// 2. Prepare __runtime metadata for the task environment
	runtime := map[string]any{
		"taskId":         input.TaskID,
		"consignmentId":  input.MacroWorkflowID,
		"serviceBaseURL": m.serviceBaseURL,
	}

	if input.InitialContext == nil {
		input.InitialContext = make(map[string]any)
	}
	input.InitialContext["__runtime"] = runtime

	initialDataBytes, _ := json.Marshal(input.InitialContext)

	// 3. Create a record in task_workflow_tasks table
	taskRecord := &persistence.TaskWorkflowTask{
		TaskID:          input.TaskID,
		MacroWorkflowID: input.MacroWorkflowID,
		TaskTemplateID:  input.TaskTemplateID,
		State:           plugin.Initialized,
		Data:            json.RawMessage(initialDataBytes),
	}

	if err := m.store.Create(taskRecord); err != nil {
		return fmt.Errorf("failed to create task workflow record: %w", err)
	}

	// 4. Start the Micro-Workflow using the resolved definition
	if err := m.temporalManager.StartWorkflow(ctx, input.TaskID, wt.WorkflowDefinition, input.InitialContext); err != nil {
		return fmt.Errorf("failed to start micro-workflow: %w", err)
	}

	slog.InfoContext(ctx, "started micro-workflow for task", "taskId", input.TaskID, "microWorkflowId", input.Config.MicroWorkflowId)
	return nil
}

func (m *taskWorkflowManager) ExecuteAction(ctx context.Context, req ExecutionRequest) error {
	// 1. Lookup active command for this task
	command, err := m.commandStore.GetActiveCommand(ctx, req.TaskID, req.Action)
	if err != nil {
		return fmt.Errorf("action %s not allowed for task %s: %w", req.Action, req.TaskID, err)
	}

	// 2. Fetch the node template to get configuration
	template, err := m.templateService.GetWorkflowNodeTemplateByID(ctx, command.TaskTemplateID)
	if err != nil {
		return fmt.Errorf("failed to fetch task template %s: %w", command.TaskTemplateID, err)
	}

	// 3. Prepare inputs for the executor (User's content)
	inputs := make(map[string]any)
	if req.Content != nil {
		if m, ok := req.Content.(map[string]any); ok {
			inputs = m
		} else {
			// Try to handle case where content is a single value or different type
			inputs["payload"] = req.Content
		}
	}

	// Retrieve __runtime from task data or fallback to reconstruction
	task, err := m.store.GetByTaskID(req.TaskID)
	if err == nil && len(task.Data) > 0 {
		var taskData map[string]any
		if err := json.Unmarshal(task.Data, &taskData); err == nil {
			if rt, ok := taskData["__runtime"]; ok {
				inputs["__runtime"] = rt
			}
		}
	}

	// Fallback for legacy tasks or if missing in Data
	if _, ok := inputs["__runtime"]; !ok {
		inputs["__runtime"] = map[string]any{
			"taskId":         command.TaskID,
			"consignmentId":  command.MacroWorkflowID,
			"serviceBaseURL": m.serviceBaseURL,
		}
	}

	// 4. Read command metadata
	var cmdMeta struct {
		PluginState string `json:"pluginState"`
		WriteTo     string `json:"writeTo"`
	}
	if len(command.Metadata) > 0 {
		_ = json.Unmarshal(command.Metadata, &cmdMeta)
	}

	// 5. Execute business logic (AUTO nodes with a registered executor)
	var result map[string]any
	if executor, ok := m.executors[req.Action]; ok {
		res, err := executor(ctx, req.TaskID, inputs, template.Config)
		if err != nil {
			return fmt.Errorf("business logic execution failed for action %s: %w", req.Action, err)
		}
		result = res
	}

	// 6. Build and apply task context update
	update := make(map[string]any)

	if cmdMeta.WriteTo != "" {
		update[cmdMeta.WriteTo] = req.Content
	}
	for k, v := range result {
		update[k] = v
	}
	if cmdMeta.PluginState != "" {
		update["pluginState"] = cmdMeta.PluginState
	}

	if len(update) > 0 {
		if err := m.updateTaskContext(ctx, req.TaskID, update); err != nil {
			return fmt.Errorf("failed to update task context: %w", err)
		}
	}

	// 9. Resume the Micro-Workflow
	if err := m.temporalManager.TaskDone(ctx, command.SubWorkflowID, command.SubWorkflowRunID, command.NodeID, update); err != nil {
		return fmt.Errorf("failed to signal micro-workflow: %w", err)
	}

	slog.InfoContext(ctx, "executed action and resumed micro-workflow", "taskId", req.TaskID, "action", req.Action)
	return nil
}

func (m *taskWorkflowManager) updateTaskContext(ctx context.Context, taskId string, update map[string]any) error {
	// Merge the new key-values into the jsonb Data column.
	task, err := m.store.GetByTaskID(taskId)
	if err != nil {
		return err
	}

	var currentData map[string]any
	if len(task.Data) > 0 {
		_ = json.Unmarshal(task.Data, &currentData)
	}
	if currentData == nil {
		currentData = make(map[string]any)
	}

	for k, v := range update {
		currentData[k] = v
	}

	newData, _ := json.Marshal(currentData)
	return m.store.UpdateData(taskId, newData)
}
