package file

import (
	"os"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tool/pathutil"
)

// -- Read File --

type ReadFileDTO struct {
	Path   string `json:"path"`
	Offset *int64 `json:"offset,omitempty"`
	Limit  *int64 `json:"limit,omitempty"`
}

type ReadFileRequest struct {
	absPath string
	relPath string
	offset  int64
	limit   int64
}

func (r ReadFileRequest) AbsPath() string { return r.absPath }
func (r ReadFileRequest) RelPath() string { return r.relPath }
func (r ReadFileRequest) Offset() int64   { return r.offset }
func (r ReadFileRequest) Limit() int64    { return r.limit }

func NewReadFileRequest(
	dto ReadFileDTO,
	cfg *config.Config,
	workspaceRoot string,
	fs pathutil.FileSystem,
) (*ReadFileRequest, error) {
	if dto.Path == "" {
		return nil, ErrPathRequired
	}

	abs, rel, err := pathutil.Resolve(workspaceRoot, fs, dto.Path)
	if err != nil {
		return nil, err
	}

	var offset int64
	if dto.Offset != nil {
		offset = *dto.Offset
		if offset < 0 {
			return nil, ErrInvalidOffset
		}
	}

	var limit int64
	if dto.Limit != nil {
		limit = *dto.Limit
		if limit < 0 {
			return nil, ErrInvalidLimit
		}
	}
	if limit == 0 {
		limit = cfg.Tools.MaxFileSize
	}

	return &ReadFileRequest{
		absPath: abs,
		relPath: rel,
		offset:  offset,
		limit:   limit,
	}, nil
}

type ReadFileResponse struct {
	Content      string
	AbsolutePath string
	RelativePath string
	Size         int64
}

// -- Write File --

type WriteFileDTO struct {
	Path    string       `json:"path"`
	Content string       `json:"content"`
	Perm    *os.FileMode `json:"perm,omitempty"`
}

type WriteFileRequest struct {
	absPath string
	relPath string
	content string
	perm    os.FileMode
}

func (r WriteFileRequest) AbsPath() string { return r.absPath }
func (r WriteFileRequest) RelPath() string { return r.relPath }
func (r WriteFileRequest) Content() string { return r.content }
func (r WriteFileRequest) Perm() os.FileMode {
	if r.perm == 0 {
		return 0644
	}
	return r.perm
}

func NewWriteFileRequest(
	dto WriteFileDTO,
	cfg *config.Config,
	workspaceRoot string,
	fs pathutil.FileSystem,
) (*WriteFileRequest, error) {
	if dto.Path == "" {
		return nil, ErrPathRequired
	}

	if dto.Content == "" {
		return nil, ErrContentRequiredForWrite
	}

	abs, rel, err := pathutil.Resolve(workspaceRoot, fs, dto.Path)
	if err != nil {
		return nil, err
	}

	perm := os.FileMode(0644)
	if dto.Perm != nil {
		perm = *dto.Perm
		if perm > 0777 {
			return nil, ErrInvalidPermissions
		}
	}

	return &WriteFileRequest{
		absPath: abs,
		relPath: rel,
		content: dto.Content,
		perm:    perm,
	}, nil
}

type WriteFileResponse struct {
	AbsolutePath string
	RelativePath string
	BytesWritten int
	FileMode     uint32
}

// -- Edit File --

type OperationDTO struct {
	Before               string `json:"before"`
	After                string `json:"after"`
	ExpectedReplacements int    `json:"expected_replacements,omitempty"`
}

type EditFileDTO struct {
	Path       string         `json:"path"`
	Operations []OperationDTO `json:"operations"`
}

type EditFileRequest struct {
	absPath    string
	relPath    string
	operations []Operation
}

type Operation struct {
	before               string
	after                string
	expectedReplacements int
}

func (o Operation) Before() string                { return o.before }
func (o Operation) After() string                 { return o.after }
func (o Operation) ExpectedReplacements() int     { return o.expectedReplacements }
func (r EditFileRequest) AbsPath() string         { return r.absPath }
func (r EditFileRequest) RelPath() string         { return r.relPath }
func (r EditFileRequest) Operations() []Operation { return r.operations }

func NewEditFileRequest(
	dto EditFileDTO,
	cfg *config.Config,
	workspaceRoot string,
	fs pathutil.FileSystem,
) (*EditFileRequest, error) {
	if dto.Path == "" {
		return nil, ErrPathRequired
	}

	if len(dto.Operations) == 0 {
		return nil, ErrOperationsRequired
	}

	abs, rel, err := pathutil.Resolve(workspaceRoot, fs, dto.Path)
	if err != nil {
		return nil, err
	}

	var operations []Operation
	for _, op := range dto.Operations {
		if op.Before == "" {
			return nil, ErrEmptyBefore
		}

		expected := op.ExpectedReplacements
		if expected < 0 {
			return nil, ErrInvalidExpectedReplacements
		}
		operations = append(operations, Operation{
			before:               op.Before,
			after:                op.After,
			expectedReplacements: expected,
		})
	}

	return &EditFileRequest{
		absPath:    abs,
		relPath:    rel,
		operations: operations,
	}, nil
}

type EditFileResponse struct {
	AbsolutePath      string
	RelativePath      string
	OperationsApplied int
	FileSize          int64
}
