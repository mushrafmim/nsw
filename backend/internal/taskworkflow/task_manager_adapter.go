package taskworkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/OpenNSW/nsw/internal/payments"
	taskmanager "github.com/OpenNSW/nsw/internal/task/manager"
	"github.com/OpenNSW/nsw/internal/task/plugin"
	workflowmanager "github.com/OpenNSW/nsw/internal/taskworkflow/manager"
	persistence2 "github.com/OpenNSW/nsw/internal/taskworkflow/persistence"
	"github.com/OpenNSW/nsw/internal/workflow/service"
	"github.com/OpenNSW/nsw/pkg/remote"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type taskManagerAdapter struct {
	mu                    sync.RWMutex
	workflowManager       workflowmanager.TaskWorkflowManager
	store                 persistence2.Store
	templateProvider      service.TemplateProvider
	close                 func() error
	workflowDoneHandler   taskmanager.WorkflowDoneHandler
	workflowUpdateHandler taskmanager.WorkflowUpdateHandler
}

func WireTaskManagerAsLegacy(
	c client.Client,
	db *gorm.DB,
	wtp service.TemplateProvider,
	ps payments.PaymentService,
	rm *remote.Manager,
) taskmanager.TaskManager {
	store, err := persistence2.NewTaskWorkflowStore(db)
	if err != nil {
		panic(err)
	}

	adapter := &taskManagerAdapter{}
	wm, cleanup := workflowmanager.WireTaskManagerWithCleanup(c, db, wtp, ps, rm, adapter.handleMicroWorkflowCompletion)
	adapter.workflowManager = wm
	adapter.store = store
	adapter.templateProvider = wtp
	adapter.close = cleanup

	return adapter
}

func (a *taskManagerAdapter) InitTask(
	ctx context.Context,
	request taskmanager.InitTaskRequest,
) (*taskmanager.InitTaskResponse, error) {
	if request.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	var cfg workflowmanager.InitTaskConfig

	if err := json.Unmarshal(request.Config, &cfg); err != nil {
		return nil, err
	}

	if err := a.workflowManager.InitTask(ctx, workflowmanager.InitTaskInput{
		TaskID:          request.TaskID,
		MacroWorkflowID: request.WorkflowID,
		TaskTemplateID:  request.WorkflowNodeTemplateID,
		Type:            string(request.Type),
		InitialContext:  request.GlobalState,
		Config:          cfg,
	}); err != nil {
		return nil, err
	}

	return &taskmanager.InitTaskResponse{Success: true}, nil
}

func (a *taskManagerAdapter) ExecuteTask(
	ctx context.Context,
	req taskmanager.ExecuteTaskRequest,
) (*plugin.ExecutionResponse, error) {
	if req.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if req.Payload == nil {
		return nil, fmt.Errorf("payload is required")
	}

	if err := a.workflowManager.ExecuteAction(ctx, workflowmanager.ExecutionRequest{
		TaskID:  req.TaskID,
		Action:  req.Payload.Action,
		Content: req.Payload.Content,
	}); err != nil {
		return nil, err
	}

	return &plugin.ExecutionResponse{
		ApiResponse: &plugin.ApiResponse{
			Success: true,
			Data: map[string]any{
				"accepted": true,
			},
		},
	}, nil
}

func (a *taskManagerAdapter) GetTaskRenderInfo(ctx context.Context, taskID string) (*plugin.ApiResponse, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}

	task, err := a.store.GetByTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task %s not found: %w", taskID, err)
	}

	template, err := a.templateProvider.GetWorkflowNodeTemplateByID(ctx, task.TaskTemplateID)
	if err != nil {
		return nil, fmt.Errorf("get workflow node template %s: %w", task.TaskTemplateID, err)
	}

	var content any
	if len(task.Data) > 0 {
		if err := json.Unmarshal(task.Data, &content); err != nil {
			return nil, fmt.Errorf("parse task data for %s: %w", taskID, err)
		}
	}

	pluginState := ""
	if contentMap, ok := content.(map[string]any); ok {
		if value, ok := contentMap["pluginState"].(string); ok {
			pluginState = value
		}
	}

	return &plugin.ApiResponse{
		Success: true,
		Data: plugin.GetRenderInfoResponse{
			Type:        plugin.Type(template.Type),
			State:       task.State,
			PluginState: pluginState,
			Content:     content,
		},
	}, nil
}

func (a *taskManagerAdapter) RegisterUpstreamDoneCallback(callback taskmanager.WorkflowDoneHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workflowDoneHandler = callback
}

func (a *taskManagerAdapter) RegisterUpstreamUpdateCallback(callback taskmanager.WorkflowUpdateHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.workflowUpdateHandler = callback
}

func (a *taskManagerAdapter) Close() error {
	if a == nil || a.close == nil {
		return nil
	}
	return a.close()
}

func (a *taskManagerAdapter) handleMicroWorkflowCompletion(
	ctx context.Context,
	task *persistence2.TaskWorkflowTask,
	finalWorkflowVariables map[string]any,
) error {
	a.mu.RLock()
	callback := a.workflowDoneHandler
	a.mu.RUnlock()

	if callback == nil {
		return nil
	}

	callback(ctx, task.MacroWorkflowID, task.TaskID, finalWorkflowVariables)
	return nil
}
