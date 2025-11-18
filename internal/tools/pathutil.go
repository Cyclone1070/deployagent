package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CanonicaliseRoot canonicalises a workspace root path by making it absolute
// and resolving symlinks. Returns an error if the path doesn't exist or isn't a directory.
func CanonicaliseRoot(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace root: %w", err)
	}

	// Resolve symlinks in the workspace root to get canonical path
	resolved, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		// If symlink resolution fails, use the absolute path as-is
		resolved = absRoot
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("workspace root does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace root is not a directory: %s", resolved)
	}
	return resolved, nil
}

// Resolve normalises a path and ensures it's within the workspace root.
// It handles symlink resolution component-by-component, path traversal prevention,
// and validates that the resolved path stays within the workspace boundary.
// This prevents symlink escape attacks even when the final file doesn't exist.
func Resolve(ctx *WorkspaceContext, path string) (abs string, rel string, err error) {
	if ctx.WorkspaceRoot == "" {
		return "", "", fmt.Errorf("workspace root not set")
	}

	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		home, err := ctx.FS.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to expand tilde: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// If absolute, check it's within workspace
	if filepath.IsAbs(cleaned) {
		abs = cleaned
	} else {
		// Relative path - join with workspace root
		abs = filepath.Join(ctx.WorkspaceRoot, cleaned)
	}

	// Clean the absolute path
	abs = filepath.Clean(abs)

	// Resolve symlinks component-by-component to prevent escape attacks
	resolved, err := resolveSymlink(ctx, abs)
	if err != nil {
		return "", "", err
	}
	abs = resolved

	// WorkspaceRoot is already absolute and symlink-resolved
	workspaceRootAbs := filepath.Clean(ctx.WorkspaceRoot)

	// Calculate relative path
	rel, err = filepath.Rel(workspaceRootAbs, abs)
	if err != nil {
		workspaceRootWithSep := workspaceRootAbs + string(filepath.Separator)
		if abs == workspaceRootAbs {
			rel = "."
		} else if strings.HasPrefix(abs, workspaceRootWithSep) {
			rel = abs[len(workspaceRootWithSep):]
		} else {
			return "", "", ErrOutsideWorkspace
		}
	}

	// Segment-by-segment traversal validation
	relSegments := strings.SplitSeq(filepath.ToSlash(rel), "/")
	for segment := range relSegments {
		if segment == ".." {
			return "", "", ErrOutsideWorkspace
		}
	}

	// Normalise to use forward slashes for relative path
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = ""
	}

	return abs, rel, nil
}

// resolveSymlink resolves symlinks by walking each path component.
// This prevents symlink escape attacks even when the final file doesn't exist.
// It handles missing intermediate directories gracefully to allow directory creation.
// It follows symlink chains and validates that every hop stays within the workspace boundary.
func resolveSymlink(ctx *WorkspaceContext, path string) (string, error) {
	workspaceRootAbs := filepath.Clean(ctx.WorkspaceRoot)
	const maxHops = 64

	// Split path into components for component-wise traversal
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) == 0 {
		return path, nil
	}

	// Handle absolute paths (first component is empty on Unix)
	var currentAbs string
	startIdx := 0
	if filepath.IsAbs(path) {
		if len(parts) > 0 && parts[0] == "" {
			currentAbs = "/"
			startIdx = 1
		} else {
			currentAbs = path
		}
	} else {
		currentAbs = path
	}

	// Walk each component, resolving symlinks as we go
	for i := startIdx; i < len(parts); i++ {
		if parts[i] == "" || parts[i] == "." {
			continue
		}
		
		// Handle ".." by going up one directory level
		if parts[i] == ".." {
			// Go up one level
			if currentAbs == "" || currentAbs == "/" {
				// Can't go up from root
				return "", ErrOutsideWorkspace
			}
			currentAbs = filepath.Dir(currentAbs)
			// Validate we're still within workspace after going up
			if !isWithinWorkspace(currentAbs, workspaceRootAbs) {
				return "", ErrOutsideWorkspace
			}
			continue
		}

		// Build the next path component
		var next string
		switch currentAbs {
		case "":
			next = parts[i]
		case "/":
			next = "/" + parts[i]
		default:
			next = filepath.Join(currentAbs, parts[i])
		}

		// Follow symlink chain for this component
		visited := make(map[string]struct{})
		current := next
		hopCount := 0

		for {
			// Check hop count limit (enforces max 64 hops)
			if hopCount > maxHops {
				return "", fmt.Errorf("symlink chain too long (max %d hops)", maxHops)
			}

			// Check for loops
			if _, seen := visited[current]; seen {
				return "", fmt.Errorf("symlink loop detected: %s", current)
			}
			visited[current] = struct{}{}

			// Check if current path is a symlink
			info, err := ctx.FS.Lstat(current)
			if err != nil {
				// If component doesn't exist, handle missing directories
				if err == os.ErrNotExist {
					// Special case: if current equals workspace root, it's okay
					if current == workspaceRootAbs {
						currentAbs = current
						break
					}
					// If we're not at the final component, this means a directory is missing
					// Append remaining components and return (caller can create directories)
					if i < len(parts)-1 {
						// Append current and remaining components
						remaining := parts[i:]
						for j := range remaining {
							if remaining[j] == "" || remaining[j] == "." {
								continue
							}
							switch currentAbs {
							case "":
								currentAbs = remaining[j]
							case "/":
								currentAbs = "/" + remaining[j]
							default:
								currentAbs = filepath.Join(currentAbs, remaining[j])
							}
						}
						// Validate the complete path is within workspace
						if !isWithinWorkspace(currentAbs, workspaceRootAbs) {
							return "", ErrOutsideWorkspace
						}
						return currentAbs, nil
					}
					// For final component, validate parent is within workspace (if we have one)
					if currentAbs != "" && currentAbs != workspaceRootAbs {
						if !isWithinWorkspace(currentAbs, workspaceRootAbs) {
							return "", ErrOutsideWorkspace
						}
					}
					currentAbs = current
					break
				}
				return "", fmt.Errorf("failed to lstat path: %w", err)
			}

			// If not a symlink, we're done with this component
			if info.Mode()&os.ModeSymlink == 0 {
				// Validate path is within workspace
				if !isWithinWorkspace(current, workspaceRootAbs) {
					return "", ErrOutsideWorkspace
				}
				currentAbs = current
				break
			}

			// Read the symlink target
			linkTarget, err := ctx.FS.Readlink(current)
			if err != nil {
				return "", fmt.Errorf("failed to read symlink: %w", err)
			}

			// Resolve symlink target to absolute path
			var targetAbs string
			if filepath.IsAbs(linkTarget) {
				targetAbs = filepath.Clean(linkTarget)
			} else {
				// Relative symlink - resolve relative to symlink's directory
				targetAbs = filepath.Clean(filepath.Join(filepath.Dir(current), linkTarget))
			}

			// Validate symlink target is within workspace (reject immediately if outside)
			if !isWithinWorkspace(targetAbs, workspaceRootAbs) {
				return "", ErrOutsideWorkspace
			}

			// Continue following the chain
			current = targetAbs
			hopCount++
		}

		// Validate current path is within workspace after each step
		if !isWithinWorkspace(currentAbs, workspaceRootAbs) {
			return "", ErrOutsideWorkspace
		}
	}

	// Final validation that resolved path is within workspace
	if !isWithinWorkspace(currentAbs, workspaceRootAbs) {
		return "", ErrOutsideWorkspace
	}

	return currentAbs, nil
}

// isWithinWorkspace checks if a path is within the workspace root.
func isWithinWorkspace(path, workspaceRoot string) bool {
	workspaceRootAbs := filepath.Clean(workspaceRoot)
	pathAbs := filepath.Clean(path)

	// Check if path equals workspace root
	if pathAbs == workspaceRootAbs {
		return true
	}

	// Check if path is a subdirectory/file of workspace root
	rel, err := filepath.Rel(workspaceRootAbs, pathAbs)
	if err != nil {
		return false
	}

	// Check for path traversal attempts
	if strings.HasPrefix(rel, "..") {
		return false
	}

	// Ensure it's actually within (not just a sibling)
	workspaceRootWithSep := workspaceRootAbs + string(filepath.Separator)
	return strings.HasPrefix(pathAbs, workspaceRootWithSep) || pathAbs == workspaceRootAbs
}

// EnsureParentDirs creates parent directories for a given path if they don't exist.
// It validates that all parent directories remain within the workspace boundary.
func EnsureParentDirs(ctx *WorkspaceContext, path string) error {
	abs, _, err := Resolve(ctx, path)
	if err != nil {
		return err
	}

	parent := filepath.Dir(abs)
	if parent == abs {
		return nil
	}

	// Validate that parent directory is within workspace using symlink resolution
	_, err = resolveSymlink(ctx, parent)
	if err != nil {
		return err
	}

	return ctx.FS.EnsureDirs(parent)
}

// IsDirectory checks if a resolved path points to a directory.
func IsDirectory(ctx *WorkspaceContext, path string) (bool, error) {
	abs, _, err := Resolve(ctx, path)
	if err != nil {
		return false, err
	}

	return ctx.FS.IsDir(abs)
}
