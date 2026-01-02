package toolmanager

import (
	"context"

	"github.com/Cyclone1070/iav/internal/tool"
)

// ToolResult is returned by tools after execution.
type ToolResult interface {
	// LLMContent returns the string content sent to the LLM.
	LLMContent() string

	// Display returns the display type for UI rendering.
	Display() tool.ToolDisplay

	// Success returns true if the tool execution succeeded.
	Success() bool
}

// ToolRequest is implemented by tool request structs.
type ToolRequest interface {
	// Display returns a human-readable summary for tool start events.
	Display() string
}

// Tool defines the interface for individual tools.
type Tool interface {
	// Name returns the tool's identifier.
	Name() string

	// Declaration returns the tool's schema for the LLM.
	Declaration() tool.Declaration

	// Request returns a pointer to the request struct (e.g., &ReadFileRequest{}).
	Request() ToolRequest

	// Execute runs the tool with the request and returns a ToolResult.
	Execute(ctx context.Context, req ToolRequest) (ToolResult, error)
}
