package search

import (
	"context"
	"os"

	"github.com/Cyclone1070/iav/internal/tool/executor"
)

// fileSystem defines the minimal filesystem interface needed by search tools.
// This is a consumer-defined interface per architecture guidelines §2.
type fileSystem interface {
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
	Readlink(path string) (string, error)
	UserHomeDir() (string, error)
}

// commandExecutor defines the interface for executing search commands.
type commandExecutor interface {
	Run(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error)
}
