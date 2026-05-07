package executors

import (
	"context"
	"encoding/json"
	"log/slog"
)

// FormSubmissionExecutor handles the logic when a user submits a form.
func FormSubmissionExecutor(ctx context.Context, taskId string, inputs map[string]any, config json.RawMessage) (map[string]any, error) {
	slog.InfoContext(ctx, "executing user form submission", "taskId", taskId)

	// In a real implementation, you would perform schema validation here using config.

	// We return a structured map so the renderer knows this is the user's data.
	return map[string]any{
		"trader:form": inputs,
		"last_action": "SUBMIT_FORM",
	}, nil
}
