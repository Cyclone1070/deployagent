package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultMaxFileSize is the default maximum file size (5 MB)
	DefaultMaxFileSize = 5 * 1024 * 1024
	// BinaryDetectionSampleSize is how many bytes to sample for binary detection
	BinaryDetectionSampleSize = 4096
)

// OSFileSystem implements FileSystem using the local OS primitives.
// It enforces file size limits based on the MaxFileSize field.
type OSFileSystem struct {
	MaxFileSize int64
}

// NewOSFileSystem creates a new OSFileSystem with the specified max file size.
func NewOSFileSystem(maxFileSize int64) *OSFileSystem {
	return &OSFileSystem{
		MaxFileSize: maxFileSize,
	}
}

func (r *OSFileSystem) Stat(path string) (FileInfo, error) {
	return os.Stat(path)
}

func (r *OSFileSystem) Lstat(path string) (FileInfo, error) {
	return os.Lstat(path)
}

func (r *OSFileSystem) ReadFileRange(path string, offset, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := info.Size()

	// Check size limit
	if fileSize > r.MaxFileSize {
		return nil, ErrTooLarge
	}

	// If both offset and limit are 0, read entire file
	if offset == 0 && limit == 0 {
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return content, nil
	}

	// Validate offset
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	if offset >= fileSize {
		return []byte{}, nil
	}

	// Seek to offset
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Calculate how much to read
	remaining := fileSize - offset
	var readSize int64

	if limit == 0 {
		readSize = remaining
	} else {
		if remaining < limit {
			readSize = remaining
		} else {
			readSize = limit
		}
	}

	content := make([]byte, readSize)
	n, err := file.Read(content)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return content[:n], nil
}

func (r *OSFileSystem) WriteFile(path string, content []byte, perm os.FileMode) error {
	return writeFileAtomic(path, content, perm)
}

// writeFileAtomic writes content to a file atomically using temp file + rename pattern.
// This ensures that if the process crashes mid-write, the original file remains intact.
func writeFileAtomic(path string, content []byte, perm os.FileMode) error {
	// Get directory for temp file
	dir := filepath.Dir(path)

	// Create temporary file in same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write content to temp file
	if _, err := tmpFile.Write(content); err != nil {
		return err
	}

	// Sync to ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	// Close file before rename (required on some systems)
	if err := tmpFile.Close(); err != nil {
		return err
	}
	tmpFile = nil // Prevent cleanup in defer

	// Atomic rename - this is the critical operation that makes it atomic
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Set permissions on the final file
	if err := os.Chmod(path, perm); err != nil {
		return err
	}

	return nil
}

func (r *OSFileSystem) EnsureDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func (r *OSFileSystem) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (r *OSFileSystem) Readlink(path string) (string, error) {
	return os.Readlink(path)
}

func (r *OSFileSystem) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func (r *OSFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (r *OSFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// SystemBinaryDetector implements BinaryDetector using local heuristics
type SystemBinaryDetector struct{}

func (r *SystemBinaryDetector) IsBinary(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buf := make([]byte, BinaryDetectionSampleSize)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}

func (r *SystemBinaryDetector) IsBinaryContent(content []byte) bool {
	sampleSize := BinaryDetectionSampleSize
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	for i := 0; i < sampleSize; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}

// SHA256Checksum implements ChecksumComputer using SHA-256
type SHA256Checksum struct{}

func (r *SHA256Checksum) ComputeChecksum(data []byte) string {
	return computeChecksum(data)
}

// computeChecksum computes SHA-256 checksum of data
func computeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SystemClock implements Clock using time.Now
type SystemClock struct{}

func (r *SystemClock) Now() time.Time {
	return time.Now()
}
