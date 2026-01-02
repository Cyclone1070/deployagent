package session

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestSession_Add(t *testing.T) {
	s := &Session{
		id:       "test-id",
		messages: []provider.Message{},
	}

	s.Add(provider.Message{Role: provider.RoleUser, Content: "Hello"})

	assert.Equal(t, 1, len(s.Messages()))
	assert.Equal(t, "Hello", s.Messages()[0].Content)
}

func TestSession_Messages(t *testing.T) {
	s := &Session{
		id: "test-id",
		messages: []provider.Message{
			{Role: provider.RoleUser, Content: "First"},
			{Role: provider.RoleAssistant, Content: "Second"},
		},
	}

	msgs := s.Messages()

	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, "First", msgs[0].Content)
	assert.Equal(t, "Second", msgs[1].Content)
}

func TestSession_ID(t *testing.T) {
	s := &Session{id: "my-session-id"}

	assert.Equal(t, "my-session-id", s.ID())
}

func TestSession_Clear(t *testing.T) {
	s := &Session{
		id: "test-id",
		messages: []provider.Message{
			{Role: provider.RoleUser, Content: "Hello"},
		},
	}

	s.Clear()

	assert.Empty(t, s.Messages())
}
