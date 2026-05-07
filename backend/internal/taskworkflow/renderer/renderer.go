package renderer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenNSW/nsw/internal/form"
	"github.com/OpenNSW/nsw/internal/taskworkflow/persistence"
	"github.com/OpenNSW/nsw/pkg/uiprojector"
)

// TaskRenderer is the service that assembles task UI using projectors.
type TaskRenderer struct {
	assembler *uiprojector.Assembler
}

func NewTaskRenderer(formService form.FormService) *TaskRenderer {
	adapter := NewTemplateAdapter(formService)
	projectors := map[string]uiprojector.Projector{
		"FORM":     uiprojector.NewFormProjector(),
		"MARKDOWN": uiprojector.NewMarkdownProjector(),
	}
	assembler := uiprojector.NewAssembler(adapter, projectors)

	return &TaskRenderer{
		assembler: assembler,
	}
}

// Render produces a legacy-compatible UI map for a given task and blueprint.
func (r *TaskRenderer) Render(ctx context.Context, task *persistence.TaskWorkflowTask, blueprint *uiprojector.Blueprint) (map[string]any, error) {
	// 1. Build Facts from the Task record
	var data map[string]any
	if err := json.Unmarshal(task.Data, &data); err != nil {
		return nil, fmt.Errorf("renderer: failed to unmarshal task data: %w", err)
	}

	// Determine Plugin-Level State from Data
	pluginState, _ := data["pluginState"].(string)

	facts := uiprojector.Facts{
		State: pluginState,
		Data:  data,
	}

	// 2. Assemble Sections
	sections, err := r.assembler.Assemble(ctx, blueprint, facts)
	if err != nil {
		return nil, fmt.Errorf("renderer: assembly failed: %w", err)
	}

	// 3. Reshape for Legacy Frontend
	return Reshape(sections), nil
}
