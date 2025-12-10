package models

import (
	"os"
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
)

func TestReadFileRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     ReadFileRequest
		wantErr bool
	}{
		{"Valid", ReadFileRequest{Path: "file.txt"}, false},
		{"EmptyPath", ReadFileRequest{Path: ""}, true},
		{"NegativeOffset", ReadFileRequest{Path: "file.txt", Offset: ptr(int64(-1))}, true},
		{"NegativeLimit", ReadFileRequest{Path: "file.txt", Limit: ptr(int64(-1))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteFileRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     WriteFileRequest
		wantErr bool
	}{
		{"Valid", WriteFileRequest{Path: "file.txt", Content: "content"}, false},
		{"EmptyPath", WriteFileRequest{Path: "", Content: "content"}, true},
		{"EmptyContent", WriteFileRequest{Path: "file.txt", Content: ""}, true},
		{"InvalidPerm", WriteFileRequest{Path: "file.txt", Content: "content", Perm: ptr(os.FileMode(07777))}, true}, // > 0777
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditFileRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     EditFileRequest
		wantErr bool
	}{
		{"Valid", EditFileRequest{Path: "file.txt", Operations: []Operation{{Before: "old", After: "new"}}}, false},
		{"EmptyPath", EditFileRequest{Path: "", Operations: []Operation{{Before: "old"}}}, true},
		{"EmptyOperations", EditFileRequest{Path: "file.txt", Operations: []Operation{}}, true},
		{"EmptyBefore", EditFileRequest{Path: "file.txt", Operations: []Operation{{Before: ""}}}, true},
		{"NegativeReplacements", EditFileRequest{Path: "file.txt", Operations: []Operation{{Before: "old", ExpectedReplacements: -1}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindFileRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     FindFileRequest
		wantErr bool
	}{
		{"Valid", FindFileRequest{Pattern: "*.txt"}, false},
		{"EmptyPattern", FindFileRequest{Pattern: ""}, true},
		{"PathTraversal", FindFileRequest{Pattern: "../outside"}, true}, // Pattern traversal check
		{"AbsolutePath", FindFileRequest{Pattern: "/etc/passwd"}, true},
		{"NegativeOffset", FindFileRequest{Pattern: "*.txt", Offset: -1}, true},
		{"NegativeLimit", FindFileRequest{Pattern: "*.txt", Limit: -1}, true},
		{"LimitExceedsMax", FindFileRequest{Pattern: "*.txt", Limit: cfg.Tools.MaxFindFileLimit + 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchContentRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     SearchContentRequest
		wantErr bool
	}{
		{"Valid", SearchContentRequest{Query: "foo"}, false},
		{"EmptyQuery", SearchContentRequest{Query: ""}, true},
		{"NegativeOffset", SearchContentRequest{Query: "foo", Offset: -1}, true},
		{"NegativeLimit", SearchContentRequest{Query: "foo", Limit: -1}, true},
		{"LimitExceedsMax", SearchContentRequest{Query: "foo", Limit: cfg.Tools.MaxSearchContentLimit + 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListDirectoryRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     ListDirectoryRequest
		wantErr bool
	}{
		{"Valid", ListDirectoryRequest{Path: "."}, false},
		{"EmptyPath", ListDirectoryRequest{Path: ""}, false},                  // path defaults to .
		{"NegativeOffset", ListDirectoryRequest{Path: ".", Offset: -1}, true}, // ListDirectoryRequest has Offset field? Yes.
		{"NegativeLimit", ListDirectoryRequest{Path: ".", Limit: -1}, true},
		{"LimitExceedsMax", ListDirectoryRequest{Path: ".", Limit: cfg.Tools.MaxListDirectoryLimit + 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShellRequest_Validate(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     ShellRequest
		wantErr bool
	}{
		{"Valid", ShellRequest{Command: []string{"echo", "hello"}}, false},
		{"EmptyCommand", ShellRequest{Command: []string{}}, true},
		{"NegativeTimeout", ShellRequest{Command: []string{"echo"}, TimeoutSeconds: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(cfg); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helpers
func ptr[T any](v T) *T {
	return &v
}

// Note: In tests we use explicit cast to os.FileMode if needed, but ptr helper works with any T.
// os.FileMode is uint32 on most systems but check if I can just use os.FileMode
// To solve imports, I need "os".
