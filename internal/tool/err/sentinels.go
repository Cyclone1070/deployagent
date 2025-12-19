package err

import "errors"

// Common Filesystem Errors
var (
	ErrFileMissing         = errors.New("file or path does not exist")
	ErrFileExists          = errors.New("file already exists")
	ErrNotADirectory       = errors.New("path is not a directory")
	ErrIsDirectory         = errors.New("path is a directory")
	ErrOutsideWorkspace    = errors.New("path is outside workspace root")
	ErrWorkspaceRootNotSet = errors.New("workspace root not set")
	ErrPathTraversal       = errors.New("path traversal detected")
	ErrBinaryFile          = errors.New("file is binary")
	ErrFileTooLarge        = errors.New("file too large")
)

// Common Validation Errors
var (
	ErrPathRequired        = errors.New("path is required")
	ErrInvalidOffset       = errors.New("offset cannot be negative")
	ErrInvalidLimit        = errors.New("limit cannot be negative")
	ErrLimitExceeded       = errors.New("limit exceeds maximum")
	ErrInvalidPermission   = errors.New("invalid permissions")
	ErrSymlinkLoop         = errors.New("symlink loop detected")
	ErrSymlinkChainTooLong = errors.New("symlink chain too long")
)

// Feature: File Edit/Content
var (
	ErrContentRequired             = errors.New("content is required")
	ErrOperationsRequired          = errors.New("operations cannot be empty")
	ErrBeforeRequired              = errors.New("operation before field required")
	ErrSnippetNotFound             = errors.New("snippet not found")
	ErrReplacementMismatch         = errors.New("replacement mismatch")
	ErrInvalidExpectedReplacements = errors.New("invalid expected replacements")
	ErrEditConflict                = errors.New("edit conflict")
)

// Feature: Directory/Search
var (
	ErrPatternRequired = errors.New("pattern is required")
	ErrInvalidPattern  = errors.New("invalid glob pattern")
	ErrQueryRequired   = errors.New("query is required")
)

// Feature: Shell
var (
	ErrCommandRequired = errors.New("command cannot be empty")
	ErrInvalidTimeout  = errors.New("timeout_seconds cannot be negative")
	ErrEnvFileParse    = errors.New("invalid line in env file")
	ErrEnvFileScan     = errors.New("error reading env file")
)

// Feature: Todo
var (
	ErrInvalidStatus      = errors.New("invalid status")
	ErrEmptyDescription   = errors.New("description cannot be empty")
	ErrStoreNotConfigured = errors.New("todo store not configured")
)
