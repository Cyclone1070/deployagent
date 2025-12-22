package executil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

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

// Process defines the interface for a running process.
type Process interface {
	Wait() error
	Kill() error
	Signal(sig os.Signal) error
}

// osProcess wraps an os/exec.Cmd to implement the Process interface.
type osProcess struct {
	cmd *exec.Cmd
}

func (p *osProcess) Wait() error {
	return p.cmd.Wait()
}

func (p *osProcess) Kill() error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *osProcess) Signal(sig os.Signal) error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Signal(sig)
	}
	return nil
}

// CommandExecutor defines the interface for executing commands.
type CommandExecutor interface {
	Run(ctx context.Context, command []string) ([]byte, error)
	Start(ctx context.Context, command []string, dir string, env []string) (Process, io.Reader, io.Reader, error)
}

// OSCommandExecutor implements CommandExecutor using os/exec for real system commands.
type OSCommandExecutor struct{}

// NewOSCommandExecutor creates a new OSCommandExecutor.
func NewOSCommandExecutor() *OSCommandExecutor {
	return &OSCommandExecutor{}
}

// Run executes a command and returns the combined output (stdout + stderr).
func (f *OSCommandExecutor) Run(ctx context.Context, command []string) ([]byte, error) {
	if len(command) == 0 {
		return nil, os.ErrInvalid
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Stdin = nil

	return cmd.CombinedOutput()
}

// Start starts a process and returns control immediately for streaming output or process management.
func (f *OSCommandExecutor) Start(ctx context.Context, command []string, dir string, env []string) (Process, io.Reader, io.Reader, error) {
	if len(command) == 0 {
		return nil, nil, nil, os.ErrInvalid
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = nil

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	return &osProcess{cmd: cmd}, stdout, stderr, nil
}
