package views

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// MockMarkdownRenderer for testing
type MockMarkdownRenderer struct {
	RenderFunc func(string, int) (string, error)
}

func (m *MockMarkdownRenderer) Render(content string, width int) (string, error) {
	if m.RenderFunc != nil {
		return m.RenderFunc(content, width)
	}
	return content, nil
}

// MockSpinner for testing
type MockSpinner struct {
	ViewResult string
}

// We can't easily mock bubbles components because they are structs.
// But we can initialize them.

func createTestTextInput(value string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(value)
	return ti
}

func createTestViewport() viewport.Model {
	return viewport.New(80, 20)
}

func createTestSpinner() spinner.Model {
	return spinner.New()
}
