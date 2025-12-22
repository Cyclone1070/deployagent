package executil

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
)

func TestCollector_Write_Buffering(t *testing.T) {
	c := newCollector(1024, config.DefaultConfig().Tools.BinaryDetectionSampleSize)
	data := []byte("hello world")
	n, err := c.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}
	if got := c.String(); got != "hello world" {
		t.Errorf("String() = %q, want %q", got, "hello world")
	}
}

func TestCollector_Write_Truncation(t *testing.T) {
	c := newCollector(10, config.DefaultConfig().Tools.BinaryDetectionSampleSize)
	data := []byte("hello world") // 11 bytes
	n, err := c.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	// Write should return full length even if truncated internally to satisfy io.Writer contract
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}

	if !c.Truncated() {
		t.Error("Truncated() = false, want true")
	}

	// Should contain first 10 bytes
	if got := c.String(); got != "hello worl" {
		t.Errorf("String() = %q, want %q", got, "hello worl")
	}
}

func TestCollector_Write_Binary(t *testing.T) {
	c := newCollector(1024, config.DefaultConfig().Tools.BinaryDetectionSampleSize)
	data := []byte("hello\x00world")
	_, err := c.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Binary content should be indicated via String() output
	if got := c.String(); got != "[Binary Content]" {
		t.Errorf("String() = %q, want %q", got, "[Binary Content]")
	}

	// Should also report as truncated since we stop collecting
	if !c.Truncated() {
		t.Error("Truncated() = false, want true (binary content stops collection)")
	}
}

func TestCollector_UTF8_Boundary(t *testing.T) {
	// 3-byte character: ⌘ (E2 8C 98)
	c := newCollector(1024, config.DefaultConfig().Tools.BinaryDetectionSampleSize)

	// Write first byte
	_, _ = c.Write([]byte{0xE2})
	// Write remaining bytes
	_, _ = c.Write([]byte{0x8C, 0x98})

	if got := c.String(); got != "⌘" {
		t.Errorf("String() = %q, want %q", got, "⌘")
	}
}
