package plugin

import (
	"testing"
)

func strPtr(s string) *string { return &s }

func TestEmissionConfig_Evaluate(t *testing.T) {
	tests := []struct {
		name   string
		config EmissionConfig
		data   map[string]any
		want   *string
	}{
		{
			name: "single condition matches, outcome emitted",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:manual_review_required",
					Conditions: []FieldCondition{{Field: "decision", Value: "MANUAL_REVIEW"}},
				},
			}},
			data: map[string]any{"decision": "MANUAL_REVIEW"},
			want: strPtr("npqs:phytosanitary:manual_review_required"),
		},
		{
			name: "single condition does not match, no outcome",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:manual_review_required",
					Conditions: []FieldCondition{{Field: "decision", Value: "MANUAL_REVIEW"}},
				},
			}},
			data: map[string]any{"decision": "APPROVED"},
			want: nil,
		},
		{
			name: "multiple conditions all match, outcome emitted",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome: "npqs:phytosanitary:high_risk_manual_review",
					Conditions: []FieldCondition{
						{Field: "decision", Value: "MANUAL_REVIEW"},
						{Field: "riskLevel", Value: "HIGH"},
					},
				},
			}},
			data: map[string]any{"decision": "MANUAL_REVIEW", "riskLevel": "HIGH"},
			want: strPtr("npqs:phytosanitary:high_risk_manual_review"),
		},
		{
			name: "multiple conditions, one fails, no outcome",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome: "npqs:phytosanitary:high_risk_manual_review",
					Conditions: []FieldCondition{
						{Field: "decision", Value: "MANUAL_REVIEW"},
						{Field: "riskLevel", Value: "HIGH"},
					},
				},
			}},
			data: map[string]any{"decision": "MANUAL_REVIEW", "riskLevel": "LOW"},
			want: nil,
		},
		{
			name: "first matching rule wins",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:manual_review_required",
					Conditions: []FieldCondition{{Field: "decision", Value: "MANUAL_REVIEW"}},
				},
				{
					Outcome:    "npqs:phytosanitary:approved",
					Conditions: []FieldCondition{{Field: "decision", Value: "APPROVED"}},
				},
			}},
			data: map[string]any{"decision": "MANUAL_REVIEW"},
			want: strPtr("npqs:phytosanitary:manual_review_required"),
		},
		{
			name: "second rule matches when first does not",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:manual_review_required",
					Conditions: []FieldCondition{{Field: "decision", Value: "MANUAL_REVIEW"}},
				},
				{
					Outcome:    "npqs:phytosanitary:approved",
					Conditions: []FieldCondition{{Field: "decision", Value: "APPROVED"}},
				},
			}},
			data: map[string]any{"decision": "APPROVED"},
			want: strPtr("npqs:phytosanitary:approved"),
		},
		{
			name: "nested field via dot-path matches",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:flagged",
					Conditions: []FieldCondition{{Field: "review.decision", Value: "MANUAL_REVIEW"}},
				},
			}},
			data: map[string]any{"review": map[string]any{"decision": "MANUAL_REVIEW"}},
			want: strPtr("npqs:phytosanitary:flagged"),
		},
		{
			name: "field missing from data, no outcome",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:manual_review_required",
					Conditions: []FieldCondition{{Field: "decision", Value: "MANUAL_REVIEW"}},
				},
			}},
			data: map[string]any{"status": "DONE"},
			want: nil,
		},
		{
			name: "field value is non-string, no outcome",
			config: EmissionConfig{Rules: []EmissionRule{
				{
					Outcome:    "npqs:phytosanitary:flagged",
					Conditions: []FieldCondition{{Field: "decision", Value: "1"}},
				},
			}},
			data: map[string]any{"decision": 1},
			want: nil,
		},
		{
			name:   "empty rules, no outcome",
			config: EmissionConfig{Rules: []EmissionRule{}},
			data:   map[string]any{"decision": "MANUAL_REVIEW"},
			want:   nil,
		},
		{
			name: "rule with no conditions always matches",
			config: EmissionConfig{Rules: []EmissionRule{
				{Outcome: "npqs:phytosanitary:always", Conditions: []FieldCondition{}},
			}},
			data: map[string]any{"decision": "APPROVED"},
			want: strPtr("npqs:phytosanitary:always"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.config.Evaluate(tc.data)
			if tc.want == nil && got != nil {
				t.Errorf("Evaluate() = %q, want nil", *got)
				return
			}
			if tc.want != nil && got == nil {
				t.Errorf("Evaluate() = nil, want %q", *tc.want)
				return
			}
			if tc.want != nil && *got != *tc.want {
				t.Errorf("Evaluate() = %q, want %q", *got, *tc.want)
			}
		})
	}
}
