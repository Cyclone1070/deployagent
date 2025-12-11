package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Cyclone1070/iav/internal/tools/model"
	"github.com/Cyclone1070/iav/internal/tools/service"
)

// ShellTool executes commands on the local machine.
type ShellTool struct {
	CommandExecutor model.CommandExecutor
}

// Run executes a shell command with Docker readiness checks,
// environment variable support, timeout handling, and output collection.
// NOTE: This tool does NOT enforce policy - the caller is responsible for policy checks.
func (t *ShellTool) Run(ctx context.Context, wCtx *model.WorkspaceContext, req model.ShellRequest) (*model.ShellResponse, error) {

	workingDir := req.WorkingDir
	if workingDir == "" {
		workingDir = "."
	}

	wd, _, err := service.Resolve(wCtx, workingDir)
	if err != nil {
		return nil, model.ErrShellWorkingDirOutsideWorkspace
	}

	// Policy check removed - caller is responsible for enforcement

	if service.IsDockerCommand(req.Command) {
		retryAttempts := wCtx.Config.Tools.DockerRetryAttempts
		retryIntervalMs := wCtx.Config.Tools.DockerRetryIntervalMs

		if err := service.EnsureDockerReady(ctx, t.CommandExecutor, wCtx.DockerConfig, retryAttempts, retryIntervalMs); err != nil {
			return nil, err
		}
	}

	env := os.Environ()

	for _, envFile := range req.EnvFiles {
		envFilePath, _, err := service.Resolve(wCtx, envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve env file %s: %w", envFile, err)
		}

		envVars, err := service.ParseEnvFile(wCtx.FS, envFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse env file %s: %w", envFile, err)
		}

		// EnvFiles override system env
		for k, v := range envVars {
			env = append(env, k+"="+v)
		}
	}

	// Request.Env overrides everything
	for k, v := range req.Env {
		env = append(env, k+"="+v)
	}

	opts := model.ProcessOptions{
		Dir: wd,
		Env: env,
	}

	proc, stdout, stderr, err := t.CommandExecutor.Start(ctx, req.Command, opts)
	if err != nil {
		return nil, err
	}

	// Use configured binary detection sample size
	sampleSize := wCtx.Config.Tools.BinaryDetectionSampleSize
	// Use configured max output size
	maxOutputSize := wCtx.Config.Tools.DefaultMaxCommandOutputSize

	stdoutStr, stderrStr, truncated, _ := service.CollectProcessOutput(stdout, stderr, int(maxOutputSize), sampleSize)

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = time.Duration(wCtx.Config.Tools.DefaultShellTimeout) * time.Second
	}

	// Use configured graceful shutdown
	gracefulShutdownMs := wCtx.Config.Tools.DockerGracefulShutdownMs

	execErr := service.ExecuteWithTimeout(ctx, timeout, gracefulShutdownMs, proc)

	resp := &model.ShellResponse{
		Stdout:     stdoutStr,
		Stderr:     stderrStr,
		WorkingDir: wd,
		Truncated:  truncated,
	}

	if execErr != nil {
		if execErr == model.ErrShellTimeout {
			resp.ExitCode = -1
			return resp, model.ErrShellTimeout
		}
		// Check for context cancellation
		if errors.Is(execErr, context.Canceled) || errors.Is(execErr, context.DeadlineExceeded) {
			resp.ExitCode = -1
			return resp, execErr
		}
		// Command ran but failed - extract exit code and return success
		resp.ExitCode = service.GetExitCode(execErr)
		return resp, nil
	}

	resp.ExitCode = 0

	if service.IsDockerComposeUpDetached(req.Command) {
		ids, err := service.CollectComposeContainers(ctx, t.CommandExecutor, wd)
		if err == nil {
			resp.Notes = append(resp.Notes, service.FormatContainerStartedNote(ids))
		} else {
			resp.Notes = append(resp.Notes, fmt.Sprintf("Warning: Could not list started containers: %v", err))
		}
	}

	return resp, nil
}
