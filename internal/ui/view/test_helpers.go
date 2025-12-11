package view

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

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
