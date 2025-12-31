package workflow

import (
	"context"

	"github.com/Cyclone1070/iav/internal/provider"
	"github.com/Cyclone1070/iav/internal/tool"
)

// Provider communicates with an LLM.
type Provider interface {
	// Generate sends messages to the LLM and returns its response.
	Generate(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error)
}

// ToolManager manages tool storage and execution.
type ToolManager interface {
	// Declarations returns all tool schemas for the LLM.
	Declarations() []tool.Declaration

	// Execute runs a tool call and returns the result as a Message.
	Execute(ctx context.Context, tc provider.ToolCall) (provider.Message, error)
}

// Tool defines the interface for individual tools.
type Tool interface {
	// Name returns the tool's identifier.
	Name() string

	// Declaration returns the tool's schema for the LLM.
	Declaration() tool.Declaration

	// Input returns a pointer to the input struct (e.g., &ReadFileInput{}).
	Input() any

	// Execute runs the tool with typed input.
	Execute(ctx context.Context, input any) (any, error)
}
