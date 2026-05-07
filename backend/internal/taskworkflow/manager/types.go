package manager

import (
	"context"
	"encoding/json"
)

// ActionExecutor defines the interface for executing business logic for a task action.
type ActionExecutor func(ctx context.Context, taskId string, inputs map[string]any, config json.RawMessage) (map[string]any, error)

// InitTaskConfig represents the configuration for initializing a task-level micro-workflow.
type InitTaskConfig struct {
	RenderSchemaId  string `json:"renderSchemaId"`
	MicroWorkflowId string `json:"microWorkflowId"`
}

// InitTaskInput is the input for starting a new micro-workflow for a task.
type InitTaskInput struct {
	TaskID          string         `json:"taskId"`
	MacroWorkflowID string         `json:"macroWorkflowId"`
	TaskTemplateID  string         `json:"taskTemplateId"`
	Type            string         `json:"type"`
	InitialContext  map[string]any `json:"initialContext"`
	Config          InitTaskConfig `json:"config"`
}

// ExecutionRequest represents a user or external action to be performed on a task.
type ExecutionRequest struct {
	TaskID  string `json:"taskId"`
	Action  string `json:"action"`
	Content any    `json:"content,omitempty"`
}

// AtomicActionType identifies the type of an atomic business task (node) in the micro-workflow.
type AtomicActionType string

const (
	ActionDataSubmission    AtomicActionType = "DATA_SUBMISSION"
	ActionPaymentInitiation AtomicActionType = "PAYMENT_INITIATION"
	ActionPaymentWebhook    AtomicActionType = "PAYMENT_WEBHOOK"
	ActionStatusUpdate      AtomicActionType = "STATUS_UPDATE"
)

// CommandMetadata stores additional info about a registered WAIT command.
type CommandMetadata struct {
	ExecutionType AtomicActionType `json:"executionType"`
	SchemaID      string           `json:"schemaId,omitempty"`
}
