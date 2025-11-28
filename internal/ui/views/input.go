package views

import (
	"github.com/Cyclone1070/deployforme/internal/ui/models"
)

// RenderInput renders the input bar
func RenderInput(s models.State) string {
	return InputStyle.Render(s.Input.View())
}
