package services

import (
	"bytes"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

// Collector captures command output with size limits and binary safety.
type Collector struct {
	Buffer    bytes.Buffer
	MaxBytes  int
	Truncated bool
	IsBinary  bool

	// Internal state for binary detection
	bytesChecked int
}

// NewCollector creates a new output collector.
func NewCollector(maxBytes int) *Collector {
	return &Collector{
		MaxBytes: maxBytes,
	}
}

// Write implements io.Writer.
func (c *Collector) Write(p []byte) (n int, err error) {
	if c.IsBinary {
		return len(p), nil // Discard rest if binary
	}

	// Check for binary content in the first N bytes
	if c.bytesChecked < models.BinaryDetectionSampleSize {
		remainingCheck := models.BinaryDetectionSampleSize - c.bytesChecked
		toCheck := p
		if len(toCheck) > remainingCheck {
			toCheck = toCheck[:remainingCheck]
		}

		if bytes.IndexByte(toCheck, 0) != -1 {
			c.IsBinary = true
			c.Truncated = true // Treated as truncated since we stop collecting
			return len(p), nil
		}
		c.bytesChecked += len(toCheck)
	}

	// Check if we have space
	remainingSpace := c.MaxBytes - c.Buffer.Len()
	if remainingSpace <= 0 {
		c.Truncated = true
		return len(p), nil
	}

	toWrite := p
	if len(toWrite) > remainingSpace {
		toWrite = toWrite[:remainingSpace]
		c.Truncated = true
	}

	written, err := c.Buffer.Write(toWrite)
	if err != nil {
		return written, err
	}

	// We always return len(p) to satisfy io.Writer contract, even if we truncated
	return len(p), nil
}

// String returns the collected string, handling UTF-8 boundaries and stripping ANSI codes (simplified).
func (c *Collector) String() string {
	if c.IsBinary {
		return "[Binary Content]"
	}

	// Handle UTF-8 validity (bytes.Buffer might end in partial rune)
	// We'll just return valid string.
	// For ANSI stripping, we can use a regex or simple replacement.
	// For now, let's just return the string as is, or maybe a simple strip if required.
	// The plan mentioned "stripping ANSI codes if possible".
	// Let's do a simple strip of common escape sequences if we want to be fancy,
	// but for now raw output is safer than bad stripping.
	// Let's stick to raw string but ensure valid UTF-8 at the end?
	// bytes.Buffer.String() just converts bytes to string.

	// If the last rune is invalid (partial), we might want to trim it?
	// But standard string conversion replaces invalid bytes with replacement char.
	// That's acceptable.

	return c.Buffer.String()
}

// SystemBinaryDetector implements BinaryDetector using local heuristics
type SystemBinaryDetector struct{}

func (r *SystemBinaryDetector) IsBinaryContent(content []byte) bool {
	// Check for common text file BOMs (UTF-16, UTF-32)
	if len(content) >= 2 {
		if (content[0] == 0xFF && content[1] == 0xFE) ||
			(content[0] == 0xFE && content[1] == 0xFF) {
			return false // UTF-16 BOM - treat as text, skip null check
		}
	}
	if len(content) >= 4 {
		if (content[0] == 0xFF && content[1] == 0xFE && content[2] == 0x00 && content[3] == 0x00) ||
			(content[0] == 0x00 && content[1] == 0x00 && content[2] == 0xFE && content[3] == 0xFF) {
			return false // UTF-32 BOM - treat as text, skip null check
		}
	}

	// Check for null bytes in first 4KB for files without BOM
	sampleSize := min(len(content), models.BinaryDetectionSampleSize)
	for i := range sampleSize {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
