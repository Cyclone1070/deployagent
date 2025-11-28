package models

// Message represents a single chat message
type Message struct {
	Role    string // "user" or "assistant"
	Content string // Markdown for assistant, plain for user
}

// PermissionRequest represents a request for user permission
type PermissionRequest struct {
	Prompt  string
	Preview *ToolPreview
}

// ToolPreview contains preview data for a tool execution
type ToolPreview struct {
	Type string         // "edit_operations", "shell_command"
	Data map[string]any // Tool-specific data
}

// UICommand represents a command from the UI to the orchestrator
type UICommand struct {
	Type string // "list_models", "switch_model"
	Args map[string]string
}
