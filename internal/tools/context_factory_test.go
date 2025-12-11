package tools

import (
	"context"
	"testing"

	"github.com/Cyclone1070/iav/internal/config"
	"github.com/Cyclone1070/iav/internal/testing/mock"
	"github.com/Cyclone1070/iav/internal/tools/model"
	"github.com/Cyclone1070/iav/internal/tools/service"
)

func TestMultiContextIsolation(t *testing.T) {

	// Create two separate contexts with different workspace roots
	fs1 := mock.NewMockFileSystem()
	checksumManager1 := service.NewChecksumManager()

	fs2 := mock.NewMockFileSystem()
	checksumManager2 := service.NewChecksumManager()

	ctx1 := &model.WorkspaceContext{
		FS:              fs1,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: checksumManager1,
		WorkspaceRoot:   "/workspace1",
		Config:          *config.DefaultConfig(),
	}

	ctx2 := &model.WorkspaceContext{
		FS:              fs2,
		BinaryDetector:  mock.NewMockBinaryDetector(),
		ChecksumManager: checksumManager2,
		WorkspaceRoot:   "/workspace2",
		Config:          *config.DefaultConfig(),
	}

	// Create files in both contexts
	content1 := "content1"
	content2 := "content2"

	resp1, err := WriteFile(context.Background(), ctx1, model.WriteFileRequest{Path: "file.txt", Content: content1})
	if err != nil {
		t.Fatalf("failed to write file in ctx1: %v", err)
	}

	resp2, err := WriteFile(context.Background(), ctx2, model.WriteFileRequest{Path: "file.txt", Content: content2})
	if err != nil {
		t.Fatalf("failed to write file in ctx2: %v", err)
	}

	// Verify caches are isolated
	checksum1, ok1 := ctx1.ChecksumManager.Get(resp1.AbsolutePath)
	if !ok1 {
		t.Error("ctx1 cache should contain file1")
	}
	if checksum1 == "" {
		t.Error("ctx1 cache checksum should not be empty")
	}

	checksum2, ok2 := ctx2.ChecksumManager.Get(resp2.AbsolutePath)
	if !ok2 {
		t.Error("ctx2 cache should contain file2")
	}
	if checksum2 == "" {
		t.Error("ctx2 cache checksum should not be empty")
	}

	// Verify ctx1 cache doesn't contain ctx2's file
	_, ok := ctx1.ChecksumManager.Get(resp2.AbsolutePath)
	if ok {
		t.Error("ctx1 cache should not contain ctx2's file")
	}

	// Verify ctx2 cache doesn't contain ctx1's file
	_, ok = ctx2.ChecksumManager.Get(resp1.AbsolutePath)
	if ok {
		t.Error("ctx2 cache should not contain ctx1's file")
	}

	// Verify filesystems are isolated
	read1, err := ReadFile(context.Background(), ctx1, model.ReadFileRequest{Path: "file.txt"})
	if err != nil {
		t.Fatalf("failed to read file from ctx1: %v", err)
	}
	if read1.Content != content1 {
		t.Errorf("ctx1 should read its own content, got %q", read1.Content)
	}

	read2, err := ReadFile(context.Background(), ctx2, model.ReadFileRequest{Path: "file.txt"})
	if err != nil {
		t.Fatalf("failed to read file from ctx2: %v", err)
	}
	if read2.Content != content2 {
		t.Errorf("ctx2 should read its own content, got %q", read2.Content)
	}
}

func TestCustomFileSizeLimit(t *testing.T) {
	workspaceRoot := "/workspace"
	smallLimit := int64(100)              // 100 bytes
	largeLimit := int64(10 * 1024 * 1024) // 10MB

	t.Run("small limit enforced", func(t *testing.T) {
		fs := mock.NewMockFileSystem()
		checksumManager := service.NewChecksumManager()

		cfg := config.DefaultConfig()
		cfg.Tools.MaxFileSize = smallLimit
		ctx := &model.WorkspaceContext{
			FS:              fs,
			BinaryDetector:  mock.NewMockBinaryDetector(),
			ChecksumManager: checksumManager,
			WorkspaceRoot:   workspaceRoot,
			Config:          *cfg,
		}

		// Create content that exceeds the limit
		largeContent := make([]byte, smallLimit+1)
		for i := range largeContent {
			largeContent[i] = 'A'
		}

		_, err := WriteFile(context.Background(), ctx, model.WriteFileRequest{Path: "large.txt", Content: string(largeContent)})
		if err != model.ErrTooLarge {
			t.Errorf("expected ErrTooLarge for content exceeding limit, got %v", err)
		}
	})

	t.Run("large limit allows bigger files", func(t *testing.T) {
		fs := mock.NewMockFileSystem()
		checksumManager := service.NewChecksumManager()

		cfg := config.DefaultConfig()
		cfg.Tools.MaxFileSize = largeLimit
		ctx := &model.WorkspaceContext{
			FS:              fs,
			BinaryDetector:  mock.NewMockBinaryDetector(),
			ChecksumManager: checksumManager,
			WorkspaceRoot:   workspaceRoot,
			Config:          *cfg,
		}

		// Create content within the large limit but exceeding default
		content := make([]byte, 6*1024*1024) // 6MB, exceeds default 5MB
		for i := range content {
			content[i] = 'A'
		}

		_, err := WriteFile(context.Background(), ctx, model.WriteFileRequest{Path: "large.txt", Content: string(content)})
		if err != nil {
			t.Errorf("expected success with large limit, got %v", err)
		}
	})

	t.Run("different limits in different contexts", func(t *testing.T) {
		fs1 := mock.NewMockFileSystem()
		checksumManager1 := service.NewChecksumManager()

		fs2 := mock.NewMockFileSystem()
		checksumManager2 := service.NewChecksumManager()

		cfg1 := config.DefaultConfig()
		cfg1.Tools.MaxFileSize = smallLimit
		ctx1 := &model.WorkspaceContext{
			FS:              fs1,
			BinaryDetector:  mock.NewMockBinaryDetector(),
			ChecksumManager: checksumManager1,
			WorkspaceRoot:   workspaceRoot,
			Config:          *cfg1,
		}

		cfg2 := config.DefaultConfig()
		cfg2.Tools.MaxFileSize = largeLimit
		ctx2 := &model.WorkspaceContext{
			FS:              fs2,
			BinaryDetector:  mock.NewMockBinaryDetector(),
			ChecksumManager: checksumManager2,
			WorkspaceRoot:   workspaceRoot,
			Config:          *cfg2,
		}

		// Content that fits in ctx2 but not ctx1
		content := make([]byte, smallLimit+50)
		for i := range content {
			content[i] = 'A'
		}

		// Should fail in ctx1
		_, err := WriteFile(context.Background(), ctx1, model.WriteFileRequest{Path: "file.txt", Content: string(content)})
		if err != model.ErrTooLarge {
			t.Errorf("expected ErrTooLarge in ctx1, got %v", err)
		}

		// Should succeed in ctx2
		_, err = WriteFile(context.Background(), ctx2, model.WriteFileRequest{Path: "file.txt", Content: string(content)})
		if err != nil {
			t.Errorf("expected success in ctx2, got %v", err)
		}
	})
}
