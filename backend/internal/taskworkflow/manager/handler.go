package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	workflowmanager "github.com/OpenNSW/go-temporal-workflow"
	"github.com/OpenNSW/nsw/internal/taskworkflow/persistence"
)

// handleAtomicTaskActivation is the core logic for executing a node within a Micro-Workflow.
func handleAtomicTaskActivation(
	ctx context.Context,
	m *taskWorkflowManager,
	payload workflowmanager.TaskPayload,
) error {
	// 1. Fetch the atomic task template
	template, err := m.templateService.GetWorkflowNodeTemplateByID(ctx, payload.TaskTemplateID)
	if err != nil {
		return fmt.Errorf("failed to fetch atomic template %s: %w", payload.TaskTemplateID, err)
	}

	// 2. Parse Node Configuration
	var config struct {
		ExecutionType string `json:"executionType"` // "AUTO" or "WAIT"
		Action        string `json:"action"`        // The action name for executors
		Commands      []struct {
			Command       string          `json:"command"`
			PayloadSchema json.RawMessage `json:"payloadSchema"`
		} `json:"commands"`
	}

	if err := json.Unmarshal(template.Config, &config); err != nil {
		return fmt.Errorf("failed to unmarshal atomic task config: %w", err)
	}

	// 3. Dispatch based on ExecutionType
	if config.ExecutionType == "AUTO" {
		return handleAutoTask(ctx, m, template.Config, config.Action, payload)
	}

	// 4. Handle WAIT Task: Register commands in the registry
	var commandsToRegister []persistence.TaskWorkflowCommand
	for _, cmd := range config.Commands {
		commandsToRegister = append(commandsToRegister, persistence.TaskWorkflowCommand{
			TaskID:           payload.WorkflowID, // In Micro-Workflow, WorkflowID is the TaskID
			MacroWorkflowID:  payload.WorkflowID,
			NodeID:           payload.NodeID,
			TaskTemplateID:   payload.TaskTemplateID,
			SubWorkflowID:    payload.WorkflowID,
			SubWorkflowRunID: payload.RunID,
			SignalName:       "TaskDone",
			Command:          cmd.Command,
			AllowedState:     "IN_PROGRESS",
			PayloadSchema:    cmd.PayloadSchema,
			Active:           true,
		})
	}

	if err := m.commandStore.ReplaceActiveByTaskID(ctx, payload.WorkflowID, commandsToRegister); err != nil {
		return fmt.Errorf("failed to register commands for wait task: %w", err)
	}

	slog.InfoContext(ctx, "wait task registered commands", "taskId", payload.WorkflowID, "nodeId", payload.NodeID)
	return nil
}

func handleAutoTask(ctx context.Context, m *taskWorkflowManager, templateConfig json.RawMessage, action string, payload workflowmanager.TaskPayload) error {
	slog.InfoContext(ctx, "executing auto task", "taskId", payload.WorkflowID, "action", action)

	// 1. Execute Business Logic
	var result map[string]any
	if executor, ok := m.executors[action]; ok {
		res, err := executor(ctx, payload.WorkflowID, payload.Inputs, templateConfig)
		if err != nil {
			return fmt.Errorf("auto task business logic failed: %w", err)
		}
		result = res
	}

	// 2. Update Context
	if len(result) > 0 {
		if err := m.updateTaskContext(ctx, payload.WorkflowID, result); err != nil {
			return fmt.Errorf("failed to update task context in auto task: %w", err)
		}
	}

	// 3. Signal Completion to Micro-Workflow
	output := map[string]any{
		"action": action,
		"result": result,
	}

	return m.temporalManager.TaskDone(ctx, payload.WorkflowID, payload.RunID, payload.NodeID, output)
}
