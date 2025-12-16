package file

import "errors"

// FileMissingError indicates a file does not exist.
type FileMissingError struct {
	Path string
}

func (e *FileMissingError) Error() string {
	if e.Path != "" {
		return "file does not exist: " + e.Path
	}
	return "file does not exist"
}

func (e *FileMissingError) FileMissing() bool {
	return true
}

// File operation errors
var (
	ErrFileExists                   = errors.New("file already exists, use EditFile instead")
	ErrBinaryFile                   = errors.New("binary files are not supported")
	ErrEditConflict                 = errors.New("file was modified since last read, please re-read first")
	ErrSnippetNotFound              = errors.New("snippet not found in file")
	ErrExpectedReplacementsMismatch = errors.New("expected replacements count does not match actual occurrences")
	ErrTooLarge                     = errors.New("file or content exceeds size limit")
	ErrFileMissing                  = &FileMissingError{}
)
