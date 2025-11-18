package tools

import (
	"os"
	"time"
)

// FileInfo represents file metadata
type FileInfo interface {
	Name() string
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	IsDir() bool
}

// FileSystem provides filesystem operations.
// ReadFileRange must respect size limits provided by the context.
// EnsureDirs should create parent directories recursively, but only within the workspace boundary.
type FileSystem interface {
	// Stat returns file info for a path (follows symlinks)
	Stat(path string) (FileInfo, error)
	// Lstat returns file info for a path without following symlinks
	Lstat(path string) (FileInfo, error)
	// ReadFileRange reads a range of bytes from a file.
	// If offset and limit are both 0, reads the entire file.
	// Must enforce size limits based on context configuration.
	ReadFileRange(path string, offset, limit int64) ([]byte, error)
	// WriteFile writes content to a file atomically
	WriteFile(path string, content []byte, perm os.FileMode) error
	// EnsureDirs creates parent directories if they don't exist.
	// Must only create directories within the workspace boundary.
	EnsureDirs(path string) error
	// IsDir checks if a path is a directory
	IsDir(path string) (bool, error)
	// Readlink reads the target of a symlink
	Readlink(path string) (string, error)
	// EvalSymlinks evaluates symlinks in a path, following chains
	EvalSymlinks(path string) (string, error)
	// Abs returns an absolute representation of path
	Abs(path string) (string, error)
	// UserHomeDir returns the current user's home directory
	UserHomeDir() (string, error)
}

// BinaryDetector checks if content is binary
type BinaryDetector interface {
	// IsBinary checks if a file contains binary data
	IsBinary(path string) (bool, error)
	// IsBinaryContent checks if content bytes contain binary data
	IsBinaryContent(content []byte) bool
}

// ChecksumComputer computes checksums
type ChecksumComputer interface {
	// ComputeChecksum computes SHA-256 checksum of data
	ComputeChecksum(data []byte) string
}

// Clock provides time operations
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// ChecksumStore provides checksum cache operations.
// Implementations must be thread-safe.
type ChecksumStore interface {
	// Get retrieves checksum for a file path
	Get(path string) (checksum string, ok bool)
	// Update stores or updates checksum for a file path
	Update(path string, checksum string)
	// Clear removes all cached checksums
	Clear()
}

// RootCanonicalizer canonicalizes workspace root paths.
// This interface allows dependency injection to avoid real filesystem operations in tests.
type RootCanonicalizer interface {
	// CanonicalizeRoot makes a path absolute, resolves symlinks, and validates it's a directory.
	CanonicalizeRoot(root string) (string, error)
}
