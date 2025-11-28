//go:build integration

package gemini

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

func TestGeminiProvider_FreeAPI_ListModels(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping free API test")
	}

	// Create real Gemini client
	genaiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: apiKey})
	assert.NoError(t, err)
	// defer genaiClient.Close() // Client doesn't have Close method

	geminiClient := NewRealGeminiClient(genaiClient)
	provider, err := NewGeminiProvider(geminiClient, "gemini-2.0-flash-exp")
	assert.NoError(t, err)

	// List models (free API call)
	models, err := provider.ListModels(context.Background())
	assert.NoError(t, err)

	// Should return some models
	assert.NotEmpty(t, models)

	// Should include at least one gemini model
	hasGemini := false
	for _, model := range models {
		// Google API returns "models/gemini-..." format
		if strings.HasPrefix(model, "models/gemini-") || strings.HasPrefix(model, "gemini-") {
			hasGemini = true
			t.Logf("Found gemini model: %s", model)
			break
		}
	}
	assert.True(t, hasGemini, "Should have at least one gemini model. Got models: %v", models)
}
