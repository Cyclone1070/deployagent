package services

import (
	"context"
	"os"
	"time"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

// ExecuteWithTimeout runs a process with a timeout.
// It assumes the process has already been started.
func ExecuteWithTimeout(ctx context.Context, timeout time.Duration, proc models.Process) error {
	done := make(chan error, 1)
	go func() {
		done <- proc.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Context cancelled (e.g. user cancellation)
		_ = proc.Kill()
		return ctx.Err()
	case <-time.After(timeout):
		// Timeout reached
		// Try graceful shutdown first
		_ = proc.Signal(os.Interrupt) // SIGINT/SIGTERM equivalent

		// Wait a bit for graceful shutdown
		select {
		case <-done:
			return models.ErrShellTimeout
		case <-time.After(2 * time.Second):
			_ = proc.Kill()
			return models.ErrShellTimeout
		}
	}
}
