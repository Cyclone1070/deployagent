package service

import (
	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer defines the interface for rendering markdown
type MarkdownRenderer interface {
	Render(content string, width int) (string, error)
}

// GlamourRenderer implements MarkdownRenderer using glamour
type GlamourRenderer struct {
	// We don't store a renderer anymore because we need to recreate it on width change
	// or we could cache it. For simplicity, let's recreate or cache.
	// Recreating is safer for now.
}

// NewGlamourRenderer creates a new GlamourRenderer
func NewGlamourRenderer() *GlamourRenderer {
	return &GlamourRenderer{}
}

// Render renders the content
func (g *GlamourRenderer) Render(content string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content, err
	}
	return renderer.Render(content)
}

// RenderMarkdown renders markdown content using the provided renderer
func RenderMarkdown(content string, width int, renderer MarkdownRenderer) (string, error) {
	if renderer == nil {
		renderer = NewGlamourRenderer()
	}

	rendered, err := renderer.Render(content, width)
	if err != nil {
		return content, err
	}

	return rendered, nil
}
