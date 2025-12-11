package view

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/testing/mock"
	"github.com/Cyclone1070/iav/internal/ui/model"
	"github.com/stretchr/testify/assert"
)

func TestRenderChat_NoMessages(t *testing.T) {
	state := model.State{Messages: []model.Message{}}
	result := RenderChat(state, mock.NewMockMarkdownRenderer())
	assert.Contains(t, result, "No messages yet")
}

func TestRenderChat_WithMessages(t *testing.T) {
	// Since RenderChat just delegates to Viewport.View(), we test that it returns the viewport content
	vp := createTestViewport()
	vp.SetContent("Rendered Content")

	state := model.State{
		Messages: []model.Message{{Role: "user", Content: "Hello"}},
		Viewport: vp,
	}

	result := RenderChat(state, mock.NewMockMarkdownRenderer())
	assert.Contains(t, result, "Rendered Content")
}
