package view

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/ui/model"
	"github.com/stretchr/testify/assert"
)

func TestRenderStatus_Executing(t *testing.T) {
	state := model.State{
		StatusPhase:   "executing",
		StatusMessage: "EditFile main.go",
		Spinner:       createTestSpinner(),
	}

	result := RenderStatus(state)

	// Spinner view might change based on time, but it should contain the message
	assert.Contains(t, result, "EditFile main.go")
	// Check for blue color (lipgloss uses ANSI codes)
	// We can't easily check exact ANSI codes without being fragile,
	// but we can check it's not empty.
	assert.NotEmpty(t, result)
}

func TestRenderStatus_Done(t *testing.T) {
	state := model.State{
		StatusPhase:   "done",
		StatusMessage: "EditFile main.go",
	}

	result := RenderStatus(state)

	assert.Contains(t, result, "✔")
	assert.Contains(t, result, "EditFile main.go")
}

func TestRenderStatus_Thinking(t *testing.T) {
	state := model.State{
		StatusPhase: "thinking",
		DotCount:    2,
		Spinner:     createTestSpinner(),
	}

	result := RenderStatus(state)

	assert.Contains(t, result, "Generating..") // 2 dots
}

func TestRenderStatus_DefaultReady(t *testing.T) {
	state := model.State{
		StatusPhase:   "",
		StatusMessage: "",
	}

	result := RenderStatus(state)

	assert.Contains(t, result, "Ready")
}
