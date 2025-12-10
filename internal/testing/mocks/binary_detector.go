package mocks

// MockBinaryDetector implements models.BinaryDetector with configurable behaviour
type MockBinaryDetector struct {
	BinaryPaths   map[string]bool
	BinaryContent map[string]bool // content hash -> is binary
	SampleSize    int             // Number of bytes to sample for binary detection
}

// NewMockBinaryDetector creates a new mock binary detector
func NewMockBinaryDetector() *MockBinaryDetector {
	return &MockBinaryDetector{
		BinaryPaths:   make(map[string]bool),
		BinaryContent: make(map[string]bool),
		SampleSize:    4096, // Default matches config
	}
}

// SetBinaryPath marks a path as binary (deprecated, kept for test compatibility)
// Note: Binary detection now uses IsBinaryContent which checks actual file content
func (f *MockBinaryDetector) SetBinaryPath(path string, isBinary bool) {
	f.BinaryPaths[path] = isBinary
}

func (f *MockBinaryDetector) IsBinaryContent(content []byte) bool {
	// Check for common text file BOMs (UTF-16, UTF-32)
	if len(content) >= 2 {
		if (content[0] == 0xFF && content[1] == 0xFE) ||
			(content[0] == 0xFE && content[1] == 0xFF) {
			return false
		}
	}
	if len(content) >= 4 {
		if (content[0] == 0xFF && content[1] == 0xFE && content[2] == 0x00 && content[3] == 0x00) ||
			(content[0] == 0x00 && content[1] == 0x00 && content[2] == 0xFE && content[3] == 0xFF) {
			return false
		}
	}

	sampleSize := len(content)
	if f.SampleSize < sampleSize {
		sampleSize = f.SampleSize
	}

	for i := 0; i < sampleSize; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}
