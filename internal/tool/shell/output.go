package shell

import "bytes"

// Collector captures command output with size limits and binary content detection.
// It implements io.Writer and can be used to collect stdout/stderr from processes.
type Collector struct {
	buffer    bytes.Buffer
	maxBytes  int
	truncated bool
	isBinary  bool

	// Internal state for binary detection
	bytesChecked int
	sampleSize   int // Number of bytes to check for binary content
}

// NewCollector creates a new output collector with the specified maximum byte limit and binary detection sample size.
func NewCollector(maxBytes int, sampleSize int) *Collector {
	return &Collector{
		maxBytes:   maxBytes,
		sampleSize: sampleSize,
	}
}

// Write implements io.Writer for collecting process output.
// It detects binary content and enforces size limits, truncating if necessary.
func (c *Collector) Write(p []byte) (n int, err error) {
	if c.isBinary {
		return len(p), nil // Discard rest if binary
	}

	// Check for binary content in the first N bytes
	if c.bytesChecked < c.sampleSize {
		remainingCheck := c.sampleSize - c.bytesChecked
		toCheck := p
		if len(toCheck) > remainingCheck {
			toCheck = toCheck[:remainingCheck]
		}

		if bytes.IndexByte(toCheck, 0) != -1 {
			c.isBinary = true
			c.truncated = true // Treated as truncated since we stop collecting
			return len(p), nil
		}
		c.bytesChecked += len(toCheck)
	}

	// Check if we have space
	remainingSpace := c.maxBytes - c.buffer.Len()
	if remainingSpace <= 0 {
		c.truncated = true
		return len(p), nil
	}

	toWrite := p
	if len(toWrite) > remainingSpace {
		toWrite = toWrite[:remainingSpace]
		c.truncated = true
	}

	written, err := c.buffer.Write(toWrite)
	if err != nil {
		return written, err
	}

	// We always return len(p) to satisfy io.Writer contract, even if we truncated
	return len(p), nil
}

// String returns the collected output as a string.
// Returns "[Binary Content]" if binary data was detected.
func (c *Collector) String() string {
	if c.isBinary {
		return "[Binary Content]"
	}
	return c.buffer.String()
}

// Truncated returns whether the output was truncated due to size limits or binary content.
func (c *Collector) Truncated() bool {
	return c.truncated
}
