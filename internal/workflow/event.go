package workflow

// EventType identifies the kind of event.
type EventType int

const (
	EventTextChunk EventType = iota
	EventToolStart
	EventToolEnd
	EventError
	EventDone
)

// Event is a real-time notification for UI.
type Event struct {
	Type     EventType
	Text     string // for EventTextChunk
	ToolName string // for EventToolStart/EventToolEnd
	ToolArgs string // for EventToolStart
	Error    error  // for EventError
}
