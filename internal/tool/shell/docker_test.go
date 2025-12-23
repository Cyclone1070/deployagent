package shell

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Cyclone1070/iav/internal/tool/service/executor"
)

// mockCommandExecutorForDocker is a local mock for testing docker functions
type mockCommandExecutorForDocker struct {
	runFunc func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error)
}

func (m *mockCommandExecutorForDocker) Run(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, cmd, dir, env)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCommandExecutorForDocker) RunWithTimeout(ctx context.Context, cmd []string, dir string, env []string, timeout time.Duration) (*executor.Result, error) {
	return nil, errors.New("not implemented")
}

func TestIsDockerCommand(t *testing.T) {
	tests := []struct {
		name    string
		command []string
		want    bool
	}{
		{"Empty", []string{}, false},
		{"SimpleDocker", []string{"docker", "run"}, true},
		{"FullPathDocker", []string{"/usr/bin/docker", "ps"}, true},
		{"OtherCommand", []string{"ls", "-la"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDockerCommand(tt.command); got != tt.want {
				t.Errorf("IsDockerCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDockerComposeUpDetached(t *testing.T) {
	tests := []struct {
		name    string
		command []string
		want    bool
	}{
		{"NotDocker", []string{"ls", "-la"}, false},
		{"JustUp", []string{"docker", "compose", "up"}, false},
		{"UpDetachedShort", []string{"docker", "compose", "up", "-d"}, true},
		{"UpDetachedLong", []string{"docker", "compose", "up", "--detach"}, true},
		{"UpDetachedMixedOrder", []string{"docker", "compose", "-f", "docker-compose.yml", "up", "-d"}, true},
		{"NotCompose", []string{"docker", "run", "-d"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDockerComposeUpDetached(tt.command); got != tt.want {
				t.Errorf("IsDockerComposeUpDetached() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatContainerStartedNote(t *testing.T) {
	tests := []struct {
		name string
		ids  []string
		want string
	}{
		{"Empty", []string{}, ""},
		{"Single", []string{"abc12345"}, "Started 1 Docker container: abc12345"},
		{"Multiple", []string{"abc", "def"}, "Started 2 Docker containers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatContainerStartedNote(tt.ids); got != tt.want {
				t.Errorf("FormatContainerStartedNote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureDockerReady(t *testing.T) {
	config := DockerConfig{
		CheckCommand: []string{"docker", "info"},
		StartCommand: []string{"open", "-a", "Docker"},
	}

	t.Run("Success immediately", func(t *testing.T) {
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			if cmd[0] == "docker" && cmd[1] == "info" {
				return &executor.Result{Stdout: "", ExitCode: 0}, nil
			}
			return nil, errors.New("unexpected command")
		}
		err := EnsureDockerReady(context.Background(), runner, config, 5, 10)
		if err != nil {
			t.Errorf("EnsureDockerReady failed: %v", err)
		}
	})

	t.Run("Start required and succeeds", func(t *testing.T) {
		checkCalls := 0
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			if cmd[0] == "docker" && cmd[1] == "info" {
				checkCalls++
				if checkCalls == 1 {
					return &executor.Result{Stdout: "", ExitCode: 1}, errors.New("docker not running")
				}
				return &executor.Result{Stdout: "", ExitCode: 0}, nil // Success on second call
			}
			if cmd[0] == "open" {
				return &executor.Result{Stdout: "", ExitCode: 0}, nil // Start command succeeds
			}
			return nil, errors.New("unexpected command")
		}

		err := EnsureDockerReady(context.Background(), runner, config, 5, 10)
		if err != nil {
			t.Errorf("EnsureDockerReady failed: %v", err)
		}
		if checkCalls != 2 {
			t.Errorf("Expected 2 check calls, got %d", checkCalls)
		}
	})

	t.Run("Start fails", func(t *testing.T) {
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			return &executor.Result{Stdout: "", ExitCode: 1}, errors.New("command failed")
		}
		err := EnsureDockerReady(context.Background(), runner, config, 5, 10)
		if err == nil {
			t.Error("EnsureDockerReady succeeded, want error")
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		runner := &mockCommandExecutorForDocker{}
		checkCount := 0
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			if cmd[0] == "docker" {
				checkCount++
				if checkCount == 1 {
					// First check fails needs start
					return &executor.Result{Stdout: "", ExitCode: 1}, nil
				}
				// Retry loop...
				cancel() // Cancel context during retry
				return &executor.Result{Stdout: "", ExitCode: 1}, nil
			}
			if cmd[0] == "open" {
				return &executor.Result{Stdout: "", ExitCode: 0}, nil
			}
			return nil, nil
		}

		// Short interval
		err := EnsureDockerReady(ctx, runner, config, 10, 1)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestCollectComposeContainers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			if len(cmd) > 2 && cmd[1] == "compose" && cmd[len(cmd)-2] == "ps" {
				return &executor.Result{Stdout: "id1\nid2\n", ExitCode: 0}, nil
			}
			return nil, errors.New("unexpected command")
		}
		ids, err := CollectComposeContainers(context.Background(), runner, ".")
		if err != nil {
			t.Fatalf("CollectComposeContainers failed: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("Expected 2 ids, got %d", len(ids))
		}
		if ids[0] != "id1" || ids[1] != "id2" {
			t.Errorf("Unexpected ids: %v", ids)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			return &executor.Result{Stdout: "", ExitCode: 0}, nil
		}
		ids, err := CollectComposeContainers(context.Background(), runner, ".")
		if err != nil {
			t.Fatalf("CollectComposeContainers failed: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("Expected 0 ids, got %d", len(ids))
		}
	})

	t.Run("Error", func(t *testing.T) {
		runner := &mockCommandExecutorForDocker{}
		runner.runFunc = func(ctx context.Context, cmd []string, dir string, env []string) (*executor.Result, error) {
			return nil, errors.New("command failed")
		}
		_, err := CollectComposeContainers(context.Background(), runner, ".")
		if err == nil {
			t.Error("CollectComposeContainers succeeded, want error")
		}
	})
}
