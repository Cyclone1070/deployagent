package shell

import "errors"

// Shell operation errors
var (
	ErrShellTimeout                    = errors.New("shell command timed out")
	ErrShellWorkingDirOutsideWorkspace = errors.New("working directory is outside workspace")
)
