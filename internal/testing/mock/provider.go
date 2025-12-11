package mock

import (
	"context"
	"io"
	"sync"

	orchmodel "github.com/Cyclone1070/iav/internal/orchestrator/model"
	"github.com/Cyclone1070/iav/internal/provider/model"
)

// MockProvider is a controllable mock for the Provider interface.
// Supports both builder pattern (WithTextResponse) and function injection.
type MockProvider struct {
	mu sync.Mutex

	// Response queue for builder pattern
	responses     []model.GenerateResponse
	responseIndex int

	// Function injection for custom behavior
	GenerateFunc         func(ctx context.Context, req *model.GenerateRequest) (*model.GenerateResponse, error)
	GenerateStreamFunc   func(ctx context.Context, req *model.GenerateRequest) (model.ResponseStream, error)
	CountTokensFunc      func(ctx context.Context, messages []orchmodel.Message) (int, error)
	GetContextWindowFunc func() int
	SetModelFunc         func(model string) error
	GetModelFunc         func() string
	GetCapabilitiesFunc  func() model.Capabilities
	DefineToolsFunc      func(ctx context.Context, tools []model.ToolDefinition) error
	ListModelsFunc       func(ctx context.Context) ([]string, error)

	// Observable callbacks
	OnGenerateCalled func(*model.GenerateRequest)

	// Default values
	contextWindow int
	modelName     string
}

// NewMockProvider creates a mock with sensible defaults.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		responses:     make([]model.GenerateResponse, 0),
		contextWindow: 1000000, // Canonical default
		modelName:     "mock-model",
	}
}

// WithTextResponse adds a simple text response to the queue
func (m *MockProvider) WithTextResponse(text string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, model.GenerateResponse{
		Content: model.ResponseContent{
			Type: model.ResponseTypeText,
			Text: text,
		},
	})
	return m
}

// WithToolCallResponse adds a tool call response to the queue
func (m *MockProvider) WithToolCallResponse(toolCalls []orchmodel.ToolCall) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, model.GenerateResponse{
		Content: model.ResponseContent{
			Type:      model.ResponseTypeToolCall,
			ToolCalls: toolCalls,
		},
	})
	return m
}

// WithContextWindow sets the context window size
func (m *MockProvider) WithContextWindow(size int) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contextWindow = size
	return m
}

// Generate implements provider.Provider
func (m *MockProvider) Generate(ctx context.Context, req *model.GenerateRequest) (*model.GenerateResponse, error) {
	if m.OnGenerateCalled != nil {
		m.OnGenerateCalled(req)
	}

	// No lock needed for function read as it's immutable in tests
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, req)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.responseIndex < len(m.responses) {
		resp := m.responses[m.responseIndex]
		m.responseIndex++
		return &resp, nil
	}

	// Default empty response
	return &model.GenerateResponse{
		Content: model.ResponseContent{
			Type: model.ResponseTypeText,
			Text: "",
		},
	}, nil
}

// GenerateStream implements provider.Provider
func (m *MockProvider) GenerateStream(ctx context.Context, req *model.GenerateRequest) (model.ResponseStream, error) {
	if m.GenerateStreamFunc != nil {
		return m.GenerateStreamFunc(ctx, req)
	}
	return &MockResponseStream{}, nil
}

// CountTokens implements provider.Provider
func (m *MockProvider) CountTokens(ctx context.Context, history []orchmodel.Message) (int, error) {
	if m.CountTokensFunc != nil {
		return m.CountTokensFunc(ctx, history)
	}
	return len(history) * 50, nil // Canonical default
}

// GetContextWindow implements provider.Provider
func (m *MockProvider) GetContextWindow() int {
	if m.GetContextWindowFunc != nil {
		return m.GetContextWindowFunc()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// Default to 1M if not set to avoid accidental truncation in tests
	if m.contextWindow == 0 {
		return 1000000
	}
	return m.contextWindow
}

// SetModel implements provider.Provider
func (m *MockProvider) SetModel(model string) error {
	if m.SetModelFunc != nil {
		return m.SetModelFunc(model)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelName = model
	return nil
}

// GetModel implements provider.Provider
func (m *MockProvider) GetModel() string {
	if m.GetModelFunc != nil {
		return m.GetModelFunc()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modelName
}

// GetCapabilities implements provider.Provider
func (m *MockProvider) GetCapabilities() model.Capabilities {
	if m.GetCapabilitiesFunc != nil {
		return m.GetCapabilitiesFunc()
	}
	return model.Capabilities{
		SupportsToolCalling: true,
		SupportsStreaming:   true,
		SupportsJSONMode:    true,
	}
}

// DefineTools implements provider.Provider
func (m *MockProvider) DefineTools(ctx context.Context, tools []model.ToolDefinition) error {
	if m.DefineToolsFunc != nil {
		return m.DefineToolsFunc(ctx, tools)
	}
	return nil
}

// ListModels implements provider.Provider
func (m *MockProvider) ListModels(ctx context.Context) ([]string, error) {
	if m.ListModelsFunc != nil {
		return m.ListModelsFunc(ctx)
	}
	return []string{"mock-model", "mock-model-flash"}, nil
}

// MockResponseStream implements model.ResponseStream
type MockResponseStream struct {
	NextFunc  func() (*model.StreamChunk, error)
	CloseFunc func() error
}

func (m *MockResponseStream) Next() (*model.StreamChunk, error) {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return nil, io.EOF
}

func (m *MockResponseStream) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
