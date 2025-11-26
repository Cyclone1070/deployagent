package gemini

import (
	"context" // Added for fmt.Errorf
	"fmt"
	"slices"
	"sync"

	"github.com/Cyclone1070/deployforme/internal/orchestrator/models"
	provider "github.com/Cyclone1070/deployforme/internal/provider/models"
)

// GeminiProvider implements the Provider interface for Google Gemini.
type GeminiProvider struct {
	client     GeminiClient
	model      string // Renamed from modelName
	mu         sync.RWMutex
	tools      []provider.ToolDefinition
	modelCache []string // Cached list of available models
}

// NewGeminiProvider creates a new Gemini provider with the given client and model
func NewGeminiProvider(client GeminiClient, model string) (*GeminiProvider, error) {
	p := &GeminiProvider{
		client: client,
		model:  model,
	}

	// Populate model cache
	ctx := context.Background()
	models, err := client.ListModels(ctx)
	if err == nil {
		p.modelCache = models
	}
	// Ignore error - validation will be skipped if cache is empty

	return p, nil
}

// ListModels returns a list of available model names
func (g *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	// Return cached list if available
	if len(g.modelCache) > 0 {
		return g.modelCache, nil
	}

	// Otherwise fetch from client
	models, err := g.client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Update cache
	g.modelCache = models
	return models, nil
}

// Generate sends a request to the Gemini API and returns the response.
func (p *GeminiProvider) Generate(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
	p.mu.RLock()
	model := p.model
	tools := p.tools
	p.mu.RUnlock()

	// Convert internal types to Gemini types
	contents := toGeminiContents(req.Prompt, req.History)
	config := toGeminiConfig(req.Config)

	// Add tools if defined
	if len(tools) > 0 {
		config.Tools = toGeminiTools(tools)
	}

	// Call Gemini API
	resp, err := p.client.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, mapGeminiError(err)
	}

	// Convert response
	return fromGeminiResponse(resp, model)
}

// GenerateStream is not yet implemented.
func (p *GeminiProvider) GenerateStream(ctx context.Context, req *provider.GenerateRequest) (provider.ResponseStream, error) {
	return nil, provider.ErrStreamingNotSupported
}

// CountTokens counts the number of tokens in the given messages.
func (p *GeminiProvider) CountTokens(ctx context.Context, messages []models.Message) (int, error) {
	p.mu.RLock()
	model := p.model
	p.mu.RUnlock()

	// Convert messages to Gemini contents
	contents := messagesToGeminiContents(messages)

	// Call Gemini API
	resp, err := p.client.CountTokens(ctx, model, contents)
	if err != nil {
		return 0, mapGeminiError(err)
	}

	return int(resp.TotalTokens), nil
}

// GetContextWindow returns the maximum context size for the current model.
func (p *GeminiProvider) GetContextWindow() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metadata := GetModelMetadata(p.model)
	return metadata.InputTokenLimit
}

// SetModel sets the model to use for generation
func (p *GeminiProvider) SetModel(model string) error {
	// Validate model if cache is available
	if len(p.modelCache) > 0 {
		valid := slices.Contains(p.modelCache, model)
		if !valid {
			return fmt.Errorf("invalid model name: %s", model)
		}
	}
	p.mu.Lock()
	p.model = model
	p.mu.Unlock()
	return nil
}

// GetModel returns the current model name
func (p *GeminiProvider) GetModel() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.model
}

// GetCapabilities returns what features the provider/model supports.
func (p *GeminiProvider) GetCapabilities() provider.Capabilities {
	p.mu.RLock()
	metadata := GetModelMetadata(p.model)
	p.mu.RUnlock()

	return provider.Capabilities{
		SupportsStreaming:   metadata.SupportsStreaming,
		SupportsToolCalling: metadata.SupportsTools,
		SupportsJSONMode:    true,
		MaxContextTokens:    metadata.InputTokenLimit,
		MaxOutputTokens:     metadata.OutputTokenLimit,
	}
}

// DefineTools registers tool definitions with the provider for native tool calling.
func (p *GeminiProvider) DefineTools(ctx context.Context, tools []provider.ToolDefinition) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tools = tools
	return nil
}
