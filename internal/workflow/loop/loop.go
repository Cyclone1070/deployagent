package loop

import (
	"context"
	"fmt"

	"github.com/Cyclone1070/iav/internal/provider"
	"github.com/Cyclone1070/iav/internal/workflow"
)

type Loop struct {
	provider      llmProvider
	tools         toolManager
	session       session
	events        chan<- workflow.Event
	maxIterations int
}

func NewLoop(
	provider llmProvider,
	tools toolManager,
	session session,
	events chan<- workflow.Event,
	maxIterations int,
) *Loop {
	return &Loop{
		provider:      provider,
		tools:         tools,
		session:       session,
		events:        events,
		maxIterations: maxIterations,
	}
}

func (l *Loop) Run(ctx context.Context, userInput string) error {
	l.session.Add(provider.Message{
		Role:    provider.RoleUser,
		Content: userInput,
	})

	defer func() {
		if l.events != nil {
			l.events <- workflow.DoneEvent{}
		}
	}()

	for i := 0; i < l.maxIterations; i++ {
		if err := ctx.Err(); err != nil {
			l.session.Add(provider.Message{
				Role:    provider.RoleUser,
				Content: "[Session cancelled by user]",
			})
			_ = l.session.Save() // Best effort
			return err
		}

		if l.events != nil {
			l.events <- workflow.ThinkingEvent{}
		}

		resp, err := l.provider.Generate(ctx, l.session.Messages(), l.tools.Declarations())
		if err != nil {
			_ = l.session.Save() // Best effort
			return fmt.Errorf("provider.Generate: %w", err)
		}

		l.session.Add(*resp)

		if resp.Content != "" && l.events != nil {
			l.events <- workflow.TextEvent{Text: resp.Content}
		}

		if len(resp.ToolCalls) == 0 {
			_ = l.session.Save() // Best effort
			return nil
		}

		for _, tc := range resp.ToolCalls {
			toolResp, err := l.tools.Execute(ctx, tc, l.events)
			if err != nil {
				_ = l.session.Save() // Best effort
				return fmt.Errorf("tools.Execute (%s): %w", tc.Function.Name, err)
			}
			l.session.Add(toolResp)
		}
	}

	l.session.Add(provider.Message{
		Role:    provider.RoleUser,
		Content: "[Max iterations reached]",
	})

	_ = l.session.Save() // Best effort
	return fmt.Errorf("max iterations (%d) reached", l.maxIterations)
}
