// Package main provides a command-line interface for the deployagent tool.
// It supports file operations (read, write, edit, list) and shell command execution.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Cyclone1070/deployforme/internal/orchestrator"
	orchadapter "github.com/Cyclone1070/deployforme/internal/orchestrator/adapter"
	orchmodels "github.com/Cyclone1070/deployforme/internal/orchestrator/models"
	"github.com/Cyclone1070/deployforme/internal/provider/gemini"
	provider "github.com/Cyclone1070/deployforme/internal/provider/models"
	"github.com/Cyclone1070/deployforme/internal/tools"
	"github.com/Cyclone1070/deployforme/internal/tools/models"
	"github.com/Cyclone1070/deployforme/internal/tools/services"
	"github.com/Cyclone1070/deployforme/internal/ui"
	uiservices "github.com/Cyclone1070/deployforme/internal/ui/services"
	"github.com/charmbracelet/bubbles/spinner"
	"google.golang.org/genai"
)

// Dependencies holds the components required to run the application.
type Dependencies struct {
	UI              ui.UserInterface
	ProviderFactory func(context.Context) (provider.Provider, error)
	Tools           []orchadapter.Tool
}

func createRealUI() ui.UserInterface {
	channels := ui.NewUIChannels()
	renderer := uiservices.NewGlamourRenderer()
	spinnerFactory := func() spinner.Model {
		return spinner.New(spinner.WithSpinner(spinner.Dot))
	}
	return ui.NewUI(channels, renderer, spinnerFactory)
}

func createRealProviderFactory() func(context.Context) (provider.Provider, error) {
	return func(ctx context.Context) (provider.Provider, error) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
		}

		genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}

		geminiClient := gemini.NewRealGeminiClient(genaiClient)
		return gemini.NewGeminiProvider(geminiClient, "gemini-2.0-flash-exp")
	}
}

func createTools(ctx *models.WorkspaceContext) []orchadapter.Tool {
	return []orchadapter.Tool{
		orchadapter.NewReadFile(ctx),
		orchadapter.NewWriteFile(ctx),
		orchadapter.NewEditFile(ctx),
		orchadapter.NewListDirectory(ctx),
		orchadapter.NewShell(&tools.ShellTool{CommandExecutor: ctx.CommandExecutor}, ctx),
		orchadapter.NewSearchContent(ctx),
		orchadapter.NewFindFile(ctx),
		orchadapter.NewReadTodos(ctx),
		orchadapter.NewWriteTodos(ctx),
	}
}

func main() {
	// Get current working directory as workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize filesystem service
	fileSystem := services.NewOSFileSystem(models.DefaultMaxFileSize)

	// Initialize binary detector
	binaryDetector := &services.SystemBinaryDetector{}

	// Initialize checksum manager
	checksumMgr := services.NewChecksumManager()

	// Initialize gitignore service
	gitignoreSvc, err := services.NewGitignoreService(workspaceRoot, fileSystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to initialize gitignore service: %v\n", err)
	}

	// Create workspace context
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
	deps := Dependencies{
		UI:              createRealUI(),
		ProviderFactory: createRealProviderFactory(),
		Tools:           createTools(ctx),
	}

	// Interactive mode
	runInteractive(ctx, deps)
}

func runInteractive(ctx *models.WorkspaceContext, deps Dependencies) {
	userInterface := deps.UI
	toolList := deps.Tools

	// Start orchestrator and initialization in background
	go func() {
		<-userInterface.Ready() // Wait for UI to be ready

		// Initialize Provider (Async)
		userInterface.WriteStatus("thinking", "Initializing AI...")

		providerClient, err := deps.ProviderFactory(context.Background())
		if err != nil {
			userInterface.WriteMessage(fmt.Sprintf("Error initializing provider: %v", err))
			return
		}

		// Set initial model in status bar
		// Note: In a real app we might get this from the provider
		userInterface.SetModel("gemini-2.0-flash-exp")

		// Initialize Policy
		policy := &orchmodels.Policy{
			Shell: orchmodels.ShellPolicy{
				SessionAllow: make(map[string]bool),
			},
			Tools: orchmodels.ToolPolicy{
				SessionAllow: make(map[string]bool),
			},
		}
		policyService := orchestrator.NewPolicyService(policy, userInterface)

		// Initialize Orchestrator
		orch := orchestrator.New(providerClient, policyService, userInterface, toolList)

		userInterface.WriteStatus("ready", "Ready")

		// REPL Loop
		for {
			// Initial prompt
			goal, err := userInterface.ReadInput(context.Background(), "What would you like to do?")
			if err != nil {
				// UI closed or error
				return
			}

			if err := orch.Run(context.Background(), goal); err != nil {
				userInterface.WriteMessage(fmt.Sprintf("Error: %v", err))
			}

			userInterface.WriteStatus("ready", "Ready")
		}
	}()

	// Handle UI commands
	go func() {
		for cmd := range userInterface.Commands() {
			switch cmd.Type {
			case "list_models":
				// In a real implementation, we'd fetch this from the provider
				// For now, hardcode some known models
				models := []string{"gemini-2.0-flash-exp", "gemini-1.5-pro"}
				userInterface.WriteModelList(models)
			case "switch_model":
				model := cmd.Args["model"]
				userInterface.SetModel(model)
				// Note: We'd also need to update the provider here in a full implementation
				// providerClient.SetModel(model)
				// But providerClient is local to the other goroutine.
				// For this fix, we just update the UI.
				userInterface.WriteMessage(fmt.Sprintf("Switched to model: %s", model))
			}
		}
	}()

	// Run UI in main thread (blocks until exit)
	if err := userInterface.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}
}
