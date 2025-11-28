# UI Implementation Walkthrough

## Overview
This document outlines the implementation of the new Bubble Tea-based UI for the `deployagent`. The UI provides an interactive terminal experience with chat history, tool execution status, and permission requests with previews.

## Architecture
The UI follows the Model-View-Update (ELM) architecture provided by Bubble Tea, structured into:
- **Models**: `internal/ui/models` defines the state and data structures.
- **Update**: `internal/ui/update.go` handles state transitions and events.
- **Views**: `internal/ui/views` contains rendering logic for different components (Chat, Status, Popup).
- **Services**: `internal/ui/services` provides helpers for Markdown rendering and Tool Previews.

## Key Features
1.  **Interactive Chat**: Users can type goals and see the agent's responses rendered in Markdown.
2.  **Tool Status**: Real-time updates on what tool the agent is executing.
3.  **Permission System**:
    -   Intercepts sensitive actions (Shell, Edit File).
    -   Displays a preview of the action (e.g., diff of file edits, command to run).
    -   Asks for user confirmation (Allow/Deny).
4.  **Responsive Design**: The UI adapts to terminal resize events.

## Integration
The UI is integrated into `cmd/deployagent/main.go`. Running the binary without arguments starts the interactive mode.

## Testing
Comprehensive unit tests cover:
-   UI state transitions.
-   View rendering (using mocks).
-   Service logic (Markdown, Previews).
-   Concurrency safety (verified with `-race`).

## Usage
```bash
# Set API Key
export GEMINI_API_KEY=your_key_here

# Run interactive mode
go run ./cmd/deployagent
```
