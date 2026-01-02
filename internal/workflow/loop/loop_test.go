package loop

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Cyclone1070/iav/internal/provider"
	"github.com/Cyclone1070/iav/internal/tool"
	"github.com/Cyclone1070/iav/internal/workflow"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	generateFunc func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error)
}

func (m *mockProvider) Generate(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
	return m.generateFunc(ctx, messages, tools)
}

type mockToolManager struct {
	declarations []tool.Declaration
	executeFunc  func(ctx context.Context, tc provider.ToolCall, events chan<- workflow.Event) (provider.Message, error)
}

func (m *mockToolManager) Declarations() []tool.Declaration {
	return m.declarations
}

func (m *mockToolManager) Execute(ctx context.Context, tc provider.ToolCall, events chan<- workflow.Event) (provider.Message, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, tc, events)
	}
	return provider.Message{Role: provider.RoleTool, Content: "ok"}, nil
}

type mockSession struct {
	messages []provider.Message
}

func (m *mockSession) Messages() []provider.Message {
	return m.messages
}

func (m *mockSession) Add(msg provider.Message) {
	m.messages = append(m.messages, msg)
}

func (m *mockSession) Save() error {
	return nil
}

func TestRun_SingleTurn_TextOnly(t *testing.T) {
	ctx := context.Background()
	events := make(chan workflow.Event, 10)

	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			return &provider.Message{Role: provider.RoleAssistant, Content: "Hello!"}, nil
		},
	}
	mtm := &mockToolManager{}
	ms := &mockSession{}

	l := NewLoop(mp, mtm, ms, events, 5)
	err := l.Run(ctx, "Hi")

	assert.NoError(t, err)

	assert.Equal(t, 2, len(ms.Messages()))
	assert.Equal(t, "Hi", ms.Messages()[0].Content)
	assert.Equal(t, "Hello!", ms.Messages()[1].Content)

	assert.IsType(t, workflow.ThinkingEvent{}, <-events)
	assert.Equal(t, workflow.TextEvent{Text: "Hello!"}, <-events)
	assert.IsType(t, workflow.DoneEvent{}, <-events)
}

func TestRun_SingleToolCall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	events := make(chan workflow.Event, 10)

	callCount := 0
	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			callCount++
			if callCount == 1 {
				return &provider.Message{
					Role: provider.RoleAssistant,
					ToolCalls: []provider.ToolCall{
						{Function: provider.FunctionCall{Name: "get_weather"}},
					},
				}, nil
			}
			return &provider.Message{Role: provider.RoleAssistant, Content: "It's sunny!"}, nil
		},
	}

	mtm := &mockToolManager{
		executeFunc: func(ctx context.Context, tc provider.ToolCall, events chan<- workflow.Event) (provider.Message, error) {
			return provider.Message{Role: provider.RoleTool, Content: "Sunny"}, nil
		},
	}
	ms := &mockSession{}

	l := NewLoop(mp, mtm, ms, events, 5)
	err := l.Run(ctx, "Weather?")

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Equal(t, 4, len(ms.Messages())) // User, Assist(ToolCall), ToolResp, Assist(Text)

	// Thinking
	assert.IsType(t, workflow.ThinkingEvent{}, <-events)
	// Thinking (second turn)
	assert.IsType(t, workflow.ThinkingEvent{}, <-events)
	// Text
	assert.Equal(t, workflow.TextEvent{Text: "It's sunny!"}, <-events)
	// Done
	assert.IsType(t, workflow.DoneEvent{}, <-events)
}

func TestRun_MaxIterationsExceeded_ReturnsError(t *testing.T) {
	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			return &provider.Message{
				Role: provider.RoleAssistant,
				ToolCalls: []provider.ToolCall{
					{Function: provider.FunctionCall{Name: "infinite"}},
				},
			}, nil
		},
	}
	ms := &mockSession{}
	l := NewLoop(mp, &mockToolManager{}, ms, nil, 3)
	err := l.Run(context.Background(), "go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations (3) reached")
	assert.Equal(t, "[Max iterations reached]", ms.Messages()[len(ms.Messages())-1].Content)
}

func TestRun_ProviderError_ReturnsError(t *testing.T) {
	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			return nil, fmt.Errorf("provider fail")
		},
	}
	ms := &mockSession{}
	l := NewLoop(mp, &mockToolManager{}, ms, make(chan workflow.Event, 10), 5)
	err := l.Run(context.Background(), "hi")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider.Generate")
}

func TestRun_ToolError_ReturnsError(t *testing.T) {
	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			return &provider.Message{
				Role: provider.RoleAssistant,
				ToolCalls: []provider.ToolCall{
					{Function: provider.FunctionCall{Name: "tool"}},
				},
			}, nil
		},
	}
	mtm := &mockToolManager{
		executeFunc: func(ctx context.Context, tc provider.ToolCall, events chan<- workflow.Event) (provider.Message, error) {
			return provider.Message{}, fmt.Errorf("tool fail")
		},
	}
	ms := &mockSession{}
	l := NewLoop(mp, mtm, ms, make(chan workflow.Event, 10), 5)
	err := l.Run(context.Background(), "hi")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tools.Execute")
}

func TestRun_ContextCancelled_DuringThinking_ReturnsError(t *testing.T) {
	mp := &mockProvider{
		generateFunc: func(ctx context.Context, messages []provider.Message, tools []tool.Declaration) (*provider.Message, error) {
			return &provider.Message{Role: provider.RoleAssistant, Content: "ok"}, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ms := &mockSession{}
	l := NewLoop(mp, &mockToolManager{}, ms, nil, 5)
	err := l.Run(ctx, "hi")

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, "[Session cancelled by user]", ms.Messages()[len(ms.Messages())-1].Content)
}
