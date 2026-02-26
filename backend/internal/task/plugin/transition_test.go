package plugin

import (
	"testing"
)

func TestTransitionConfig_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		config     TransitionConfig
		data       map[string]any
		wantAction string
		wantErr    bool
	}{
		{
			name: "field matches mapping entry",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE", "rejected": "REJECT"},
			},
			data:       map[string]any{"status": "approved"},
			wantAction: "APPROVE",
		},
		{
			name: "second mapping entry matched",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE", "rejected": "REJECT"},
			},
			data:       map[string]any{"status": "rejected"},
			wantAction: "REJECT",
		},
		{
			name: "field not found, default returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
				Default: "DEFAULT_ACTION",
			},
			data:       map[string]any{"other": "value"},
			wantAction: "DEFAULT_ACTION",
		},
		{
			name: "field not found, no default, error returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
			},
			data:    map[string]any{"other": "value"},
			wantErr: true,
		},
		{
			name: "field value not in mapping, default returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
				Default: "DEFAULT_ACTION",
			},
			data:       map[string]any{"status": "pending"},
			wantAction: "DEFAULT_ACTION",
		},
		{
			name: "field value not in mapping, no default, error returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
			},
			data:    map[string]any{"status": "pending"},
			wantErr: true,
		},
		{
			name: "field value is not a string, error returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
			},
			data:    map[string]any{"status": 42},
			wantErr: true,
		},
		{
			name: "field value is bool, error returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"true": "APPROVE"},
			},
			data:    map[string]any{"status": true},
			wantErr: true,
		},
		{
			name: "nested field via dot-path matched",
			config: TransitionConfig{
				Field:   "result.status",
				Mapping: map[string]string{"done": "COMPLETE"},
			},
			data:       map[string]any{"result": map[string]any{"status": "done"}},
			wantAction: "COMPLETE",
		},
		{
			name: "nested field not found, no default, error returned",
			config: TransitionConfig{
				Field:   "result.status",
				Mapping: map[string]string{"done": "COMPLETE"},
			},
			data:    map[string]any{"result": map[string]any{"other": "done"}},
			wantErr: true,
		},
		{
			name: "empty data map, no default, error returned",
			config: TransitionConfig{
				Field:   "status",
				Mapping: map[string]string{"approved": "APPROVE"},
			},
			data:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action, err := tc.config.Resolve(tc.data)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if action != tc.wantAction {
				t.Errorf("got action %q, want %q", action, tc.wantAction)
			}
		})
	}
}
