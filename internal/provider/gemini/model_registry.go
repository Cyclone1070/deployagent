package gemini

// ModelMetadata contains information about a Gemini model
type ModelMetadata struct {
	Name              string
	InputTokenLimit   int
	OutputTokenLimit  int
	SupportsStreaming bool
	SupportsTools     bool
}

// modelRegistry is a centralized registry of known Gemini models and their capabilities
var modelRegistry = map[string]ModelMetadata{
	// Gemini 2.5 models
	"gemini-2.5-pro":          {Name: "gemini-2.5-pro", InputTokenLimit: 2_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.5-pro-latest":   {Name: "gemini-2.5-pro-latest", InputTokenLimit: 2_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.5-flash":        {Name: "gemini-2.5-flash", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.5-flash-latest": {Name: "gemini-2.5-flash-latest", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.5-flash-lite":   {Name: "gemini-2.5-flash-lite", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},

	// Gemini 2.0 models
	"gemini-2.0-flash":      {Name: "gemini-2.0-flash", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.0-flash-exp":  {Name: "gemini-2.0-flash-exp", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-2.0-flash-lite": {Name: "gemini-2.0-flash-lite", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},

	// Gemini 1.5 models
	"gemini-1.5-pro":          {Name: "gemini-1.5-pro", InputTokenLimit: 2_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-1.5-pro-latest":   {Name: "gemini-1.5-pro-latest", InputTokenLimit: 2_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-1.5-flash":        {Name: "gemini-1.5-flash", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
	"gemini-1.5-flash-latest": {Name: "gemini-1.5-flash-latest", InputTokenLimit: 1_000_000, OutputTokenLimit: 8192, SupportsStreaming: false, SupportsTools: true},
}

// GetModelMetadata returns metadata for a given model name
// Returns a default metadata if the model is not found in the registry
func GetModelMetadata(modelName string) ModelMetadata {
	if metadata, ok := modelRegistry[modelName]; ok {
		return metadata
	}

	// Default fallback for unknown models
	return ModelMetadata{
		Name:              modelName,
		InputTokenLimit:   1_000_000, // Conservative default
		OutputTokenLimit:  8192,
		SupportsStreaming: false, // Not yet implemented
		SupportsTools:     true,
	}
}
