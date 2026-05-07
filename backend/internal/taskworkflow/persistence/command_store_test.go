package persistence

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OpenNSW/nsw/internal/task/plugin"
)

func TestNewTaskWorkflowCommandStore(t *testing.T) {
	t.Run("rejects nil db", func(t *testing.T) {
		store, err := NewTaskWorkflowCommandStore(nil)

		assert.Nil(t, store)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection cannot be nil")
	})

	t.Run("returns store", func(t *testing.T) {
		db, mock := setupStoreTestDB(t)

		store, err := NewTaskWorkflowCommandStore(db)

		assert.NoError(t, err)
		assert.NotNil(t, store)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTaskWorkflowCommandStoreRegister(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	command := &TaskWorkflowCommand{
		ID:                 "command-1",
		TaskID:             "task-1",
		MacroWorkflowID:    "macro-1",
		MacroRunID:         "macro-run-1",
		NodeID:             "node-1",
		TaskTemplateID:     "template-1",
		SubWorkflowID:      "sub-workflow-1",
		SubWorkflowRunID:   "sub-run-1",
		SignalName:         "simple_form_user_submission",
		Command:            "SUBMIT_FORM",
		AllowedState:       plugin.InProgress,
		AllowedPluginState: []string{"DRAFT"},
		PayloadSchema:      json.RawMessage(`{"type":"object"}`),
		Metadata:           json.RawMessage(`{"source":"test"}`),
		Active:             true,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "task_workflow_commands"`).
		WithArgs(
			command.ID,
			command.TaskID,
			command.MacroWorkflowID,
			command.MacroRunID,
			command.NodeID,
			command.TaskTemplateID,
			command.SubWorkflowID,
			command.SubWorkflowRunID,
			command.SignalName,
			command.Command,
			command.AllowedState,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			command.Active,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.Register(context.Background(), command)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreRegisterRejectsNilCommand(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	err = store.Register(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command cannot be nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreGetActiveCommand(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(time.Minute)

	mock.ExpectQuery(`SELECT \* FROM "task_workflow_commands" WHERE task_id = \$1 AND command = \$2 AND active = \$3 ORDER BY updated_at DESC,"task_workflow_commands"."id" LIMIT \$4`).
		WithArgs("task-1", "SUBMIT_FORM", true, 1).
		WillReturnRows(sqlmock.NewRows(commandStoreColumns()).
			AddRow(
				"command-1",
				"task-1",
				"macro-1",
				"macro-run-1",
				"node-1",
				"template-1",
				"sub-workflow-1",
				"sub-run-1",
				"simple_form_user_submission",
				"SUBMIT_FORM",
				plugin.InProgress,
				[]byte(`["DRAFT"]`),
				[]byte(`{"type":"object"}`),
				[]byte(`{"source":"test"}`),
				true,
				createdAt,
				updatedAt,
			))

	command, err := store.GetActiveCommand(context.Background(), "task-1", "SUBMIT_FORM")

	require.NoError(t, err)
	assert.Equal(t, "command-1", command.ID)
	assert.Equal(t, "sub-workflow-1", command.SubWorkflowID)
	assert.Equal(t, "sub-run-1", command.SubWorkflowRunID)
	assert.Equal(t, "simple_form_user_submission", command.SignalName)
	assert.Equal(t, plugin.InProgress, command.AllowedState)
	assert.Equal(t, []string{"DRAFT"}, command.AllowedPluginState)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreGetActiveSubWorkflowBySignal(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(time.Minute)

	mock.ExpectQuery(`SELECT \* FROM "task_workflow_commands" WHERE task_id = \$1 AND signal_name = \$2 AND active = \$3 ORDER BY updated_at DESC,"task_workflow_commands"."id" LIMIT \$4`).
		WithArgs("task-1", "simple_form_user_submission", true, 1).
		WillReturnRows(sqlmock.NewRows(commandStoreColumns()).
			AddRow(
				"command-1",
				"task-1",
				"macro-1",
				"macro-run-1",
				"node-1",
				"template-1",
				"sub-workflow-1",
				"sub-run-1",
				"simple_form_user_submission",
				"SUBMIT_FORM",
				plugin.InProgress,
				[]byte(`["DRAFT"]`),
				[]byte(`{"type":"object"}`),
				[]byte(`{"source":"test"}`),
				true,
				createdAt,
				updatedAt,
			))

	command, err := store.GetActiveSubWorkflowBySignal(
		context.Background(),
		"task-1",
		"simple_form_user_submission",
	)

	require.NoError(t, err)
	assert.Equal(t, "command-1", command.ID)
	assert.Equal(t, "task-1", command.TaskID)
	assert.Equal(t, "sub-workflow-1", command.SubWorkflowID)
	assert.Equal(t, "sub-run-1", command.SubWorkflowRunID)
	assert.Equal(t, "simple_form_user_submission", command.SignalName)
	assert.Equal(t, "SUBMIT_FORM", command.Command)
	assert.Equal(t, plugin.InProgress, command.AllowedState)
	assert.Equal(t, []string{"DRAFT"}, command.AllowedPluginState)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreListActiveByTaskID(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT \* FROM "task_workflow_commands" WHERE task_id = \$1 AND active = \$2 ORDER BY created_at ASC`).
		WithArgs("task-1", true).
		WillReturnRows(sqlmock.NewRows(commandStoreColumns()).
			AddRow("command-1", "task-1", "macro-1", "", "node-1", "template-1", "sub-1", "", "signal", "SAVE_AS_DRAFT", plugin.InProgress, []byte(`["INITIALIZED"]`), nil, []byte(`{}`), true, now, now).
			AddRow("command-2", "task-1", "macro-1", "", "node-1", "template-1", "sub-1", "", "signal", "SUBMIT_FORM", plugin.InProgress, []byte(`["DRAFT"]`), nil, []byte(`{}`), true, now, now))

	commands, err := store.ListActiveByTaskID(context.Background(), "task-1")

	require.NoError(t, err)
	assert.Len(t, commands, 2)
	assert.Equal(t, "SAVE_AS_DRAFT", commands[0].Command)
	assert.Equal(t, "SUBMIT_FORM", commands[1].Command)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreDeactivateByTaskID(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "task_workflow_commands" SET "active"=\$1,"updated_at"=\$2 WHERE task_id = \$3 AND active = \$4`).
		WithArgs(false, sqlmock.AnyArg(), "task-1", true).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err = store.DeactivateByTaskID(context.Background(), "task-1")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskWorkflowCommandStoreReplaceActiveByTaskID(t *testing.T) {
	db, mock := setupStoreTestDB(t)
	store, err := NewTaskWorkflowCommandStore(db)
	require.NoError(t, err)

	commands := []TaskWorkflowCommand{
		{
			ID:              "command-1",
			Command:         "SUBMIT_FORM",
			MacroWorkflowID: "macro-1",
			NodeID:          "node-1",
			TaskTemplateID:  "template-1",
			SubWorkflowID:   "sub-1",
			SignalName:      "simple_form_user_submission",
			AllowedState:    plugin.InProgress,
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "task_workflow_commands" SET "active"=\$1,"updated_at"=\$2 WHERE task_id = \$3 AND active = \$4`).
		WithArgs(false, sqlmock.AnyArg(), "task-1", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO "task_workflow_commands"`).
		WithArgs(
			"command-1",
			"task-1",
			"macro-1",
			"",
			"node-1",
			"template-1",
			"sub-1",
			"",
			"simple_form_user_submission",
			"SUBMIT_FORM",
			plugin.InProgress,
			nil,
			nil,
			sqlmock.AnyArg(),
			true,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.ReplaceActiveByTaskID(context.Background(), "task-1", commands)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func commandStoreColumns() []string {
	return []string{
		"id",
		"task_id",
		"macro_workflow_id",
		"macro_run_id",
		"node_id",
		"task_template_id",
		"sub_workflow_id",
		"sub_workflow_run_id",
		"signal_name",
		"command",
		"allowed_state",
		"allowed_plugin_state",
		"payload_schema",
		"metadata",
		"active",
		"created_at",
		"updated_at",
	}
}
