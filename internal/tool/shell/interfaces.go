package shell

import (
	"context"
	"os"
	"time"

	"github.com/Cyclone1070/iav/internal/tool/executor"
)

// fileSystem defines the minimal filesystem interface needed by shell tools.
// This is a consumer-defined interface per architecture guidelines §2.
type fileSystem interface {
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)
	Readlink(path string) (string, error)
	UserHomeDir() (string, error)
	ReadFileRange(path string, offset, limit int64) ([]byte, error)
}

// commandExecutor defines the interface for executing shell commands.
type commandExecutor interface {
	Run(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error)
	RunWithTimeout(ctx context.Context, cmd []string, dir string, env []string, timeout time.Duration) (*executor.Result, error)
}
