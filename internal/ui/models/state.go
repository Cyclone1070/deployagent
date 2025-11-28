package models

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// State holds the complete UI state
type State struct {
	// Bubble Tea components
	Input    textinput.Model
	Viewport viewport.Model

	// Message state
	Messages []Message

	// Status
	StatusPhase   string
	StatusMessage string
	Spinner       spinner.Model
	DotCount      int // For animating "..." (0-3)
	CurrentModel  string

	// Permission request
	PendingPermission *PermissionRequest

	// Model selection popup
	ModelList      []string
	ModelListIndex int
	ShowModelList  bool

	// Input control
	CanSubmit bool // Enter enabled only when true

	// Dimensions
	Width  int
	Height int
}
