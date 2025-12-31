package file

import (
	"context"
	"fmt"
	"os"

	"strings"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tool/helper/content"
)

// fileReader defines the minimal filesystem operations needed for reading files.
type fileReader interface {
	Stat(path string) (os.FileInfo, error)
	ReadFileRange(path string, offset, limit int64) ([]byte, error)
}

// checksumComputer defines the interface for checksum computation and updates.
type checksumComputer interface {
	Compute(data []byte) string
	Update(path string, checksum string)
}

// ReadFileTool handles file reading operations.
type ReadFileTool struct {
	fileOps         fileReader
	checksumManager checksumComputer
	config          *config.Config
	pathResolver    pathResolver
}

// NewReadFileTool creates a new ReadFileTool with injected dependencies.
func NewReadFileTool(
	fileOps fileReader,
	checksumManager checksumComputer,
	cfg *config.Config,
	pathResolver pathResolver,
) *ReadFileTool {
	if fileOps == nil {
		panic("fileOps is required")
	}
	if checksumManager == nil {
		panic("checksumManager is required")
	}
	if cfg == nil {
		panic("cfg is required")
	}
	if pathResolver == nil {
		panic("pathResolver is required")
	}
	return &ReadFileTool{
		fileOps:         fileOps,
		checksumManager: checksumManager,
		config:          cfg,
		pathResolver:    pathResolver,
	}
}

// Run reads a file from the workspace with optional offset and limit for partial reads.
// It validates the path is within workspace boundaries, checks for binary content,
// enforces size limits, and caches checksums for full file reads.
// Returns an error if the file is binary or outside the workspace. Large files are truncated.
//
// Note: ctx is accepted for API consistency but not used - file I/O is synchronous.
func (t *ReadFileTool) Run(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error) {
	if err := req.Validate(t.config); err != nil {
		return nil, err
	}

	abs, err := t.pathResolver.Abs(req.Path)
	if err != nil {
		return nil, err
	}
	rel, err := t.pathResolver.Rel(abs)
	if err != nil {
		return nil, err
	}

	// Get file info (single stat syscall)
	info, err := t.fileOps.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", abs, err)
	}

	// Check if it's a directory using info we already have
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", abs)
	}

	var offset int64
	if req.Offset != nil {
		offset = *req.Offset
	}

	limit := *req.Limit

	// Read the file range (single open+read syscall)
	contentBytes, err := t.fileOps.ReadFileRange(abs, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", abs, err)
	}

	if content.IsBinaryContent(contentBytes) {
		return nil, fmt.Errorf("file is binary: %s", abs)
	}

	var startLine int64 = 1
	if offset > 0 {
		// To get the correct line number, we must count newlines in the preceding content
		prefixBytes, err := t.fileOps.ReadFileRange(abs, 0, offset)
		if err != nil {
			// If we fail to read prefix, we can't determine line number accurately.
			// Fallback to 1 or return error?
			// Providing inaccurate line numbers (like 1) is better than failing the read, 
			// but we should probably note it. For now, let's propagate the error as it suggests FS issues.
			return nil, fmt.Errorf("failed to read file prefix for line counting: %w", err)
		}
		startLine = int64(strings.Count(string(prefixBytes), "\n")) + 1
	}

	content := string(contentBytes)

	// Only cache checksum if we read the entire file
	isFullRead := offset == 0 && int64(len(contentBytes)) == info.Size()

	if isFullRead {
		checksum := t.checksumManager.Compute(contentBytes)
		t.checksumManager.Update(abs, checksum)
	}

	truncated := (offset + int64(len(contentBytes))) < info.Size()
	return &ReadFileResponse{
		AbsolutePath: abs,
		RelativePath: rel,
		Size:         info.Size(),
		Content:      formatFileContent(content, startLine),
		Offset:       offset,
		Limit:        limit,
		Truncated:    truncated,
	}, nil
}

// formatFileContent wraps file content in <file> tags and adds line number prefixes.
func formatFileContent(text string, startLine int64) string {
	if text == "" {
		return "<file>\n(Empty file)\n</file>"
	}

	var sb strings.Builder
	sb.WriteString("<file>\n")

	lines := strings.Split(text, "\n")
	// If the file ends with a newline, strings.Split returns an empty string as the last element.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for i, line := range lines {
		lineNum := startLine + int64(i)
		sb.WriteString(fmt.Sprintf("%05d| %s\n", lineNum, line))
	}

	sb.WriteString("</file>")
	return sb.String()
}
