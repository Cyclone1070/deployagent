package pathutil

import (
	"errors"
	"fmt"
)

// -- Error Types --

// WorkspaceRootError is returned when the workspace root is invalid.
type WorkspaceRootError struct {
	Root  string
	Cause error
}

func (e *WorkspaceRootError) Error() string {
	return fmt.Sprintf("invalid workspace root %s: %v", e.Root, e.Cause)
}
func (e *WorkspaceRootError) Unwrap() error { return e.Cause }

// TildeExpansionError is returned when tilde expansion fails.
type TildeExpansionError struct {
	Cause error
}

func (e *TildeExpansionError) Error() string {
	return fmt.Sprintf("failed to expand tilde: %v", e.Cause)
}
func (e *TildeExpansionError) Unwrap() error { return e.Cause }

// LstatError is returned when lstat fails.
type LstatError struct {
	Path  string
	Cause error
}

func (e *LstatError) Error() string {
	return fmt.Sprintf("failed to lstat path %s: %v", e.Path, e.Cause)
}
func (e *LstatError) Unwrap() error { return e.Cause }

// ReadlinkError is returned when readlink fails.
type ReadlinkError struct {
	Path  string
	Cause error
}

func (e *ReadlinkError) Error() string {
	return fmt.Sprintf("failed to readlink path %s: %v", e.Path, e.Cause)
}
func (e *ReadlinkError) Unwrap() error { return e.Cause }

// -- Sentinels --

var (
	ErrOutsideWorkspace    = errors.New("path is outside workspace root")
	ErrWorkspaceRootNotSet = errors.New("workspace root not set")
	ErrSymlinkLoop         = errors.New("symlink loop detected")
	ErrSymlinkChainTooLong = errors.New("symlink chain too long")
	ErrNotADirectory       = errors.New("not a directory")
)
