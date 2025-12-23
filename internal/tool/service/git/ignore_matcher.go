package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// GitignoreReadError is returned when .gitignore cannot be read.
type GitignoreReadError struct {
	Path  string
	Cause error
}

func (e *GitignoreReadError) Error() string {
	return fmt.Sprintf("failed to read .gitignore at %s: %v", e.Path, e.Cause)
}
func (e *GitignoreReadError) Unwrap() error { return e.Cause }

// fileSystem defines the minimal filesystem interface needed for gitignore service.
type fileSystem interface {
	Stat(path string) (os.FileInfo, error)
	ReadFileRange(path string, offset, limit int64) ([]byte, error)
}

// IgnoreMatcher implements gitignore pattern matching using go-git's gitignore matcher.
type IgnoreMatcher struct {
	matcher gitignore.Matcher
}

// NewIgnoreMatcher creates a new gitignore matcher by loading .gitignore from workspace root.
// Returns a matcher that never ignores if .gitignore doesn't exist (no error).
func NewIgnoreMatcher(workspaceRoot string, fs fileSystem) (*IgnoreMatcher, error) {
	if workspaceRoot == "" {
		panic("workspaceRoot is required")
	}
	if fs == nil {
		panic("fs is required")
	}
	gitignorePath := filepath.Join(workspaceRoot, ".gitignore")

	// Check if .gitignore exists
	_, err := fs.Stat(gitignorePath)
	if err != nil {
		// .gitignore doesn't exist - return a matcher that never ignores
		return &IgnoreMatcher{matcher: nil}, nil
	}

	// Read .gitignore file
	content, err := fs.ReadFileRange(gitignorePath, 0, 0)
	if err != nil {
		return nil, &GitignoreReadError{Path: gitignorePath, Cause: err}
	}

	// Parse gitignore patterns line by line
	var patterns []gitignore.Pattern
	lines := splitLines(string(content))
	for _, line := range lines {
		if line == "" {
			continue // Skip blank lines
		}
		pattern := gitignore.ParsePattern(line, nil)
		if pattern != nil {
			patterns = append(patterns, pattern)
		}
	}
	matcher := gitignore.NewMatcher(patterns)

	return &IgnoreMatcher{matcher: matcher}, nil
}

// ShouldIgnore checks if a relative path matches any gitignore patterns.
// Returns false if no .gitignore was loaded.
func (m *IgnoreMatcher) ShouldIgnore(relativePath string) bool {
	if m.matcher == nil {
		return false
	}

	// Convert to gitignore format (split path into segments)
	segments := splitPath(relativePath)
	return m.matcher.Match(segments, false)
}

// splitPath splits a path into segments for gitignore matching.
// It normalizes path separators and filters out empty and "." segments.
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}

	// Normalize path separators
	normalized := filepath.ToSlash(path)

	// Split by forward slash
	parts := strings.Split(normalized, "/")
	var segments []string
	for _, part := range parts {
		if part != "" && part != "." {
			segments = append(segments, part)
		}
	}

	return segments
}

// splitLines splits content into lines, handling both \n and \r\n line endings.
func splitLines(content string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		} else if content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 2
			i++ // Skip the \n
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

// NoOpMatcher is a gitignore matcher that never ignores any files.
// It is used when gitignore functionality is disabled or fails to initialize.
type NoOpMatcher struct{}

// ShouldIgnore always returns false for NoOpMatcher.
func (m *NoOpMatcher) ShouldIgnore(relativePath string) bool {
	return false
}
