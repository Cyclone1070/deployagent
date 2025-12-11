package ui

import (
	"context"

	"github.com/Cyclone1070/iav/internal/ui/model"
)

// UserInterface defines the contract for all user interactions.
// It follows a Read/Write pattern for clarity.
//
// Context Usage:
// All methods accept context.Context for cancellation support.
// If the user cancels (Ctrl+C), the context will be cancelled,
// and implementations should return immediately with context.Canceled error.
type UserInterface interface {
	// ReadInput prompts the user for general text input
	ReadInput(ctx context.Context, prompt string) (string, error)

	// ReadPermission prompts the user for a yes/no/always permission decision
	ReadPermission(ctx context.Context, prompt string, preview *model.ToolPreview) (model.PermissionDecision, error)

	// WriteStatus displays ephemeral status updates (e.g., "Thinking...")
	WriteStatus(phase string, message string)

	// WriteMessage displays the agent's actual text responses
	WriteMessage(content string)

	// WriteModelList sends a list of available models to the UI
	WriteModelList(models []string)

	// SetModel updates the current model name displayed in the UI
	SetModel(model string)

	// Commands returns a channel for UI-initiated commands (e.g., /models)
	Commands() <-chan model.UICommand

	// Start starts the UI loop (blocking)
	Start() error

	// Ready returns a channel that is closed when the UI is ready
	Ready() <-chan struct{}
}
