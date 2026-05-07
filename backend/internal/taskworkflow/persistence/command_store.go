package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/OpenNSW/nsw/internal/task/plugin"
)

// TaskWorkflowCommand describes one command currently accepted by a task
// workflow and the Temporal signal route that should receive it.
type TaskWorkflowCommand struct {
	ID                 string          `gorm:"type:text;column:id;not null;primaryKey" json:"id"`
	TaskID             string          `gorm:"type:text;column:task_id;not null;index" json:"taskId"`
	MacroWorkflowID    string          `gorm:"type:text;column:macro_workflow_id;not null;index" json:"macroWorkflowId"`
	MacroRunID         string          `gorm:"type:text;column:macro_run_id" json:"macroRunId,omitempty"`
	NodeID             string          `gorm:"type:text;column:node_id;not null;index" json:"nodeId"`
	TaskTemplateID     string          `gorm:"type:text;column:task_template_id;not null;index" json:"taskTemplateId"`
	SubWorkflowID      string          `gorm:"type:text;column:sub_workflow_id;not null;index" json:"subWorkflowId"`
	SubWorkflowRunID   string          `gorm:"type:text;column:sub_workflow_run_id" json:"subWorkflowRunId,omitempty"`
	SignalName         string          `gorm:"type:text;column:signal_name;not null" json:"signalName"`
	Command            string          `gorm:"type:text;column:command;not null;index" json:"command"`
	AllowedState       plugin.State    `gorm:"type:varchar(50);column:allowed_state" json:"allowedState,omitempty"`
	AllowedPluginState []string        `gorm:"type:jsonb;column:allowed_plugin_state;serializer:json" json:"allowedPluginState,omitempty"`
	PayloadSchema      json.RawMessage `gorm:"type:jsonb;column:payload_schema;serializer:json" json:"payloadSchema,omitempty"`
	Metadata           json.RawMessage `gorm:"type:jsonb;column:metadata;serializer:json" json:"metadata,omitempty"`
	Active             bool            `gorm:"type:boolean;column:active;not null;default:true;index" json:"active"`
	CreatedAt          time.Time       `gorm:"type:timestamptz;column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time       `gorm:"type:timestamptz;column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (TaskWorkflowCommand) TableName() string {
	return "task_workflow_commands"
}

// CommandStore is the routing registry used by task workflow managers and
// subtask activity services. It records which commands a task currently accepts
// and where those commands should be signaled.
type CommandStore interface {
	Register(ctx context.Context, command *TaskWorkflowCommand) error
	RegisterMany(ctx context.Context, commands []TaskWorkflowCommand) error
	ReplaceActiveByTaskID(ctx context.Context, taskID string, commands []TaskWorkflowCommand) error
	GetActiveCommand(ctx context.Context, taskID, command string) (*TaskWorkflowCommand, error)
	GetActiveSubWorkflowBySignal(ctx context.Context, taskID, signalName string) (*TaskWorkflowCommand, error)
	ListActiveByTaskID(ctx context.Context, taskID string) ([]TaskWorkflowCommand, error)
	DeactivateByTaskID(ctx context.Context, taskID string) error
	DeactivateBySubWorkflowID(ctx context.Context, subWorkflowID string) error
	DeactivateCommand(ctx context.Context, taskID, command string) error
}

type taskWorkflowCommandStore struct {
	db *gorm.DB
}

func NewTaskWorkflowCommandStore(db *gorm.DB) (CommandStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &taskWorkflowCommandStore{db: db}, nil
}

func (s *taskWorkflowCommandStore) Register(ctx context.Context, command *TaskWorkflowCommand) error {
	if command == nil {
		return fmt.Errorf("command cannot be nil")
	}
	prepareTaskWorkflowCommand(command)
	return s.db.WithContext(ctx).Create(command).Error
}

func (s *taskWorkflowCommandStore) RegisterMany(ctx context.Context, commands []TaskWorkflowCommand) error {
	if len(commands) == 0 {
		return nil
	}

	for i := range commands {
		prepareTaskWorkflowCommand(&commands[i])
	}

	return s.db.WithContext(ctx).Create(&commands).Error
}

func (s *taskWorkflowCommandStore) ReplaceActiveByTaskID(
	ctx context.Context,
	taskID string,
	commands []TaskWorkflowCommand,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&TaskWorkflowCommand{}).
			Where("task_id = ? AND active = ?", taskID, true).
			Update("active", false).Error; err != nil {
			return err
		}

		if len(commands) == 0 {
			return nil
		}

		for i := range commands {
			if commands[i].TaskID == "" {
				commands[i].TaskID = taskID
			}
			prepareTaskWorkflowCommand(&commands[i])
		}

		return tx.Create(&commands).Error
	})
}

func (s *taskWorkflowCommandStore) GetActiveCommand(
	ctx context.Context,
	taskID string,
	command string,
) (*TaskWorkflowCommand, error) {
	var out TaskWorkflowCommand
	if err := s.db.WithContext(ctx).
		Where("task_id = ? AND command = ? AND active = ?", taskID, command, true).
		Order("updated_at DESC").
		First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *taskWorkflowCommandStore) GetActiveSubWorkflowBySignal(
	ctx context.Context,
	taskID string,
	signalName string,
) (*TaskWorkflowCommand, error) {
	var command TaskWorkflowCommand
	if err := s.db.WithContext(ctx).
		Where("task_id = ? AND signal_name = ? AND active = ?", taskID, signalName, true).
		Order("updated_at DESC").
		First(&command).Error; err != nil {
		return nil, err
	}

	return &command, nil
}

func (s *taskWorkflowCommandStore) ListActiveByTaskID(
	ctx context.Context,
	taskID string,
) ([]TaskWorkflowCommand, error) {
	var commands []TaskWorkflowCommand
	if err := s.db.WithContext(ctx).
		Where("task_id = ? AND active = ?", taskID, true).
		Order("created_at ASC").
		Find(&commands).Error; err != nil {
		return nil, err
	}
	return commands, nil
}

func (s *taskWorkflowCommandStore) DeactivateByTaskID(ctx context.Context, taskID string) error {
	return s.db.WithContext(ctx).
		Model(&TaskWorkflowCommand{}).
		Where("task_id = ? AND active = ?", taskID, true).
		Update("active", false).Error
}

func (s *taskWorkflowCommandStore) DeactivateBySubWorkflowID(ctx context.Context, subWorkflowID string) error {
	return s.db.WithContext(ctx).
		Model(&TaskWorkflowCommand{}).
		Where("sub_workflow_id = ? AND active = ?", subWorkflowID, true).
		Update("active", false).Error
}

func (s *taskWorkflowCommandStore) DeactivateCommand(ctx context.Context, taskID, command string) error {
	return s.db.WithContext(ctx).
		Model(&TaskWorkflowCommand{}).
		Where("task_id = ? AND command = ? AND active = ?", taskID, command, true).
		Update("active", false).Error
}

func prepareTaskWorkflowCommand(command *TaskWorkflowCommand) {
	if command.ID == "" {
		command.ID = uuid.NewString()
	}
	command.Active = true
	if len(command.Metadata) == 0 {
		command.Metadata = json.RawMessage(`{}`)
	}
}
