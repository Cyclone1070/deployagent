package search

import (
	"errors"
	"fmt"
)

// -- Errors --

type StatError struct {
	Path  string
	Cause error
}

func (e *StatError) Error() string { return fmt.Sprintf("failed to stat %s: %v", e.Path, e.Cause) }
func (e *StatError) Unwrap() error { return e.Cause }

// -- Sentinels --

var (
	ErrQueryRequired = errors.New("query is required")
	ErrFileMissing   = errors.New("file or path does not exist")
	ErrNotADirectory = errors.New("path is not a directory")
)
