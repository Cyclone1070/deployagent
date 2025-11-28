//go:build integration

package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	orchadapter "github.com/Cyclone1070/deployforme/internal/orchestrator/adapter"
	orchmodels "github.com/Cyclone1070/deployforme/internal/orchestrator/models"
	"github.com/Cyclone1070/deployforme/internal/testing/testhelpers"
	"github.com/Cyclone1070/deployforme/internal/tools/models"
	"github.com/Cyclone1070/deployforme/internal/tools/services"
	"github.com/Cyclone1070/deployforme/internal/ui"
	uiservices "github.com/Cyclone1070/deployforme/internal/ui/services"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
)

func TestOrchestratorProvider_ToolCallResponse(t *testing.T) {
	t.Parallel()

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

	// Create UI
	channels := ui.NewUIChannels()
	renderer := uiservices.NewGlamourRenderer()
	spinnerFactory := func() spinner.Model {
		return spinner.New(spinner.WithSpinner(spinner.Dot))
	}
	userInterface := ui.NewUI(channels, renderer, spinnerFactory)

	// Create cancellable context
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Service UI input requests in background
	go func() {
		count := 0
		for {
			select {
			case req := <-channels.InputReq:
				count++
				if count > 1 {
					// Cancel context to stop the loop
					cancel()
					// Try to send exit, but don't block if context is done
					select {
					case channels.InputResp <- "exit":
					case <-runCtx.Done():
					}
				} else {
					// Send next input for continuation
					channels.InputResp <- "continue"
				}
				_ = req
			case <-runCtx.Done():
				return
			}
		}
	}()

	// Track UI messages
	messageDone := make(chan []string)
	go func() {
		var msgs []string
		for msg := range channels.MessageChan {
			msgs = append(msgs, msg)
		}
		messageDone <- msgs
	}()

	// Initialize tools
	toolList := []orchadapter.Tool{
		orchadapter.NewListDirectory(ctx),
	}

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
		WithTextResponse("Found 0 files")

	// Create policy
	policy := &orchmodels.Policy{
		Shell: orchmodels.ShellPolicy{
			SessionAllow: make(map[string]bool),
		},
		Tools: orchmodels.ToolPolicy{
			Allow:        []string{"list_directory"},
			SessionAllow: make(map[string]bool),
		},
	}
	policyService := NewPolicyService(policy, userInterface)

	// Create orchestrator
	orch := New(mockProvider, policyService, userInterface, toolList)

	// Run orchestrator
	err := orch.Run(runCtx, "List files")

	// Close channel to signal completion and get collected messages
	close(channels.MessageChan)
	messages := <-messageDone

	// Should complete with cancellation error or nil
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("Run returned unexpected error: %v", err)
	}

	// History progression - this will need to be verified once we have access to history
	// For now, we verify it doesn't crash and completes
	assert.NotNil(t, orch)

	// Check if expected message was delivered
	found := false
	for _, msg := range messages {
		if strings.Contains(msg, "Found") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected message containing 'Found' not received. Messages: %v", messages)
}

func TestOrchestratorProvider_ContextTruncation(t *testing.T) {
	t.Parallel()

	// Create small context window provider
	mockProvider := testhelpers.NewMockProvider().
		WithContextWindow(200) // Very small window

	// Create workspace context
	workspaceRoot := t.TempDir()
	fileSystem := services.NewOSFileSystem(models.DefaultMaxFileSize)
	ctx := &models.WorkspaceContext{
		FS:              fileSystem,
		BinaryDetector:  &services.SystemBinaryDetector{},
		ChecksumManager: services.NewChecksumManager(),
		MaxFileSize:     models.DefaultMaxFileSize,
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: &services.OSCommandExecutor{},
	}

	// Create UI
	channels := ui.NewUIChannels()
	renderer := uiservices.NewGlamourRenderer()
	userInterface := ui.NewUI(channels, renderer, func() spinner.Model {
		return spinner.New(spinner.WithSpinner(spinner.Dot))
	})

	// Service UI
	go func() {
		for range channels.InputReq {
			channels.InputResp <- "done"
		}
	}()
	go func() {
		for range channels.MessageChan {
		}
	}()

	// Create policy
	policy := &orchmodels.Policy{
		Shell: orchmodels.ShellPolicy{SessionAllow: make(map[string]bool)},
		Tools: orchmodels.ToolPolicy{SessionAllow: make(map[string]bool)},
	}
	policyService := NewPolicyService(policy, userInterface)

	// Create orchestrator with tools
	toolList := []orchadapter.Tool{
		orchadapter.NewListDirectory(ctx),
	}

	orch := New(mockProvider, policyService, userInterface, toolList)

	// Build large history by adding many messages
	for i := 0; i < 10; i++ {
		orch.history = append(orch.history, orchmodels.Message{
			Role:    "user",
			Content: "This is a test message with some content to make it longer and consume more tokens",
		})
	}

	// Initial token count should be high
	initialTokens, err := mockProvider.CountTokens(context.Background(), orch.history)
	assert.NoError(t, err)
	assert.Greater(t, initialTokens, 200)

	// Truncation should succeed
	err = orch.checkAndTruncateHistory(context.Background())
	assert.NoError(t, err)

	// Token count after truncation should be within limit
	finalTokens, err := mockProvider.CountTokens(context.Background(), orch.history)
	assert.NoError(t, err)
	assert.LessOrEqual(t, finalTokens, 200)

	//First message should be preserved (if we added one as the goal)
	// This test might need adjustment based on actual truncation logic
}
