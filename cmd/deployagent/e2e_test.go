//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	orchmodels "github.com/Cyclone1070/deployforme/internal/orchestrator/models"
	providermodels "github.com/Cyclone1070/deployforme/internal/provider/models"
	"github.com/Cyclone1070/deployforme/internal/testing/testhelpers"
	"github.com/Cyclone1070/deployforme/internal/tools/models"
	"github.com/Cyclone1070/deployforme/internal/tools/services"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveMode_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create workspace context
	workspaceRoot := t.TempDir()
	fileSystem := services.NewOSFileSystem(models.DefaultMaxFileSize)
	binaryDetector := &services.SystemBinaryDetector{}
	checksumMgr := services.NewChecksumManager()
	gitignoreSvc, _ := services.NewGitignoreService(workspaceRoot, fileSystem)

	ctx := &models.WorkspaceContext{
		FS:               fileSystem,
		BinaryDetector:   binaryDetector,
		ChecksumManager:  checksumMgr,
		MaxFileSize:      models.DefaultMaxFileSize,
		WorkspaceRoot:    workspaceRoot,
		GitignoreService: gitignoreSvc,
		CommandExecutor:  &services.OSCommandExecutor{},
		DockerConfig: models.DockerConfig{
			CheckCommand: []string{"docker", "info"},
			StartCommand: []string{"docker", "desktop", "start"},
		},
	}

	// Create Mock UI
	mockUI := &testhelpers.MockUI{
		InputFunc: func(ctx context.Context, prompt string) (string, error) {
			return "List files", nil
		},
	}

	// Create mock provider that will:
	// 1. Call list_directory tool
	// 2. Return final text response
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

	providerFactory := func(ctx context.Context) (providermodels.Provider, error) {
		return mockProvider, nil
	}

	// Create dependencies
	deps := Dependencies{
		UI:              mockUI,
		ProviderFactory: providerFactory,
		Tools:           createTools(ctx),
	}

	// Run interactive mode in background
	done := make(chan bool)
	go func() {
		runInteractive(ctx, deps)
		done <- true
	}()

	// Wait for completion (should be fast with mocks)
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for runInteractive to complete")
	}

	// Verify UI received final message
	// Since orchestrator runs in background, we need to wait for it
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
