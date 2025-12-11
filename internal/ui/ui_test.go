package ui

import (
	"context"
	"testing"
	"time"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/testing/mocks"
	"github.com/Cyclone1070/iav/internal/ui/models"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
)

func mockSpinnerFactory() spinner.Model {
	return spinner.New()
}

func TestReadInput_ReturnsUserInput(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx := context.Background()
	expected := "hello world"
	prompt := "You: "

	go func() {
		// Verify request sent
		select {
		case req := <-channels.InputReq:
			if req.Prompt != prompt {
				t.Errorf("Expected prompt '%s', got '%s'", prompt, req.Prompt)
			}
			// Send response
			channels.InputResp <- expected
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for input request")
		}
	}()

	result, err := ui.ReadInput(ctx, prompt)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestReadInput_ContextCancelled(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := ui.ReadInput(ctx, "You: ")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, result)
}

func TestReadPermission_Allow(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)
	ctx := context.Background()
	prompt := "Allow?"
	var preview *models.ToolPreview = nil

	go func() {
		// Verify request sent
		select {
		case req := <-channels.PermReq:
			if req.Prompt != prompt {
				t.Errorf("Expected prompt '%s', got '%s'", prompt, req.Prompt)
			}
			if req.Preview != preview {
				t.Error("Expected preview to be passed")
			}
			// Send response
			channels.PermResp <- models.DecisionAllow
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for permission request")
		}
	}()

	decision, err := ui.ReadPermission(ctx, prompt, preview)
	assert.NoError(t, err)
	assert.Equal(t, models.DecisionAllow, decision)
}

func TestWriteStatus(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)

	go func() {
		// Verify status update
		select {
		case msg := <-channels.StatusChan:
			if msg.Phase != "test" {
				t.Errorf("Expected phase 'test', got '%s'", msg.Phase)
			}
			if msg.Message != "message" {
				t.Errorf("Expected message 'message', got '%s'", msg.Message)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for status update")
		}
	}()

	ui.WriteStatus("test", "message")
}

func TestWriteMessage_AddsMessage(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)

	go func() {
		msg := <-channels.MessageChan
		assert.Equal(t, "Hello", msg)
	}()

	ui.WriteMessage("Hello")
}

func TestWriteModelList_SendsList(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)
	models := []string{"a", "b"}

	go func() {
		list := <-channels.ModelListChan
		assert.Equal(t, models, list)
	}()

	ui.WriteModelList(models)
}

func TestCommands_ReturnsValidChannel(t *testing.T) {
	channels := NewUIChannels(config.DefaultConfig())
	ui := NewUI(channels, &mocks.MockMarkdownRenderer{}, mockSpinnerFactory)

	ch := ui.Commands()
	assert.NotNil(t, ch)

	// Verify we can send/receive
	go func() {
		channels.CommandChan <- models.UICommand{Type: "test"}
	}()

	select {
	case cmd := <-ch:
		assert.Equal(t, "test", cmd.Type)
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout receiving command")
	}
}
