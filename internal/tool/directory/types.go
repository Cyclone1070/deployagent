package directory

import (
	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/tool/pathutil"
)

// -- Directory Tool Contract Types --

// DirectoryEntry represents a single entry in a directory listing
type DirectoryEntry struct {
	relativePath string
	isDir        bool
}

// NewDirectoryEntry creates a new DirectoryEntry.
func NewDirectoryEntry(relativePath string, isDir bool) DirectoryEntry {
	return DirectoryEntry{relativePath: relativePath, isDir: isDir}
}

// RelativePath returns the relative path of the entry.
func (e DirectoryEntry) RelativePath() string {
	return e.relativePath
}

// IsDir returns whether the entry is a directory.
func (e DirectoryEntry) IsDir() bool {
	return e.isDir
}

// ListDirectoryDTO is the wire format for ListDirectory operation
type ListDirectoryDTO struct {
	Path           string `json:"path"`
	MaxDepth       int    `json:"max_depth,omitempty"`
	IncludeIgnored bool   `json:"include_ignored,omitempty"`
	Offset         int    `json:"offset,omitempty"`
	Limit          int    `json:"limit,omitempty"`
}

// ListDirectoryRequest is the validated domain entity for ListDirectory operation
type ListDirectoryRequest struct {
	absPath        string
	relPath        string
	maxDepth       int
	includeIgnored bool
	offset         int
	limit          int
}

// NewListDirectoryRequest creates and validates a new ListDirectoryRequest.
func NewListDirectoryRequest(
	dto ListDirectoryDTO,
	cfg *config.Config,
	workspaceRoot string,
	fs pathutil.FileSystem,
) (*ListDirectoryRequest, error) {
	if dto.Path == "" {
		return nil, ErrPathRequired
	}

	absPath, relPath, err := pathutil.Resolve(workspaceRoot, fs, dto.Path)
	if err != nil {
		return nil, err
	}

	offset := dto.Offset
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	limit := dto.Limit
	if limit < 0 {
		return nil, ErrInvalidLimit
	}
	if limit == 0 {
		limit = cfg.Tools.DefaultListDirectoryLimit
	}
	if limit > cfg.Tools.MaxListDirectoryLimit {
		return nil, ErrLimitExceeded
	}

	return &ListDirectoryRequest{
		absPath:        absPath,
		relPath:        relPath,
		maxDepth:       dto.MaxDepth,
		includeIgnored: dto.IncludeIgnored,
		offset:         offset,
		limit:          limit,
	}, nil
}

// AbsPath returns the absolute path
func (r *ListDirectoryRequest) AbsPath() string {
	return r.absPath
}

// RelPath returns the relative path
func (r *ListDirectoryRequest) RelPath() string {
	return r.relPath
}

// MaxDepth returns the max depth
func (r *ListDirectoryRequest) MaxDepth() int {
	return r.maxDepth
}

// IncludeIgnored returns whether to include ignored files
func (r *ListDirectoryRequest) IncludeIgnored() bool {
	return r.includeIgnored
}

// Offset returns the offset
func (r *ListDirectoryRequest) Offset() int {
	return r.offset
}

// Limit returns the limit
func (r *ListDirectoryRequest) Limit() int {
	return r.limit
}

// ListDirectoryResponse contains the result of a ListDirectory operation
type ListDirectoryResponse struct {
	DirectoryPath    string
	Entries          []DirectoryEntry
	Offset           int
	Limit            int
	TotalCount       int
	Truncated        bool
	TruncationReason string
}

// FindFileDTO is the wire format for FindFile operation
type FindFileDTO struct {
	Pattern        string `json:"pattern"`
	SearchPath     string `json:"search_path"`
	MaxDepth       int    `json:"max_depth,omitempty"`
	IncludeIgnored bool   `json:"include_ignored,omitempty"`
	Offset         int    `json:"offset,omitempty"`
	Limit          int    `json:"limit,omitempty"`
}

// FindFileRequest is the validated domain entity for FindFile operation
type FindFileRequest struct {
	pattern        string
	searchAbsPath  string
	searchRelPath  string
	maxDepth       int
	includeIgnored bool
	offset         int
	limit          int
}

// NewFindFileRequest creates and validates a new FindFileRequest.
func NewFindFileRequest(
	dto FindFileDTO,
	cfg *config.Config,
	workspaceRoot string,
	fs pathutil.FileSystem,
) (*FindFileRequest, error) {
	if dto.Pattern == "" {
		return nil, ErrPatternRequired
	}

	searchPath := dto.SearchPath
	if searchPath == "" {
		searchPath = "."
	}

	absPath, relPath, err := pathutil.Resolve(workspaceRoot, fs, searchPath)
	if err != nil {
		return nil, err
	}

	offset := dto.Offset
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	limit := dto.Limit
	if limit < 0 {
		return nil, ErrInvalidLimit
	}
	if limit == 0 {
		limit = cfg.Tools.DefaultFindFileLimit
	}
	if limit > cfg.Tools.MaxFindFileLimit {
		return nil, ErrLimitExceeded
	}

	return &FindFileRequest{
		pattern:        dto.Pattern,
		searchAbsPath:  absPath,
		searchRelPath:  relPath,
		maxDepth:       dto.MaxDepth,
		includeIgnored: dto.IncludeIgnored,
		offset:         offset,
		limit:          limit,
	}, nil
}

// Pattern returns the search pattern
func (r *FindFileRequest) Pattern() string {
	return r.pattern
}

// SearchAbsPath returns the absolute search path
func (r *FindFileRequest) SearchAbsPath() string {
	return r.searchAbsPath
}

// SearchRelPath returns the relative search path
func (r *FindFileRequest) SearchRelPath() string {
	return r.searchRelPath
}

// MaxDepth returns the max depth
func (r *FindFileRequest) MaxDepth() int {
	return r.maxDepth
}

// IncludeIgnored returns whether to include ignored files
func (r *FindFileRequest) IncludeIgnored() bool {
	return r.includeIgnored
}

// Offset returns the offset
func (r *FindFileRequest) Offset() int {
	return r.offset
}

// Limit returns the limit
func (r *FindFileRequest) Limit() int {
	return r.limit
}

// FindFileResponse contains the result of a FindFile operation
type FindFileResponse struct {
	Matches    []string
	Offset     int
	Limit      int
	TotalCount int
	Truncated  bool
}
