package mock

// MockMarkdownRenderer implements ui.MarkdownRenderer for testing.
type MockMarkdownRenderer struct {
	RenderFunc func(content string, width int) (string, error)
}

func NewMockMarkdownRenderer() *MockMarkdownRenderer {
	return &MockMarkdownRenderer{}
}

func (m *MockMarkdownRenderer) Render(content string, width int) (string, error) {
	if m.RenderFunc != nil {
		return m.RenderFunc(content, width)
	}
	return content, nil // Pass-through default
}
