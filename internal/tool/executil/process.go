package executil

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// TimeoutError is returned when a command exceeds its timeout.
type TimeoutError struct {
	Command  []string
	Duration time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("command %v timed out after %v", e.Command, e.Duration)
}
func (e *TimeoutError) Timeout() bool { return true }

// ExecuteWithTimeout runs a process with a timeout, handling graceful shutdown.
func ExecuteWithTimeout(ctx context.Context, command []string, timeout time.Duration, gracefulShutdownMs int, proc Process) error {
	done := make(chan error, 1)
	go func() {
		done <- proc.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = proc.Kill()
		return ctx.Err()
	case <-time.After(timeout):
		// Try graceful shutdown first
		_ = proc.Signal(os.Interrupt)

		select {
		case <-done:
			return &TimeoutError{Command: command, Duration: timeout}
		case <-time.After(time.Duration(gracefulShutdownMs) * time.Millisecond):
			_ = proc.Kill()
			return &TimeoutError{Command: command, Duration: timeout}
		}
	}
}

// CollectProcessOutput reads stdout and stderr concurrently and returns them as strings.
func CollectProcessOutput(stdout, stderr io.Reader, maxBytes int, sampleSize int) (string, string, bool, error) {
	stdoutCollector := newCollector(maxBytes, sampleSize)
	stderrCollector := newCollector(maxBytes, sampleSize)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(stdoutCollector, stdout)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(stderrCollector, stderr)
	}()

	wg.Wait()

	truncated := stdoutCollector.Truncated() || stderrCollector.Truncated()
	return stdoutCollector.String(), stderrCollector.String(), truncated, nil
}

// GetExitCode extracts the exit code from an error returned by a process.
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	type exitCoder interface {
		ExitCode() int
	}
	if ec, ok := err.(exitCoder); ok {
		return ec.ExitCode()
	}

	return -1
}
