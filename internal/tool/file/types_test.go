package file

import (
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
)

func TestReadFileRequest_Validation(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name         string
		req          ReadFileRequest
		wantErr      bool
		verifyValues func(t *testing.T, req ReadFileRequest)
	}{
		{"Valid", ReadFileRequest{Path: "file.txt"}, false, nil},
		{"EmptyPath", ReadFileRequest{Path: ""}, true, nil},
		{"NegativeOffset_Clamps", ReadFileRequest{Path: "file.txt", Offset: -1}, false, func(t *testing.T, req ReadFileRequest) {
			if req.Offset != 0 {
				t.Errorf("expected Offset 0, got %d", req.Offset)
			}
		}},
		{"ZeroLimit_Defaults", ReadFileRequest{Path: "file.txt", Limit: 0}, false, func(t *testing.T, req ReadFileRequest) {
			if req.Limit != cfg.Tools.DefaultReadFileLimit {
				t.Errorf("expected Limit %d, got %d", cfg.Tools.DefaultReadFileLimit, req.Limit)
			}
		}},
		{"HighLimit_Accepted", ReadFileRequest{Path: "file.txt", Limit: 100000}, false, func(t *testing.T, req ReadFileRequest) {
			if req.Limit != 100000 {
				t.Errorf("expected Limit 100000, got %d", req.Limit)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.verifyValues != nil {
				tt.verifyValues(t, tt.req)
			}
		})
	}
}

func TestWriteFileRequest_Validation(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name    string
		req     WriteFileRequest
		wantErr bool
	}{
		{"Valid", WriteFileRequest{Path: "file.txt", Content: "content"}, false},
		{"EmptyPath", WriteFileRequest{Path: "", Content: "content"}, true},
		{"EmptyContent", WriteFileRequest{Path: "file.txt", Content: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditFileRequest_Validation(t *testing.T) {
	tests := []struct {
		name         string
		req          EditFileRequest
		wantErr      bool
		verifyValues func(t *testing.T, req EditFileRequest)
	}{
		{"Valid", EditFileRequest{Path: "file.txt", Operations: []EditOperation{{Before: "old", After: "new"}}}, false, nil},
		{"EmptyPath", EditFileRequest{Path: "", Operations: []EditOperation{{Before: "old"}}}, true, nil},
		{"EmptyOperations", EditFileRequest{Path: "file.txt", Operations: []EditOperation{}}, true, nil},
		{"EmptyBeforeIsAppend", EditFileRequest{Path: "file.txt", Operations: []EditOperation{{Before: ""}}}, false, nil},
		{"NegativeReplacements_Clamps", EditFileRequest{Path: "file.txt", Operations: []EditOperation{{Before: "old", ExpectedReplacements: -1}}}, false, func(t *testing.T, req EditFileRequest) {
			if req.Operations[0].ExpectedReplacements != 1 {
				t.Errorf("expected ExpectedReplacements 1, got %d", req.Operations[0].ExpectedReplacements)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.verifyValues != nil {
				tt.verifyValues(t, tt.req)
			}
		})
	}
}
