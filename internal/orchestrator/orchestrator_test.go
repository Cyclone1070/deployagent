package orchestrator

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/orchestrator/adapter"
	"github.com/Cyclone1070/iav/internal/orchestrator/models"
	provider "github.com/Cyclone1070/iav/internal/provider/models"
	"github.com/Cyclone1070/iav/internal/testing/mocks"
	"github.com/Cyclone1070/iav/internal/ui"
)

// newTestOrchestrator creates an orchestrator with default config for testing
func newTestOrchestrator(p provider.Provider, pol models.PolicyService, ui ui.UserInterface, tools []adapter.Tool) *Orchestrator {
	return New(config.DefaultConfig(), p, pol, ui, tools)
}

// Test Case 1: Happy Path - Text Response
func TestRun_HappyPath_TextResponse(t *testing.T) {
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Hello, how can I help?",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		// Return error to exit loop after first response
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := New(config.DefaultConfig(), mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	err := orchestrator.Run(context.Background(), "test goal")

	// Should fail with "test complete" from ReadInput
	if err == nil || err.Error() != "failed to read user input: test complete" {
		t.Errorf("Expected 'failed to read user input: test complete', got: %v", err)
	}

	// Verify UI received the message
	if len(mockUI.Messages) != 1 || mockUI.Messages[0] != "Hello, how can I help?" {
		t.Errorf("Expected UI to receive 'Hello, how can I help?', got: %v", mockUI.Messages)
	}

	// Verify history
	if len(orchestrator.history) != 2 {
		t.Fatalf("Expected 2 messages in history, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[1].Role != "assistant" || orchestrator.history[1].Content != "Hello, how can I help?" {
		t.Errorf("Expected assistant message in history, got: %+v", orchestrator.history[1])
	}
}

// Test Case 2: Happy Path - Tool Call
func TestRun_HappyPath_ToolCall(t *testing.T) {
	toolExecuted := false

	mockTool := mocks.NewMockTool("test_tool")
	mockTool.ExecuteFunc = func(ctx context.Context, args map[string]any) (string, error) {
		toolExecuted = true
		return "tool result", nil
	}

	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		t.Logf("Generate called. Count=%d. HistoryLen=%d", callCount, len(req.History))
		if callCount == 1 {
			t.Log("Returning ToolCall")
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type: provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{
						{
							ID:   "call_1",
							Name: "test_tool",
							Args: map[string]any{"arg": "value"},
						},
					},
				},
			}, nil
		}
		t.Log("Returning Text Done")
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := New(config.DefaultConfig(), mockProvider, mockPolicy, mockUI, []adapter.Tool{mockTool})

	_ = orchestrator.Run(context.Background(), "test goal")

	if !toolExecuted {
		t.Error("Expected tool to be executed")
	}

	// Verify history has model message with tool calls and function message with results
	if len(orchestrator.history) < 3 {
		for i, msg := range orchestrator.history {
			t.Logf("Msg %d: Role=%s Content=%s ToolCalls=%v ToolResults=%v", i, msg.Role, msg.Content, msg.ToolCalls, msg.ToolResults)
		}
		t.Fatalf("Expected at least 3 messages in history, got: %d", len(orchestrator.history))
	}

	// Check model message
	if orchestrator.history[1].Role != "model" || len(orchestrator.history[1].ToolCalls) != 1 {
		t.Errorf("Expected model message with tool calls, got: %+v", orchestrator.history[1])
	}

	// Check function message
	if orchestrator.history[2].Role != "function" || len(orchestrator.history[2].ToolResults) != 1 {
		t.Errorf("Expected function message with tool results, got: %+v", orchestrator.history[2])
	}

	if orchestrator.history[2].ToolResults[0].Content != "tool result" {
		t.Errorf("Expected tool result content 'tool result', got: %s", orchestrator.history[2].ToolResults[0].Content)
	}
}

// Test Case 3: Multiple Tool Calls
func TestRun_MultipleToolCalls(t *testing.T) {
	tool1Executed := false
	tool2Executed := false

	mockTool1 := mocks.NewMockTool("tool1")
	mockTool1.ExecuteFunc = func(ctx context.Context, args map[string]any) (string, error) {
		tool1Executed = true
		return "result1", nil
	}

	mockTool2 := mocks.NewMockTool("tool2")
	mockTool2.ExecuteFunc = func(ctx context.Context, args map[string]any) (string, error) {
		tool2Executed = true
		return "result2", nil
	}

	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type: provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{
						{ID: "call_1", Name: "tool1", Args: map[string]any{}},
						{ID: "call_2", Name: "tool2", Args: map[string]any{}},
					},
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := New(config.DefaultConfig(), mockProvider, mockPolicy, mockUI, []adapter.Tool{mockTool1, mockTool2})

	_ = orchestrator.Run(context.Background(), "test goal")

	if !tool1Executed || !tool2Executed {
		t.Error("Expected both tools to be executed")
	}

	// Verify function message has 2 results
	if len(orchestrator.history) < 3 {
		t.Fatalf("Expected at least 3 messages, got: %d", len(orchestrator.history))
	}

	if len(orchestrator.history[2].ToolResults) != 2 {
		t.Errorf("Expected 2 tool results, got: %d", len(orchestrator.history[2].ToolResults))
	}
}

// Test Case 4: Refusal
func TestRun_Refusal(t *testing.T) {
	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type:          provider.ResponseTypeRefusal,
					RefusalReason: "Safety violation",
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	_ = orchestrator.Run(context.Background(), "test goal")

	// Verify system message was added
	if len(orchestrator.history) < 2 {
		t.Fatalf("Expected at least 2 messages, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[1].Role != "system" || orchestrator.history[1].Content != "Model refused: Safety violation" {
		t.Errorf("Expected system message about refusal, got: %+v", orchestrator.history[1])
	}

	// Verify UI status
	found := slices.Contains(mockUI.Statuses, "blocked: Model refused to generate")
	if !found {
		t.Error("Expected UI to show blocked status")
	}
}

// Test Case 5: Provider Error
func TestRun_ProviderError(t *testing.T) {
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		return nil, errors.New("provider error")
	}

	mockUI := mocks.NewMockUI()
	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	err := orchestrator.Run(context.Background(), "test goal")

	if err == nil || err.Error() != "provider error: provider error" {
		t.Errorf("Expected provider error, got: %v", err)
	}
}

// Test Case 6: Max Turns Reached
func TestRun_MaxTurnsReached(t *testing.T) {
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		// Always return tool call to keep looping
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeToolCall,
				ToolCalls: []models.ToolCall{
					{ID: "call_1", Name: "test_tool", Args: map[string]any{}},
				},
			},
		}, nil
	}

	mockTool := mocks.NewMockTool("test_tool")
	mockTool.ExecuteFunc = func(ctx context.Context, args map[string]any) (string, error) {
		return "result", nil
	}

	mockUI := mocks.NewMockUI()
	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{mockTool})

	err := orchestrator.Run(context.Background(), "test goal")

	if err == nil || err.Error() != "max turns (50) reached" {
		t.Errorf("Expected max turns error, got: %v", err)
	}
}

// Test Case 7: Context Cancellation
func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockProvider := mocks.NewMockProvider()
	mockUI := mocks.NewMockUI()
	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	err := orchestrator.Run(ctx, "test goal")

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// Test Case 8: Empty Tool List
func TestRun_EmptyToolList(t *testing.T) {
	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type:      provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{}, // Empty!
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	_ = orchestrator.Run(context.Background(), "test goal")

	// Verify system message was added
	if len(orchestrator.history) < 2 {
		t.Fatalf("Expected at least 2 messages, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[1].Role != "system" || orchestrator.history[1].Content != "Error: empty tool call list" {
		t.Errorf("Expected system error message, got: %+v", orchestrator.history[1])
	}
}

// Test Case 9: Unknown Tool
func TestRun_UnknownTool(t *testing.T) {
	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type: provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{
						{ID: "call_1", Name: "unknown_tool", Args: map[string]any{}},
					},
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	_ = orchestrator.Run(context.Background(), "test goal")

	// Verify function message has error result
	if len(orchestrator.history) < 3 {
		t.Fatalf("Expected at least 3 messages, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[2].ToolResults[0].Error != "unknown tool 'unknown_tool'" {
		t.Errorf("Expected unknown tool error, got: %s", orchestrator.history[2].ToolResults[0].Error)
	}
}

// Test Case 10: Policy Denial
func TestRun_PolicyDenial(t *testing.T) {
	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type: provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{
						{ID: "call_1", Name: "test_tool", Args: map[string]any{}},
					},
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockTool := mocks.NewMockTool("test_tool")

	mockPolicy := mocks.NewMockPolicy()
	mockPolicy.CheckToolFunc = func(ctx context.Context, toolName string, args map[string]any) error {
		return errors.New("policy denied")
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{mockTool})

	_ = orchestrator.Run(context.Background(), "test goal")

	// Verify function message has error result
	if len(orchestrator.history) < 3 {
		t.Fatalf("Expected at least 3 messages, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[2].ToolResults[0].Error != "policy denied: policy denied" {
		t.Errorf("Expected policy denial error, got: %s", orchestrator.history[2].ToolResults[0].Error)
	}
}

// Test Case 11: Tool Execution Error
func TestRun_ToolExecutionError(t *testing.T) {
	callCount := 0
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		callCount++
		if callCount == 1 {
			return &provider.GenerateResponse{
				Content: provider.ResponseContent{
					Type: provider.ResponseTypeToolCall,
					ToolCalls: []models.ToolCall{
						{ID: "call_1", Name: "test_tool", Args: map[string]any{}},
					},
				},
			}, nil
		}
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Done",
			},
		}, nil
	}

	mockTool := mocks.NewMockTool("test_tool")
	mockTool.ExecuteFunc = func(ctx context.Context, args map[string]any) (string, error) {
		return "", errors.New("tool execution failed")
	}

	mockPolicy := mocks.NewMockPolicy()

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("test complete")
	}

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{mockTool})

	_ = orchestrator.Run(context.Background(), "test goal")

	// Verify function message has error result
	if len(orchestrator.history) < 3 {
		t.Fatalf("Expected at least 3 messages, got: %d", len(orchestrator.history))
	}

	if orchestrator.history[2].ToolResults[0].Error != "tool execution failed" {
		t.Errorf("Expected tool execution error, got: %s", orchestrator.history[2].ToolResults[0].Error)
	}
}

// Test Case 12: User Input Error
func TestRun_UserInputError(t *testing.T) {
	mockProvider := mocks.NewMockProvider()
	mockProvider.GenerateFunc = func(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
		return &provider.GenerateResponse{
			Content: provider.ResponseContent{
				Type: provider.ResponseTypeText,
				Text: "Hello, how can I help?",
			},
		}, nil
	}

	mockUI := mocks.NewMockUI()
	mockUI.InputFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("input error")
	}

	mockPolicy := mocks.NewMockPolicy()

	orchestrator := newTestOrchestrator(mockProvider, mockPolicy, mockUI, []adapter.Tool{})

	err := orchestrator.Run(context.Background(), "test goal")

	if err == nil || err.Error() != "failed to read user input: input error" {
		t.Errorf("Expected user input error, got: %v", err)
	}
}
