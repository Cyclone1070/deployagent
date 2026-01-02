package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cyclone1070/iav/internal/provider"
)

// sessionDTO is used for JSON serialization.
type sessionDTO struct {
	ID       string             `json:"id"`
	Messages []provider.Message `json:"messages"`
}

// Session represents a conversation session with message history.
type Session struct {
	id         string
	messages   []provider.Message
	storageDir string
}

// ID returns the session identifier.
func (s *Session) ID() string {
	return s.id
}

// Messages returns the slice of messages in the session.
func (s *Session) Messages() []provider.Message {
	return s.messages
}

// Add appends a message to the session.
func (s *Session) Add(msg provider.Message) {
	s.messages = append(s.messages, msg)
}

// Save persists the session to disk.
func (s *Session) Save() error {
	path := filepath.Join(s.storageDir, s.id+".json")
	dto := sessionDTO{
		ID:       s.id,
		Messages: s.messages,
	}
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Delete removes the session from disk.
func (s *Session) Delete() error {
	path := filepath.Join(s.storageDir, s.id+".json")
	return os.Remove(path)
}

// Clear removes all messages from the session but keeps the ID.
func (s *Session) Clear() {
	s.messages = []provider.Message{}
}
