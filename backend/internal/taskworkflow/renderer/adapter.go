package renderer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OpenNSW/nsw/internal/form"
)

// templateServiceAdapter bridges the app's services to uiprojector's interface.
type templateServiceAdapter struct {
	formService form.FormService
}

func NewTemplateAdapter(formService form.FormService) *templateServiceAdapter {
	return &templateServiceAdapter{formService: formService}
}

func (a *templateServiceAdapter) GetTemplate(ctx context.Context, templateID string) ([]byte, error) {
	// Try fetching as a form first
	f, err := a.formService.GetFormByID(ctx, templateID)
	if err == nil && f != nil {
		// Wrap it in a structure that the FormProjector expects (containing "schema" and "uiSchema")
		content := map[string]any{
			"schema":   f.Schema,
			"uiSchema": f.UISchema,
		}
		return json.Marshal(content)
	}

	// Fallback or other template types can be added here.
	return nil, fmt.Errorf("renderer: template %s not found or unsupported", templateID)
}
