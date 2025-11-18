package tools

import (
	"fmt"
	"os"
)

// WriteFile creates a new file using injected dependencies
func WriteFile(ctx *WorkspaceContext, path string, content string, perm *os.FileMode) (*WriteFileResponse, error) {
	// Resolve path
	abs, rel, err := Resolve(ctx, path)
	if err != nil {
		return nil, err
	}

	// Check if file already exists
	_, err = ctx.FS.Stat(abs)
	if err == nil {
		return nil, ErrFileExists
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}

	// Ensure parent directories exist
	if err := EnsureParentDirs(ctx, path); err != nil {
		return nil, err
	}

	contentBytes := []byte(content)

	// Check for binary content
	if ctx.BinaryDetector.IsBinaryContent(contentBytes) {
		return nil, ErrBinaryFile
	}

	// Enforce size limit
	if int64(len(contentBytes)) > ctx.MaxFileSize {
		return nil, ErrTooLarge
	}

	// Determine permissions
	filePerm := os.FileMode(0644)
	if perm != nil {
		filePerm = *perm
	}

	// Write the file
	if err := ctx.FS.WriteFile(abs, contentBytes, filePerm); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Compute checksum and update cache
	checksum := ctx.ChecksumComputer.ComputeChecksum(contentBytes)
	ctx.ChecksumCache.Update(abs, checksum)

	return &WriteFileResponse{
		AbsolutePath: abs,
		RelativePath: rel,
		BytesWritten: len(contentBytes),
		FileMode:     uint32(filePerm),
	}, nil
}
