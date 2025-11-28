package ui

import (
	"context"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
)

// Mock dependencies
type MockMarkdownRenderer struct {
	RenderFunc func(string, int) (string, error)
}

func (m *MockMarkdownRenderer) Render(content string, width int) (string, error) {
	if m.RenderFunc != nil {
		return m.RenderFunc(content, width)
	}
	return content, nil
}

func mockSpinnerFactory() spinner.Model {
	return spinner.New()
}

func TestReadInput_ReturnsUserInput(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx := context.Background()
	expected := "hello world"

	go func() {
		select {
		case req := <-channels.InputReq:
			assert.Equal(t, "You: ", req.prompt)
			channels.InputResp <- expected
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for input request")
		}
	}()

	result, err := ui.ReadInput(ctx, "You: ")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestReadInput_ContextCancelled(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ui.ReadInput(ctx, "You: ")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, result)
}

func TestReadPermission_Allow(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx := context.Background()

	go func() {
		select {
		case req := <-channels.PermReq:
			assert.Equal(t, "Allow?", req.prompt)
			channels.PermResp <- DecisionAllow
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for perm request")
		}
	}()

	decision, err := ui.ReadPermission(ctx, "Allow?", nil)
	assert.NoError(t, err)
	assert.Equal(t, DecisionAllow, decision)
}

func TestWriteStatus_UpdatesStatus(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)

	go func() {
		msg := <-channels.StatusChan
		assert.Equal(t, "executing", msg.phase)
		assert.Equal(t, "EditFile main.go", msg.message)
	}()

	ui.WriteStatus("executing", "EditFile main.go")
	// Give time for channel send (it's buffered but we want to ensure no panic)
}

func TestWriteMessage_AddsMessage(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)

	go func() {
		msg := <-channels.MessageChan
		assert.Equal(t, "Hello", msg)
	}()

	ui.WriteMessage("Hello")
}

func TestWriteModelList_SendsList(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)
	models := []string{"a", "b"}

	go func() {
		list := <-channels.ModelListChan
		assert.Equal(t, models, list)
	}()

	ui.WriteModelList(models)
}

func TestCommands_ReturnsValidChannel(t *testing.T) {
	channels := NewUIChannels()
	ui := NewUI(channels, &MockMarkdownRenderer{}, mockSpinnerFactory)

	ch := ui.Commands()
	assert.NotNil(t, ch)

	// Verify we can send/receive
	go func() {
		channels.CommandChan <- UICommand{Type: "test"}
	}()

	select {
	case cmd := <-ch:
		assert.Equal(t, "test", cmd.Type)
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout receiving command")
	}
}
