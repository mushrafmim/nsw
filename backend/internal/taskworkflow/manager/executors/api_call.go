package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/OpenNSW/nsw/pkg/jsonform"
	"github.com/OpenNSW/nsw/pkg/jsonutils"
	"github.com/OpenNSW/nsw/pkg/remote"
)

// RemoteApiExecutor handles automated HTTP calls to external systems (OGAs).
// It resolves a JSON request template using placeholders from the task's data context.
func NewRemoteApiExecutor(rm *remote.Manager) func(context.Context, string, map[string]any, json.RawMessage) (map[string]any, error) {
	return func(ctx context.Context, taskId string, inputs map[string]any, config json.RawMessage) (map[string]any, error) {
		slog.InfoContext(ctx, "executing automated OGA API call", "taskId", taskId)

		// 1. Parse Node Configuration
		var nodeConfig struct {
			ServiceID string `json:"serviceId"`
			Endpoint  string `json:"endpoint"`
			Method    string `json:"method"`
			Template  any    `json:"template"`
		}

		if err := json.Unmarshal(config, &nodeConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal API call config: %w", err)
		}

		if nodeConfig.ServiceID == "" || nodeConfig.Endpoint == "" {
			return nil, fmt.Errorf("missing serviceId or endpoint for API call in task %s", taskId)
		}

		// 2. Resolve the payload template using placeholders from the inputs
		var payload any
		if nodeConfig.Template != nil {
			payload = jsonutils.ResolveTemplate(nodeConfig.Template, func(path string) any {
				// Resolve placeholders like "trader:form.species" or "consignment.destination"
				val, exists := jsonform.GetValueByPath(inputs, path)
				if !exists {
					slog.WarnContext(ctx, "template placeholder not found in context", "path", path, "taskId", taskId)
					return nil
				}
				return val
			})
		} else {
			// Fallback: use the raw form data if no template is specified
			payload = inputs["trader:form"]
		}

		// 3. Execute the call via remote manager
		method := nodeConfig.Method
		if method == "" {
			method = "POST"
		}

		req := remote.Request{
			Method: method,
			Path:   nodeConfig.Endpoint,
			Body:   payload,
		}

		var ogaResponse map[string]any
		if err := rm.Call(ctx, nodeConfig.ServiceID, req, &ogaResponse); err != nil {
			return nil, fmt.Errorf("remote API call to %s failed: %w", nodeConfig.ServiceID, err)
		}

		// 4. Return results to be merged back into the task's JSONB data context
		return map[string]any{
			"oga_response": ogaResponse,
			"submitted_to": nodeConfig.ServiceID,
			"status":       "AWAITING_OGA_REVIEW",
		}, nil
	}
}
