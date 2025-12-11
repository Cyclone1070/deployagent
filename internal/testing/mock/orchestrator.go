package mock

import (
	"context"

	provider "github.com/Cyclone1070/iav/internal/provider/model"
)

// MockPolicy implements model.PolicyService for testing.
type MockPolicy struct {
	CheckToolFunc  func(ctx context.Context, toolName string, args map[string]any) error
	CheckShellFunc func(ctx context.Context, command []string) error
}

func NewMockPolicy() *MockPolicy {
	return &MockPolicy{}
}

func (m *MockPolicy) CheckTool(ctx context.Context, toolName string, args map[string]any) error {
	if m.CheckToolFunc != nil {
		return m.CheckToolFunc(ctx, toolName, args)
	}
	return nil // Allow by default
}

func (m *MockPolicy) CheckShell(ctx context.Context, command []string) error {
	if m.CheckShellFunc != nil {
		return m.CheckShellFunc(ctx, command)
	}
	return nil // Allow by default
}

// MockTool implements adapter.Tool for testing.
type MockTool struct {
	NameFunc        func() string
	DescriptionFunc func() string
	DefinitionFunc  func() provider.ToolDefinition
	ExecuteFunc     func(ctx context.Context, args map[string]any) (string, error)
}

func NewMockTool(name string) *MockTool {
	return &MockTool{
		NameFunc: func() string { return name },
	}
}

func (m *MockTool) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock_tool"
}

func (m *MockTool) Description() string {
	if m.DescriptionFunc != nil {
		return m.DescriptionFunc()
	}
	return "Mock Tool Description"
}

func (m *MockTool) Definition() provider.ToolDefinition {
	if m.DefinitionFunc != nil {
		return m.DefinitionFunc()
	}
	return provider.ToolDefinition{
		Name:        m.Name(),
		Description: m.Description(),
	}
}

func (m *MockTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, args)
	}
	return "mock response", nil
}
