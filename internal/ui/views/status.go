package views

import (
	"fmt"
	"strings"

	"github.com/Cyclone1070/deployforme/internal/ui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatus renders the status bar
func RenderStatus(s models.State) string {
	var icon string
	var style lipgloss.Style

	switch s.StatusPhase {
	case "executing":
		icon = s.Spinner.View()
		style = StatusExecutingStyle
	case "done":
		icon = "✔"
		style = StatusDoneStyle
	case "thinking":
		icon = s.Spinner.View()
		style = StatusThinkingStyle
		// Animate the dots
		dots := strings.Repeat(".", s.DotCount)
		return style.Render(fmt.Sprintf("%s Generating%s", icon, dots))
	default:
		style = StatusDefaultStyle
	}

	if s.StatusMessage != "" {
		return style.Render(fmt.Sprintf("%s %s", icon, s.StatusMessage))
	}

	return style.Render("Ready")
}
