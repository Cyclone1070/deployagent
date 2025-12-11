package mocks

import (
	"context"
	"sync"

	uimodels "github.com/Cyclone1070/iav/internal/ui/models"
)

// MockUI implements ui.UserInterface with thread-safe operations.
type MockUI struct {
	mu sync.Mutex

	Messages  []string
	Statuses  []string
	ModelList []string

	// Function injection
	InputFunc          func(ctx context.Context, prompt string) (string, error)
	ReadPermissionFunc func(ctx context.Context, prompt string, preview *uimodels.ToolPreview) (uimodels.PermissionDecision, error)

	// Observable callbacks
	OnReadyCalled      func()
	OnModelListWritten func(models []string)

	// Channels for test control
	ReadyChan    chan struct{}
	StartBlocker chan struct{}
	CommandsChan chan uimodels.UICommand
}

func NewMockUI() *MockUI {
	return &MockUI{
		Messages:     make([]string, 0),
		Statuses:     make([]string, 0),
		CommandsChan: make(chan uimodels.UICommand, 1), // Buffered to prevent blocking in simple tests
	}
}

// WriteMessage adds a message to the history
func (m *MockUI) WriteMessage(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, message)
}

// GetMessages returns a copy of the messages
func (m *MockUI) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgs := make([]string, len(m.Messages))
	copy(msgs, m.Messages)
	return msgs
}

// WriteStatus adds a status update
func (m *MockUI) WriteStatus(phase string, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Store in format "phase: message" for simple testing
	m.Statuses = append(m.Statuses, phase+": "+message)
}

// ReadInput delegates to InputFunc or returns default "yes"
func (m *MockUI) ReadInput(ctx context.Context, prompt string) (string, error) {
	if m.InputFunc != nil {
		return m.InputFunc(ctx, prompt)
	}
	return "yes", nil
}

// ReadPermission delegates to ReadPermissionFunc or returns Allowed
func (m *MockUI) ReadPermission(ctx context.Context, prompt string, preview *uimodels.ToolPreview) (uimodels.PermissionDecision, error) {
	if m.ReadPermissionFunc != nil {
		return m.ReadPermissionFunc(ctx, prompt, preview)
	}
	return uimodels.DecisionAllow, nil
}

// Ready returns a channel that closes when UI is ready
func (m *MockUI) Ready() <-chan struct{} {
	if m.OnReadyCalled != nil {
		m.OnReadyCalled()
	}

	if m.ReadyChan != nil {
		return m.ReadyChan
	}
	// Default: always ready
	ch := make(chan struct{})
	close(ch)
	return ch
}

// WriteModelList updates the model list
func (m *MockUI) WriteModelList(models []string) {
	m.mu.Lock()
	m.ModelList = models
	cb := m.OnModelListWritten
	m.mu.Unlock()

	if cb != nil {
		cb(models)
	}
}

// Start blocks if StartBlocker is set
func (m *MockUI) Start() error {
	if m.StartBlocker != nil {
		<-m.StartBlocker
	}
	return nil
}

// Commands returns the command channel
func (m *MockUI) Commands() <-chan uimodels.UICommand {
	return m.CommandsChan
}

// SetModel updates current model (no-op in mock unless we track it)
func (m *MockUI) SetModel(model string) {
	// We could track this if needed
}
