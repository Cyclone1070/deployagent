package file

// binaryDetector defines the interface for binary content detection.
// This interface is shared by all file tools (read, write, edit).
type binaryDetector interface {
	IsBinaryContent(content []byte) bool
}
