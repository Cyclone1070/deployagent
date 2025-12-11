package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Cyclone1070/iav/internal/testing/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown_ValidMarkdown(t *testing.T) {
	mockRenderer := mocks.NewMockMarkdownRenderer()
	mockRenderer.RenderFunc = func(s string, w int) (string, error) {
		return "**RENDERED**: " + s, nil
	}

	content := "# Title\n\nParagraph"
	result, err := RenderMarkdown(content, 80, mockRenderer)

	assert.NoError(t, err)
	assert.Equal(t, "**RENDERED**: "+content, result)
}

func TestRenderMarkdown_Error(t *testing.T) {
	mockRenderer := mocks.NewMockMarkdownRenderer()
	mockRenderer.RenderFunc = func(s string, w int) (string, error) {
		return "", errors.New("render error")
	}

	content := "some content"
	result, err := RenderMarkdown(content, 80, mockRenderer)

	assert.Error(t, err)
	assert.Equal(t, content, result) // Should return original content on error
}

func TestRenderMarkdown_HugeInput(t *testing.T) {
	// Performance test with mock
	mockRenderer := mocks.NewMockMarkdownRenderer()
	mockRenderer.RenderFunc = func(s string, w int) (string, error) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		return s, nil
	}

	huge := strings.Repeat("# Title\n\n", 1000)
	start := time.Now()
	_, err := RenderMarkdown(huge, 80, mockRenderer)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 1*time.Second)
}
