package file

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tool/pathutil"
)

// fileEditor defines the minimal filesystem operations needed for editing files.
type fileEditor interface {
	Stat(path string) (os.FileInfo, error)
	ReadFileRange(path string, offset, limit int64) ([]byte, error)
	WriteFileAtomic(path string, content []byte, perm os.FileMode) error
}

// checksumManager defines the interface for full checksum management.
type checksumManager interface {
	Compute(data []byte) string
	Get(path string) (checksum string, ok bool)
	Update(path string, checksum string)
}

// EditFileTool handles file editing operations.
type EditFileTool struct {
	fileOps         fileEditor
	pathResolver    pathResolver
	binaryDetector  binaryDetector
	checksumManager checksumManager
	config          *config.Config
	workspaceRoot   string
}

// NewEditFileTool creates a new EditFileTool with injected dependencies.
func NewEditFileTool(
	fileOps fileEditor,
	pathResolver pathResolver,
	binaryDetector binaryDetector,
	checksumManager checksumManager,
	cfg *config.Config,
	workspaceRoot string,
) *EditFileTool {
	return &EditFileTool{
		fileOps:         fileOps,
		pathResolver:    pathResolver,
		binaryDetector:  binaryDetector,
		checksumManager: checksumManager,
		config:          cfg,
		workspaceRoot:   workspaceRoot,
	}
}

// Run applies edit operations to an existing file in the workspace.
// It detects concurrent modifications by comparing file checksums and validates
// operations before applying them. The file is written atomically.
//
// Note: There is a narrow race condition window between checksum validation and write.
// For guaranteed conflict-free edits, external file locking would be required.
//
// Note: ctx is accepted for API consistency but not used - file I/O is synchronous.
func (t *EditFileTool) Run(ctx context.Context, req EditFileRequest) (*EditFileResponse, error) {
	// Resolve path
	abs, rel, err := pathutil.Resolve(t.workspaceRoot, t.pathResolver, req.Path)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	info, err := t.fileOps.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileMissing
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read full file (single open+read syscall)
	contentBytes, err := t.fileOps.ReadFileRange(abs, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for binary using content we already read
	if t.binaryDetector.IsBinaryContent(contentBytes) {
		return nil, ErrBinaryFile
	}

	content := string(contentBytes)

	// Compute current checksum
	currentChecksum := t.checksumManager.Compute(contentBytes)

	// Check for conflicts with cached version
	priorChecksum, ok := t.checksumManager.Get(abs)
	if ok && priorChecksum != currentChecksum {
		return nil, ErrEditConflict
	}

	// Preserve original permissions
	originalPerm := info.Mode()

	// Apply operations sequentially
	operationsApplied := 0
	for _, op := range req.Operations {
		// Apply default ExpectedReplacements if not specified (0 = omitted)
		if op.ExpectedReplacements == 0 {
			op.ExpectedReplacements = 1
		}

		count := strings.Count(content, op.Before)

		if count == 0 {
			return nil, ErrSnippetNotFound
		}

		if count != op.ExpectedReplacements {
			return nil, ErrExpectedReplacementsMismatch
		}

		content = strings.Replace(content, op.Before, op.After, op.ExpectedReplacements)
		operationsApplied++
	}

	newContentBytes := []byte(content)

	// Check size limit
	maxFileSize := t.config.Tools.MaxFileSize
	if int64(len(newContentBytes)) > maxFileSize {
		return nil, ErrTooLarge
	}

	// Only revalidate if we had a cached checksum to check against
	// This optimizes the common case where files are edited without being read first
	if ok {
		revalidationBytes, err := t.fileOps.ReadFileRange(abs, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to revalidate file before write: %w", err)
		}
		revalidationChecksum := t.checksumManager.Compute(revalidationBytes)
		if revalidationChecksum != currentChecksum {
			return nil, ErrEditConflict
		}
	}

	// Write the modified content atomically
	if err := t.fileOps.WriteFileAtomic(abs, newContentBytes, originalPerm); err != nil {
		return nil, fmt.Errorf("failed to write edited file: %w", err)
	}

	// Compute new checksum and update cache
	newChecksum := t.checksumManager.Compute(newContentBytes)
	t.checksumManager.Update(abs, newChecksum)

	return &EditFileResponse{
		AbsolutePath:      abs,
		RelativePath:      rel,
		OperationsApplied: operationsApplied,
		FileSize:          int64(len(newContentBytes)),
	}, nil
}
