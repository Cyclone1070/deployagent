package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

// We need to mock exec.Cmd for pure unit tests.
// However, ExecuteWithTimeout takes *exec.Cmd struct, which is hard to mock directly
// because it's a struct, not an interface.
// The plan says "MockCmd".
// To test `ExecuteWithTimeout` which takes `*exec.Cmd`, we actually need to run real processes
// OR we need to wrap `exec.Cmd` in an interface.
// The plan says "NO real process execution".
// So we MUST wrap `exec.Cmd`.
// But `ExecuteWithTimeout` signature in plan is `func ExecuteWithTimeout(ctx, timeout, cmd *exec.Cmd)`.
// This contradicts "Pure DI".
// I should have caught this.
// I will change `ExecuteWithTimeout` to take an interface `ProcessRunner` or similar.
// Or better, `ExecuteWithTimeout` logic is: Start, Wait, Timer -> Signal.
// This logic is tightly coupled to `os.Process`.
// We can define an interface `ProcessControl` that `exec.Cmd` satisfies (via a wrapper).

type MockProcess struct {
	WaitDelay    time.Duration
	WaitError    error
	KillCalled   bool
	SignalCalled bool
}

func (m *MockProcess) Wait() error {
	time.Sleep(m.WaitDelay)
	return m.WaitError
}

func (m *MockProcess) Kill() error {
	m.KillCalled = true
	return nil
}

func (m *MockProcess) Signal(sig os.Signal) error {
	m.SignalCalled = true
	return nil
}

// We need to change ExecuteWithTimeout signature to use an interface.
// Let's define it in timeout.go.

func TestExecuteWithTimeout_Success(t *testing.T) {
	// This test will fail to compile until we define the interface and function.
	// I'll write the test assuming the interface exists.

	mock := &MockProcess{
		WaitDelay: 10 * time.Millisecond,
	}

	err := ExecuteWithTimeout(context.Background(), 100*time.Millisecond, mock)
	if err != nil {
		t.Errorf("ExecuteWithTimeout failed: %v", err)
	}
}

func TestExecuteWithTimeout_Fail(t *testing.T) {
	mock := &MockProcess{
		WaitDelay: 200 * time.Millisecond,
	}

	err := ExecuteWithTimeout(context.Background(), 50*time.Millisecond, mock)
	if err != models.ErrShellTimeout {
		t.Errorf("Error = %v, want ErrShellTimeout", err)
	}
	if !mock.SignalCalled {
		t.Error("Signal (SIGTERM) not called")
	}
	// Kill might be called after SIGTERM delay, but our mock Wait returns after 200ms.
	// The timeout logic waits 2s after SIGTERM.
	// So Kill should be called.
}
