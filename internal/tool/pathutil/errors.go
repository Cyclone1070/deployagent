package pathutil

// OutsideWorkspaceError indicates a path is outside the workspace boundary.
type OutsideWorkspaceError struct{}

func (e *OutsideWorkspaceError) Error() string {
	return "path is outside workspace root"
}

// OutsideWorkspace implements the behavioral interface for cross-package error checking.
func (e *OutsideWorkspaceError) OutsideWorkspace() bool {
	return true
}

// ErrOutsideWorkspace is returned when a path escapes the workspace boundary.
var ErrOutsideWorkspace = &OutsideWorkspaceError{}
