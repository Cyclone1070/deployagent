package view

import (
	"github.com/Cyclone1070/iav/internal/ui/model"
)

// RenderInput renders the input bar
func RenderInput(s model.State) string {
	return InputStyle.Render(s.Input.View())
}
