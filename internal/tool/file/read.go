package file

import (
	"context"
	"fmt"
	"strings"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tool"
	"github.com/Cyclone1070/iav/internal/tool/helper/content"
	"github.com/Cyclone1070/iav/internal/tool/helper/pagination"
	"github.com/Cyclone1070/iav/internal/workflow/toolmanager"
)

// fileReader defines the minimal filesystem operations needed for reading files.
type fileReader interface {
	ReadFile(path string) ([]byte, error)
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
	pathResolver    pathResolver
	config          *config.Config
}

// NewReadFileTool creates a new ReadFileTool with injected dependencies.
func NewReadFileTool(
	fileOps fileReader,
	checksumManager checksumComputer,
	pathResolver pathResolver,
	cfg *config.Config,
) *ReadFileTool {
	if fileOps == nil {
		panic("fileOps is required")
	}
	if checksumManager == nil {
		panic("checksumManager is required")
	}
	if pathResolver == nil {
		panic("pathResolver is required")
	}
	if cfg == nil {
		panic("config is required")
	}
	return &ReadFileTool{
		fileOps:         fileOps,
		checksumManager: checksumManager,
		pathResolver:    pathResolver,
		config:          cfg,
	}
}

// Name returns the tool's identifier.
func (t *ReadFileTool) Name() string {
	return "read_file"
}

// Declaration returns the tool's schema for the LLM.
func (t *ReadFileTool) Declaration() tool.Declaration {
	return tool.Declaration{
		Name:        "read_file",
		Description: "Read file contents with optional pagination. Use offset/limit to read large files in chunks.",
		Parameters: &tool.Schema{
			Type: tool.TypeObject,
			Properties: map[string]*tool.Schema{
				"path":   {Type: tool.TypeString, Description: "Path to file"},
				"offset": {Type: tool.TypeInteger, Description: "Start line index (0-indexed)"},
				"limit":  {Type: tool.TypeInteger, Description: "Max lines to return"},
			},
			Required: []string{"path"},
		},
	}
}

// Request returns a new request struct for JSON unmarshalling.
func (t *ReadFileTool) Request() toolmanager.ToolRequest {
	return &ReadFileRequest{}
}

// Execute runs the tool with the request and returns a ToolResult.
func (t *ReadFileTool) Execute(ctx context.Context, req toolmanager.ToolRequest) (toolmanager.ToolResult, error) {
	r, ok := req.(*ReadFileRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: %T", req)
	}

	if err := r.Validate(t.config); err != nil {
		return &ReadFileResponse{Error: err.Error()}, nil
	}

	resp, err := t.Run(ctx, r)
	if err != nil {
		// PER CONTRACT: Infra/Context errors return error, stopping the loop.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Tool failures (file not found, validation, etc) return response with Error field set
		return &ReadFileResponse{Error: err.Error()}, nil
	}

	return resp, nil
}

// Run reads a file from the workspace with line-based pagination.
func (t *ReadFileTool) Run(ctx context.Context, req *ReadFileRequest) (*ReadFileResponse, error) {
	abs, err := t.pathResolver.Abs(req.Path)
	if err != nil {
		return nil, err
	}

	// Read full file content
	data, err := t.fileOps.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	// Always update checksum since we read the full file
	checksum := t.checksumManager.Compute(data)
	t.checksumManager.Update(abs, checksum)

	// Split into lines using content.SplitLines which handles both \n and \r\n
	lines := content.SplitLines(string(data))

	// Apply pagination
	paginatedLines, pagRes := pagination.ApplyPagination(lines, req.Offset, req.Limit)

	// Calculate display lines
	startLine := req.Offset + 1
	endLine := startLine + len(paginatedLines) - 1
	if len(paginatedLines) == 0 {
		endLine = startLine - 1
	}

	return &ReadFileResponse{
		Content:    strings.Join(paginatedLines, "\n"),
		StartLine:  startLine,
		EndLine:    endLine,
		TotalLines: pagRes.TotalCount,
	}, nil
}
