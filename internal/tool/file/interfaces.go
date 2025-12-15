package file

import "os"

// pathResolver defines the filesystem operations needed for path resolution.
// This interface is shared by all file tools (read, write, edit).
type pathResolver interface {
	Lstat(path string) (os.FileInfo, error)
	Readlink(path string) (string, error)
	UserHomeDir() (string, error)
}

// binaryDetector defines the interface for binary content detection.
// This interface is shared by all file tools (read, write, edit).
type binaryDetector interface {
	IsBinaryContent(content []byte) bool
}
