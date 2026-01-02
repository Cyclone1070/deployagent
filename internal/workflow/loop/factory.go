package loop

import "github.com/Cyclone1070/iav/internal/workflow"

// LoopFactory creates new Loop instances with pre-configured dependencies.
type LoopFactory struct {
	provider      llmProvider
	tools         toolManager
	events        chan<- workflow.Event
	maxIterations int
}

// NewLoopFactory creates a new LoopFactory.
func NewLoopFactory(
	provider llmProvider,
	tools toolManager,
	events chan<- workflow.Event,
	maxIterations int,
) *LoopFactory {
	return &LoopFactory{
		provider:      provider,
		tools:         tools,
		events:        events,
		maxIterations: maxIterations,
	}
}

// Create creates a new Loop instance with the given session.
func (f *LoopFactory) Create(s session) *Loop {
	return NewLoop(f.provider, f.tools, s, f.events, f.maxIterations)
}
