package services

import (
	"context"
	"strings"
	"time"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

// EnsureDockerReady checks if Docker is running and attempts to start it if not.
func EnsureDockerReady(ctx context.Context, runner models.CommandRunner, config models.DockerConfig) error {
	// 1. Check if Docker is running
	if _, err := runner.Run(ctx, config.CheckCommand); err == nil {
		return nil
	}

	// 2. Attempt to start Docker
	if _, err := runner.Run(ctx, config.StartCommand); err != nil {
		return err
	}

	// 3. Wait for Docker to be ready
	// Retry up to 10 times with 1s delay (simplified for now)
	// In a real app, we might want this configurable or use a backoff.
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			if _, err := runner.Run(ctx, config.CheckCommand); err == nil {
				return nil
			}
		}
	}

	_, err := runner.Run(ctx, config.CheckCommand)
	return err // Return the last error
}

// CollectComposeContainers returns a list of container IDs for a compose project in the given directory.
// It assumes 'docker compose ps -q' returns one ID per line.
// Note: This function needs a runner that captures output.
// The current CommandRunner interface only returns error.
// We might need a separate interface or extend CommandRunner for output capture.
// However, for this helper, we can just use a specific runner or change the interface.
// Given the plan, let's assume we need to extend CommandRunner or use a specific one.
// But wait, the plan says `CollectComposeContainers(ctx, runner, dir)`.
// If `runner.Run` doesn't return output, we can't get IDs.
// Let's update CommandRunner to support output or add a new method.
// Or, for now, since `EnsureDockerReady` is the main one using `CommandRunner` generic,
// maybe `CollectComposeContainers` should take a `CommandOutputRunner`.

// Let's define a local interface for output running if needed, or just update models.CommandRunner?
// Updating models.CommandRunner might break other things if they expect simple Run.
// Let's check `shell.go` plan. It uses `exec.Command` directly.
// `CollectComposeContainers` is a helper.
// Let's add `RunOutput` to `CommandRunner`? Or just `Run`?
// If we change `Run` to return `([]byte, error)`, `EnsureDockerReady` ignores output.
// That seems fine.

// Let's update `models.CommandRunner` to return `([]byte, error)`?
// Or add `RunOutput`.
// Let's stick to `Run` returning error for now, and maybe `CollectComposeContainers` isn't fully implementable
// without output capture.
// Actually, `EnsureDockerReady` uses `runner.Run`.
// `CollectComposeContainers` needs output.
// Let's modify `CommandRunner` to `Run(ctx, cmd) ([]byte, error)`.
// This is a breaking change for the interface I just added, but I haven't used it much yet.
// `docker_helper_test.go` uses `RunFunc` returning error. I'll need to update that too.

// ALTERNATIVE: `CollectComposeContainers` takes a `func(ctx, cmd) ([]byte, error)`.
// Or `CommandOutputRunner`.

// Let's update `models.CommandRunner` to return `([]byte, error)`. It's more useful.
// I will update `interfaces.go` and `docker_helper_test.go` in the next steps.
// For now, I'll write `docker_helper.go` assuming `Run` returns `([]byte, error)`.

func CollectComposeContainers(ctx context.Context, runner models.CommandRunner, dir string) ([]string, error) {
	// We need to run in a specific directory.
	// The CommandRunner interface doesn't support Dir.
	// This abstraction is leaking.
	// Maybe `CollectComposeContainers` should just take the `ShellTool` or similar?
	// Or `runner` should be configured with Dir?
	// Or `cmd` should include `cd`? No.

	// Let's assume the runner handles execution.
	// If we need Dir, we might need `RunInDir(ctx, dir, cmd)`.

	// For now, let's skip `CollectComposeContainers` implementation details regarding Dir
	// and just focus on the command.
	// Actually, `docker compose` can take `-f` or `--project-directory`.
	// So we can pass `["docker", "compose", "--project-directory", dir, "ps", "-q"]`.

	cmd := []string{"docker", "compose", "--project-directory", dir, "ps", "-q"}
	output, err := runner.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var ids []string
	for _, line := range lines {
		if line != "" {
			ids = append(ids, strings.TrimSpace(line))
		}
	}
	return ids, nil
}
