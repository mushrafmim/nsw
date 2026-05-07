package renderer

import (
	"github.com/OpenNSW/nsw/pkg/uiprojector"
)

// Reshape converts the list of rendered sections into a map that matches the
// legacy frontend's expectations (key-based access).
func Reshape(sections []uiprojector.Section) map[string]any {
	legacyContent := make(map[string]any)

	for _, s := range sections {
		// The Section.ID in the Blueprint should be set to things like
		// "traderFormInfo" or "ogaResponse" to match the portal's code.
		legacyContent[s.ID] = s.Content
	}

	return legacyContent
}
