package tools

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/testing/mock"
	"github.com/Cyclone1070/iav/internal/tools/model"
	"github.com/Cyclone1070/iav/internal/tools/service"
)

func TestFindFile_BasicGlob(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		// Simulate fd output
		output := "/workspace/a/b/file.go\n/workspace/a/file.go\n"
		return mock.NewMockProcess(), strings.NewReader(output), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(resp.Matches))
	}

	// Verify sorting (alphabetical)
	expectedMatches := []string{"a/b/file.go", "a/file.go"}
	for i, expected := range expectedMatches {
		if resp.Matches[i] != expected {
			t.Errorf("match %d: expected %q, got %q", i, expected, resp.Matches[i])
		}
	}
}

func TestFindFile_Pagination(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	// Simulate 10 files
	var output string
	for i := range 10 {
		output += fmt.Sprintf("/workspace/file%d.txt\n", i)
	}

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		return mock.NewMockProcess(), strings.NewReader(output), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	// Request offset=2, limit=2
	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.txt", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(resp.Matches))
	}

	if resp.TotalCount != 10 {
		t.Errorf("expected TotalCount 10, got %d", resp.TotalCount)
	}

	if !resp.Truncated {
		t.Error("expected Truncated=true")
	}

	// Verify correct slice (file2.txt, file3.txt after sorting)
	if resp.Matches[0] != "file2.txt" {
		t.Errorf("expected file2.txt, got %s", resp.Matches[0])
	}
}

func TestFindFile_InvalidGlob(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		// Simulate fd error for invalid glob
		proc := mock.NewMockProcess()
		proc.WaitFunc = func() error {
			return mock.NewMockExitError(2)
		}
		return proc, strings.NewReader(""), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	_, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "[", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err == nil {
		t.Fatal("expected error for invalid glob, got nil")
	}
}

func TestFindFile_PathOutsideWorkspace(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mock.NewMockCommandExecutor(),
		Config:          *config.DefaultConfig(),
	}

	_, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "../outside", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != model.ErrOutsideWorkspace {
		t.Errorf("expected ErrOutsideWorkspace, got %v", err)
	}
}

func TestFindFile_NonExistentPath(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mock.NewMockCommandExecutor(),
		Config:          *config.DefaultConfig(),
	}

	_, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "nonexistent/dir", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != model.ErrFileMissing {
		t.Errorf("expected ErrFileMissing, got %v", err)
	}
}

func TestFindFile_CommandFailure(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		proc := mock.NewMockProcess()
		proc.WaitFunc = func() error {
			return mock.NewMockExitError(2)
		}
		return proc, strings.NewReader(""), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	_, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err == nil {
		t.Fatal("expected error for command failure, got nil")
	}
}

func TestFindFile_ShellInjection(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	var capturedCmd []string
	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		capturedCmd = cmd
		return mock.NewMockProcess(), strings.NewReader(""), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	pattern := "*.go; rm -rf /"
	_, _ = FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: pattern, SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})

	// Verify pattern is passed as literal argument, not shell-interpreted
	found := slices.Contains(capturedCmd, pattern)

	if !found {
		t.Errorf("expected pattern to be passed as literal argument, got cmd: %v", capturedCmd)
	}
}

func TestFindFile_UnicodeFilenames(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		output := "/workspace/🚀.txt\n/workspace/文件.txt\n"
		return mock.NewMockProcess(), strings.NewReader(output), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.txt", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(resp.Matches))
	}

	// Verify unicode handling
	foundEmoji := false
	foundChinese := false
	for _, match := range resp.Matches {
		if match == "🚀.txt" {
			foundEmoji = true
		}
		if match == "文件.txt" {
			foundChinese = true
		}
	}

	if !foundEmoji {
		t.Error("expected to find emoji filename")
	}
	if !foundChinese {
		t.Error("expected to find Chinese filename")
	}
}

func TestFindFile_DeeplyNested(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	// Simulate path with 100 segments
	deepPath := "/workspace"
	for i := range 100 {
		deepPath += fmt.Sprintf("/dir%d", i)
	}
	deepPath += "/file.txt"

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		return mock.NewMockProcess(), strings.NewReader(deepPath + "\n"), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.txt", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(resp.Matches))
	}
}

func TestFindFile_NoMatches(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		// Simulate fd returning exit code 0 (no matches, empty output)
		return mock.NewMockProcess(), strings.NewReader(""), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.nonexistent", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(resp.Matches))
	}

	if resp.Truncated {
		t.Error("expected Truncated=false for no matches")
	}
}
func TestFindFile_IncludeIgnored(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	// Test with includeIgnored=false (default behavior, should respect gitignore)
	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		// Verify --no-ignore is NOT present
		if slices.Contains(cmd, "--no-ignore") {
			t.Error("expected --no-ignore to NOT be present when includeIgnored=false")
		}
		// Simulate fd output without ignored files
		output := "/workspace/visible.go\n"
		return mock.NewMockProcess(), strings.NewReader(output), strings.NewReader(""), nil
	}

	ctx := &model.WorkspaceContext{
		FS:              fs,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: service.NewChecksumManager(),
		WorkspaceRoot:   workspaceRoot,
		CommandExecutor: mockRunner,
		Config:          *config.DefaultConfig(),
	}

	resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "", MaxDepth: 0, IncludeIgnored: false, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(resp.Matches))
	}

	// Test with includeIgnored=true (should include ignored files)
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		// Verify --no-ignore IS present
		if !slices.Contains(cmd, "--no-ignore") {
			t.Error("expected --no-ignore to be present when includeIgnored=true")
		}
		// Simulate fd output with ignored files
		output := "/workspace/ignored.go\n/workspace/visible.go\n"
		return mock.NewMockProcess(), strings.NewReader(output), strings.NewReader(""), nil
	}

	resp, err = FindFile(context.Background(), ctx, model.FindFileRequest{Pattern: "*.go", SearchPath: "", MaxDepth: 0, IncludeIgnored: true, Offset: 0, Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(resp.Matches))
	}

	// Verify both files are present
	foundIgnored := false
	foundVisible := false
	for _, match := range resp.Matches {
		if match == "ignored.go" {
			foundIgnored = true
		}
		if match == "visible.go" {
			foundVisible = true
		}
	}

	if !foundIgnored {
		t.Error("expected to find ignored.go when includeIgnored=true")
	}
	if !foundVisible {
		t.Error("expected to find visible.go when includeIgnored=true")
	}
}

func TestFindFile_LimitValidation(t *testing.T) {
	workspaceRoot := "/workspace"

	fs := mock.NewMockFileSystem()
	fs.CreateDir("/workspace")

	mockRunner := mock.NewMockCommandExecutor()
	mockRunner.StartFunc = func(ctx context.Context, cmd []string, opts model.ProcessOptions) (model.Process, io.Reader, io.Reader, error) {
		return mock.NewMockProcess(), strings.NewReader(""), strings.NewReader(""), nil
	}

	t.Run("zero limit uses default", func(t *testing.T) {
		ctx := &model.WorkspaceContext{
			FS:              fs,
			WorkspaceRoot:   workspaceRoot,
			CommandExecutor: mockRunner,
			Config:          *config.DefaultConfig(),
		}

		resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{
			Pattern: "*.go",
			Limit:   0, // Should use DefaultFindFileLimit
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Limit != ctx.Config.Tools.DefaultFindFileLimit {
			t.Errorf("expected default limit %d, got %d", ctx.Config.Tools.DefaultFindFileLimit, resp.Limit)
		}
	})

	t.Run("custom config limits are respected", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Tools.DefaultFindFileLimit = 25
		cfg.Tools.MaxFindFileLimit = 50

		ctx := &model.WorkspaceContext{
			FS:              fs,
			WorkspaceRoot:   workspaceRoot,
			CommandExecutor: mockRunner,
			Config:          *cfg,
		}

		resp, err := FindFile(context.Background(), ctx, model.FindFileRequest{
			Pattern: "*.go",
			Limit:   30,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Limit != 30 {
			t.Errorf("expected limit 30, got %d", resp.Limit)
		}
	})
}
