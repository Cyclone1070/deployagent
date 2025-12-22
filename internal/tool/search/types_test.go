package search

import (
	"errors"
	"os"
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
)

// Minimal mock for validation tests
type mockFSForTypes struct {
	dirs map[string]bool
}

func (m *mockFSForTypes) Lstat(path string) (os.FileInfo, error) {
	if m.dirs[path] {
		return &mockFileInfoForTypes{isDir: true}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFSForTypes) Readlink(path string) (string, error) {
	return "", os.ErrInvalid
}

func (m *mockFSForTypes) UserHomeDir() (string, error) {
	return "/home/user", nil
}

type mockFileInfoForTypes struct {
	os.FileInfo
	isDir bool
}

func (m *mockFileInfoForTypes) IsDir() bool { return m.isDir }

func TestSearchContentRequest_Validation(t *testing.T) {
	cfg := config.DefaultConfig()
	fs := &mockFSForTypes{dirs: map[string]bool{"/workspace": true}}
	workspaceRoot := "/workspace"

	tests := []struct {
		name    string
		dto     SearchContentDTO
		wantErr error
	}{
		{"Valid", SearchContentDTO{Query: "foo"}, nil},
		{"EmptyQuery", SearchContentDTO{Query: ""}, ErrQueryRequired},
		{"NegativeOffset", SearchContentDTO{Query: "foo", Offset: -1}, ErrInvalidOffset},
		{"NegativeLimit", SearchContentDTO{Query: "foo", Limit: -1}, ErrInvalidLimit},
		{"LimitExceedsMax", SearchContentDTO{Query: "foo", Limit: cfg.Tools.MaxSearchContentLimit + 1}, ErrLimitExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSearchContentRequest(tt.dto, cfg, workspaceRoot, fs)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("NewSearchContentRequest() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("NewSearchContentRequest() error = nil, want %v", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewSearchContentRequest() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}
