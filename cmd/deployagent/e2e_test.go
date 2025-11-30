//go:build integration

package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	orchmodels "github.com/Cyclone1070/deployforme/internal/orchestrator/models"
	providermodels "github.com/Cyclone1070/deployforme/internal/provider/models"
	"github.com/Cyclone1070/deployforme/internal/testing/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveMode_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Workspace context creation removed as it is now internal to runInteractive

	// Control when MockUI exits
	startBlocker := make(chan struct{})

	// Create Mock UI
	var inputCount int
	mockUI := &testhelpers.MockUI{
		InputFunc: func(ctx context.Context, prompt string) (string, error) {
			inputCount++
			if inputCount > 1 {
				return "", fmt.Errorf("stop test")
			}
			return "List files", nil
		},
		StartBlocker: startBlocker,
	}

	// Track what orchestrator sends to provider
	var allProviderCalls []providermodels.GenerateRequest
	var mu sync.Mutex

	// Create mock provider
	mockProvider := testhelpers.NewMockProvider().
		WithToolCallResponse([]orchmodels.ToolCall{
			{
				ID:   "call_1",
				Name: "list_directory",
				Args: map[string]any{
					"path":      ".",
					"max_depth": -1,
					"offset":    0,
					"limit":     100,
				},
			},
		}).
		WithTextResponse("Found files in current directory")

	// Capture provider inputs
	mockProvider.OnGenerateCalled = func(req *providermodels.GenerateRequest) {
		mu.Lock()
		defer mu.Unlock()
		allProviderCalls = append(allProviderCalls, *req)
	}

	providerFactory := func(ctx context.Context) (providermodels.Provider, error) {
		return mockProvider, nil
	}

	// Create dependencies
	deps := Dependencies{
		UI:              mockUI,
		ProviderFactory: providerFactory,
		Tools:           nil, // Created in goroutine
	}

	// Run interactive mode in background
	go func() {
		runInteractive(context.Background(), deps)
	}()

	// Give orchestrator time to initialize and run
	time.Sleep(300 * time.Millisecond)

	// Let UI exit
	close(startBlocker)

	// Small delay for cleanup
	time.Sleep(50 * time.Millisecond)

	// Verify provider called multiple times (tool call + final response)
	mu.Lock()
	callCount := len(allProviderCalls)
	mu.Unlock()
	assert.GreaterOrEqual(t, callCount, 2,
		"Provider should be called at least twice (initial + after tool execution)")

	// Verify orchestrator sent tool results back to provider
	mu.Lock()
	lastHistory := allProviderCalls[len(allProviderCalls)-1].History
	mu.Unlock()

	foundToolResult := false
	for _, msg := range lastHistory {
		if msg.Role == "function" && len(msg.ToolResults) > 0 {
			foundToolResult = true
			// Verify tool result structure
			assert.Equal(t, "list_directory", msg.ToolResults[0].Name)
			assert.Equal(t, "call_1", msg.ToolResults[0].ID)
			break
		}
	}
	assert.True(t, foundToolResult,
		"Orchestrator should send tool results to provider in history")

	// Verify UI received final message
	foundResponse := false
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-timeout:
			break loop
		case <-ticker.C:
			// Check messages
			for _, msg := range mockUI.GetMessages() {
				if msg == "Found files in current directory" {
					foundResponse = true
					break loop
				}
			}
		}
	}
	assert.True(t, foundResponse, "Should have received final response. Messages: %v", mockUI.GetMessages())
}
