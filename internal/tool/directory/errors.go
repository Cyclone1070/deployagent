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

// -- Sentinels --

var (
	ErrFileMissing     = errors.New("file or path does not exist")
	ErrNotADirectory   = errors.New("path is not a directory")
	ErrPathRequired    = errors.New("path is required")
	ErrPatternRequired = errors.New("pattern is required")
	ErrInvalidPattern  = errors.New("invalid glob pattern")
)
