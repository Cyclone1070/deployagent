package err

import (
	"fmt"
	"os"
	"time"
)

// -- Common System Errors --

// StatError is returned when stat fails.
type StatError struct {
	Path  string
	Cause error
}

func (e *StatError) Error() string { return fmt.Sprintf("failed to stat %s: %v", e.Path, e.Cause) }
func (e *StatError) Unwrap() error { return e.Cause }

// CommandError represents generic command execution failures (start, output, wait).
type CommandError struct {
	Cmd   string
	Cause error
	Stage string // "start", "read output", "execution"
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command %s failed at %s: %v", e.Cmd, e.Stage, e.Cause)
}
func (e *CommandError) Unwrap() error { return e.Cause }

// ReadError is returned when read fails.
type ReadError struct {
	Path  string
	Cause error
}

func (e *ReadError) Error() string { return fmt.Sprintf("failed to read file %s: %v", e.Path, e.Cause) }
func (e *ReadError) Unwrap() error { return e.Cause }

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

// EnsureDirsError is returned when mkdir fails.
type EnsureDirsError struct {
	Path  string
	Cause error
}

func (e *EnsureDirsError) Error() string {
	return fmt.Sprintf("failed to ensure directories for %s: %v", e.Path, e.Cause)
}
func (e *EnsureDirsError) Unwrap() error { return e.Cause }

// WriteError is returned when write fails.
type WriteError struct {
	Path  string
	Cause error
}

func (e *WriteError) Error() string {
	return fmt.Sprintf("failed to write file %s: %v", e.Path, e.Cause)
}
func (e *WriteError) Unwrap() error { return e.Cause }

// RevalidateError is returned when revalidation fails.
type RevalidateError struct {
	Path  string
	Cause error
}

func (e *RevalidateError) Error() string {
	return fmt.Sprintf("failed to revalidate file %s: %v", e.Path, e.Cause)
}
func (e *RevalidateError) Unwrap() error { return e.Cause }

// -- Feature Specific Errors --

// WorkspaceRootError is returned when the workspace root is invalid.
type WorkspaceRootError struct {
	Root  string
	Cause error
}

func (e *WorkspaceRootError) Error() string {
	return fmt.Sprintf("invalid workspace root %s: %v", e.Root, e.Cause)
}
func (e *WorkspaceRootError) Unwrap() error { return e.Cause }

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

// TildeExpansionError is returned when tilde expansion fails.
type TildeExpansionError struct {
	Cause error
}

func (e *TildeExpansionError) Error() string {
	return fmt.Sprintf("failed to expand tilde: %v", e.Cause)
}
func (e *TildeExpansionError) Unwrap() error { return e.Cause }

// TimeoutError is returned when a command exceeds its timeout.
type TimeoutError struct {
	Command  []string
	Duration time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("command %v timed out after %v", e.Command, e.Duration)
}
func (e *TimeoutError) Timeout() bool { return true }

// GitignoreReadError is returned when .gitignore cannot be read.
type GitignoreReadError struct {
	Path  string
	Cause error
}

func (e *GitignoreReadError) Error() string {
	return fmt.Sprintf("failed to read .gitignore at %s: %v", e.Path, e.Cause)
}
func (e *GitignoreReadError) Unwrap() error { return e.Cause }

// StoreReadError is returned when reading from a store (e.g., todo) fails.
type StoreReadError struct {
	Cause error
}

func (e *StoreReadError) Error() string { return fmt.Sprintf("failed to read store: %v", e.Cause) }
func (e *StoreReadError) Unwrap() error { return e.Cause }

// StoreWriteError is returned when writing to a store fails.
type StoreWriteError struct {
	Cause error
}

func (e *StoreWriteError) Error() string { return fmt.Sprintf("failed to write store: %v", e.Cause) }
func (e *StoreWriteError) Unwrap() error { return e.Cause }

// -- FS Utility Errors --

// TempFileError is returned when creating a temp file fails.
type TempFileError struct {
	Dir   string
	Cause error
}

func (e *TempFileError) Error() string {
	return fmt.Sprintf("failed to create temp file in %s: %v", e.Dir, e.Cause)
}
func (e *TempFileError) Unwrap() error { return e.Cause }

// TempWriteError is returned when writing to a temp file fails.
type TempWriteError struct {
	Path  string
	Cause error
}

func (e *TempWriteError) Error() string {
	return fmt.Sprintf("failed to write to temp file %s: %v", e.Path, e.Cause)
}
func (e *TempWriteError) Unwrap() error { return e.Cause }

// TempSyncError is returned when syncing a temp file fails.
type TempSyncError struct {
	Path  string
	Cause error
}

func (e *TempSyncError) Error() string {
	return fmt.Sprintf("failed to sync temp file %s: %v", e.Path, e.Cause)
}
func (e *TempSyncError) Unwrap() error { return e.Cause }

// TempCloseError is returned when closing a temp file fails.
type TempCloseError struct {
	Path  string
	Cause error
}

func (e *TempCloseError) Error() string {
	return fmt.Sprintf("failed to close temp file %s: %v", e.Path, e.Cause)
}
func (e *TempCloseError) Unwrap() error { return e.Cause }

// RenameError is returned when renaming a file fails.
type RenameError struct {
	Old   string
	New   string
	Cause error
}

func (e *RenameError) Error() string {
	return fmt.Sprintf("failed to rename %s to %s: %v", e.Old, e.New, e.Cause)
}
func (e *RenameError) Unwrap() error { return e.Cause }

// ChmodError is returned when changing file permissions fails.
type ChmodError struct {
	Path  string
	Mode  os.FileMode
	Cause error
}

func (e *ChmodError) Error() string {
	return fmt.Sprintf("failed to set permissions for %s to %v: %v", e.Path, e.Mode, e.Cause)
}
func (e *ChmodError) Unwrap() error { return e.Cause }

// EnvFileReadError is returned when reading an env file fails.
type EnvFileReadError struct {
	Path  string
	Cause error
}

func (e *EnvFileReadError) Error() string {
	return fmt.Sprintf("failed to read env file %s: %v", e.Path, e.Cause)
}
func (e *EnvFileReadError) Unwrap() error { return e.Cause }
