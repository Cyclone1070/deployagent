package services

import (
	"path/filepath"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
)

// GetCommandRoot extracts the root command (basename) from a command slice.
// e.g., ["/usr/bin/docker", "run"] -> "docker"
func GetCommandRoot(command []string) string {
	if len(command) == 0 {
		return ""
	}
	// Handle paths like /usr/bin/docker
	return filepath.Base(command[0])
}

// IsDockerCommand checks if the command is a docker command.
func IsDockerCommand(command []string) bool {
	return GetCommandRoot(command) == "docker"
}

// IsDockerComposeUpDetached checks if the command is 'docker compose up -d'.
func IsDockerComposeUpDetached(command []string) bool {
	if !IsDockerCommand(command) {
		return false
	}

	// Need at least "docker", "compose", "up"
	if len(command) < 3 {
		return false
	}

	// Check for "compose" and "up"
	// We need to be careful about flags appearing before subcommands, but for now
	// we'll assume the standard "docker compose up" structure.
	// A more robust parser would be needed for complex cases, but this covers the happy path.

	// Simple check: does it contain "compose" and "up" in that order?
	// And does it have -d or --detach?

	hasCompose := false
	hasUp := false
	hasDetach := false

	for _, arg := range command[1:] {
		if arg == "compose" {
			hasCompose = true
		} else if hasCompose && arg == "up" {
			hasUp = true
		} else if hasUp && (arg == "-d" || arg == "--detach") {
			hasDetach = true
		}
	}

	return hasCompose && hasUp && hasDetach
}

// EvaluatePolicy checks if a command is allowed by the given policy.
func EvaluatePolicy(policy models.CommandPolicy, command []string) error {
	root := GetCommandRoot(command)
	if root == "" {
		return models.ErrShellRejected
	}

	// 1. Check SessionAllow (Override)
	if policy.SessionAllow != nil && policy.SessionAllow[root] {
		return nil
	}

	// 2. Check Allow List
	for _, allowed := range policy.Allow {
		if allowed == root {
			return nil
		}
	}

	// 3. Check Ask List
	for _, ask := range policy.Ask {
		if ask == root {
			return models.ErrShellApprovalRequired
		}
	}

	// 4. Default Deny
	return models.ErrShellRejected
}
