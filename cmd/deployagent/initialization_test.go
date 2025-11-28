//go:build integration

package main

import (
	"context"
	"runtime"
	"testing"
	"time"

	providermodels "github.com/Cyclone1070/deployforme/internal/provider/models"
	"github.com/Cyclone1070/deployforme/internal/testing/testhelpers"
	"github.com/Cyclone1070/deployforme/internal/tools/models"
	"github.com/Cyclone1070/deployforme/internal/tools/services"
	"github.com/stretchr/testify/assert"
)

func TestMain_InitTools(t *testing.T) {
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

	// Initialize tools using helper
	toolList := createTools(ctx)

	// All expected tools present
	expectedTools := []string{
		"read_file",
		"write_file",
		"edit_file",
		"list_directory",
		"run_shell",
		"search_content",
		"find_file",
		"read_todos",
		"write_todos",
	}

	for _, expected := range expectedTools {
		found := false
		for _, tool := range toolList {
			if tool.Name() == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Tool %s should be in toolList", expected)
	}

	// Correct count
	assert.Equal(t, len(expectedTools), len(toolList), "Should have exactly 9 tools")

	// All tools have valid definitions
	for _, tool := range toolList {
		def := tool.Definition()
		assert.NotEmpty(t, def.Name)
		assert.NotEmpty(t, def.Description)
	}
}

func TestMain_GoroutineCoordination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine coordination test in short mode")
	}

	initialGoroutines := runtime.NumGoroutine()

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

	// Create dependencies
	mockUI := &testhelpers.MockUI{}

	mockProvider := testhelpers.NewMockProvider()
	providerFactory := func(ctx context.Context) (providermodels.Provider, error) {
		return mockProvider, nil
	}

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

	// Wait for runInteractive to finish (MockUI.Start returns immediately)
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("runInteractive did not return")
	}

	// Check goroutine count increased (Orchestrator loop + Command loop)
	// Note: runInteractive starts 2 goroutines:
	// 1. Orchestrator loop (waits for Ready)
	// 2. Command loop (ranges over Commands)
	currentGoroutines := runtime.NumGoroutine()
	assert.Greater(t, currentGoroutines, initialGoroutines, "Should have started background goroutines")
}

func TestMain_UIStartsInstantly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UI startup timing test in short mode")
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

	// Create dependencies
	readyChan := make(chan struct{})
	mockUI := &testhelpers.MockUI{
		ReadyChan: readyChan,
	}

	// Slow provider factory
	providerInitialized := make(chan bool)
	providerFactory := func(ctx context.Context) (providermodels.Provider, error) {
		time.Sleep(100 * time.Millisecond) // Simulate slow init
		providerInitialized <- true
		return testhelpers.NewMockProvider(), nil
	}

	deps := Dependencies{
		UI:              mockUI,
		ProviderFactory: providerFactory,
		Tools:           createTools(ctx),
	}

	// Start runInteractive
	go func() {
		runInteractive(ctx, deps)
	}()

	// Verify UI is ready (we control it)
	// In real app, UI signals ready when TUI starts.
	// Here we simulate TUI starting immediately.
	close(readyChan)

	// Verify provider eventually initializes
	select {
	case <-providerInitialized:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Provider did not initialize")
	}
}
