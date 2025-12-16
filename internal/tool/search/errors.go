package search

// FileMissingError implements the behavioral interface for missing files.
type FileMissingError struct {
	Path string
}

func (e *FileMissingError) Error() string {
	return "search path does not exist: " + e.Path
}

func (e *FileMissingError) FileMissing() bool {
	return true
}

// NotDirectoryError implements the behavioral interface for non-directory paths.
type NotDirectoryError struct {
	Path string
}

func (e *NotDirectoryError) Error() string {
	return "search path is not a directory: " + e.Path
}

func (e *NotDirectoryError) NotDirectory() bool {
	return true
}
