-- Migration: Create task workflow command registry table.

CREATE TABLE IF NOT EXISTS task_workflow_commands (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    macro_workflow_id TEXT NOT NULL,
    macro_run_id TEXT,
    node_id TEXT NOT NULL,
    task_template_id TEXT NOT NULL,
    sub_workflow_id TEXT NOT NULL,
    sub_workflow_run_id TEXT,
    signal_name TEXT NOT NULL,
    command TEXT NOT NULL,
    allowed_state VARCHAR(50),
    allowed_plugin_state JSONB,
    payload_schema JSONB,
    metadata JSONB DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE task_workflow_commands IS 'Active command registry for routing task workflow user actions to subtask workflows';
COMMENT ON COLUMN task_workflow_commands.task_id IS 'Task instance ID visible to the task workflow manager';
COMMENT ON COLUMN task_workflow_commands.sub_workflow_id IS 'Temporal workflow ID that should receive the command signal';
COMMENT ON COLUMN task_workflow_commands.sub_workflow_run_id IS 'Temporal run ID for the sub workflow, when exact run targeting is required';
COMMENT ON COLUMN task_workflow_commands.signal_name IS 'Temporal signal name to send for this command';
COMMENT ON COLUMN task_workflow_commands.command IS 'User/API command accepted by this task in the current state';

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_task_active_command
    ON task_workflow_commands (task_id, active, command);

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_task_active_signal
    ON task_workflow_commands (task_id, active, signal_name);

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_sub_workflow_active
    ON task_workflow_commands (sub_workflow_id, active);

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_macro_workflow_id
    ON task_workflow_commands (macro_workflow_id);

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_node_id
    ON task_workflow_commands (node_id);

CREATE INDEX IF NOT EXISTS idx_task_workflow_commands_task_template_id
    ON task_workflow_commands (task_template_id);
