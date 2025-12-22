package directory

import (
	"errors"
	"fmt"
)

// -- Directory Tool Errors --

type StatError struct {
	Path  string
	Cause error
}

func (e *StatError) Error() string { return fmt.Sprintf("failed to stat %s: %v", e.Path, e.Cause) }
func (e *StatError) Unwrap() error { return e.Cause }

// ListDirError is returned when list directory fails.
type ListDirError struct {
	Path  string
	Cause error
}

func (e *ListDirError) Error() string {
	return fmt.Sprintf("failed to list directory %s: %v", e.Path, e.Cause)
}
func (e *ListDirError) Unwrap() error { return e.Cause }

// RelPathError is returned when relative path calculation fails.
type RelPathError struct {
	Path  string
	Cause error
}

func (e *RelPathError) Error() string {
	return fmt.Sprintf("failed to calculate relative path for %s: %v", e.Path, e.Cause)
}
func (e *RelPathError) Unwrap() error { return e.Cause }

// -- Sentinels --

var (
	ErrFileMissing       = errors.New("file or path does not exist")
	ErrIsDirectory       = errors.New("path is a directory")
	ErrNotADirectory     = errors.New("path is not a directory")
	ErrPathTraversal     = errors.New("path traversal detected")
	ErrPathRequired      = errors.New("path is required")
	ErrInvalidOffset     = errors.New("offset cannot be negative")
	ErrInvalidLimit      = errors.New("limit cannot be negative")
	ErrLimitExceeded     = errors.New("limit exceeds maximum")
	ErrInvalidPermission = errors.New("invalid permissions")
	ErrPatternRequired   = errors.New("pattern is required")
	ErrInvalidPattern    = errors.New("invalid glob pattern")
)
