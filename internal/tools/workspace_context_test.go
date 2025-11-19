package tools

import (
	"os"
	"testing"

	"github.com/Cyclone1070/deployforme/internal/tools/models"
	"github.com/Cyclone1070/deployforme/internal/tools/services"
)

func TestMultiContextIsolation(t *testing.T) {
	maxFileSize := int64(1024 * 1024) // 1MB

	// Create two separate contexts with different workspace roots
	fs1 := services.NewMockFileSystem(maxFileSize)
	checksumManager1 := services.NewChecksumManager()

	fs2 := services.NewMockFileSystem(maxFileSize)
	checksumManager2 := services.NewChecksumManager()

	ctx1 := &models.WorkspaceContext{
		FS:              fs1,
		BinaryDetector:  services.NewMockBinaryDetector(),
		ChecksumManager: checksumManager1,
		MaxFileSize:      maxFileSize,
		WorkspaceRoot:    "/workspace1",
	}

	ctx2 := &models.WorkspaceContext{
		FS:              fs2,
		BinaryDetector:  services.NewMockBinaryDetector(),
		ChecksumManager: checksumManager2,
		MaxFileSize:      maxFileSize,
		WorkspaceRoot:    "/workspace2",
	}

	// Create files in both contexts
	content1 := "content1"
	content2 := "content2"

	resp1, err := WriteFile(ctx1, "file.txt", content1, nil)
	if err != nil {
		t.Fatalf("failed to write file in ctx1: %v", err)
	}

	resp2, err := WriteFile(ctx2, "file.txt", content2, nil)
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
	read1, err := ReadFile(ctx1, "file.txt", nil, nil)
	if err != nil {
		t.Fatalf("failed to read file from ctx1: %v", err)
	}
	if read1.Content != content1 {
		t.Errorf("ctx1 should read its own content, got %q", read1.Content)
	}

	read2, err := ReadFile(ctx2, "file.txt", nil, nil)
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
		fs := services.NewMockFileSystem(smallLimit)
		checksumManager := services.NewChecksumManager()

		ctx := &models.WorkspaceContext{
			FS:              fs,
			BinaryDetector:  services.NewMockBinaryDetector(),
			ChecksumManager: checksumManager,
			MaxFileSize:      smallLimit,
			WorkspaceRoot:    workspaceRoot,
		}

		// Create content that exceeds the limit
		largeContent := make([]byte, smallLimit+1)
		for i := range largeContent {
			largeContent[i] = 'A'
		}

		_, err := WriteFile(ctx, "large.txt", string(largeContent), nil)
		if err != models.ErrTooLarge {
			t.Errorf("expected ErrTooLarge for content exceeding limit, got %v", err)
		}
	})

	t.Run("large limit allows bigger files", func(t *testing.T) {
		fs := services.NewMockFileSystem(largeLimit)
		checksumManager := services.NewChecksumManager()

		ctx := &models.WorkspaceContext{
			FS:              fs,
			BinaryDetector:  services.NewMockBinaryDetector(),
			ChecksumManager: checksumManager,
			MaxFileSize:      largeLimit,
			WorkspaceRoot:    workspaceRoot,
		}

		// Create content within the large limit but exceeding default
		content := make([]byte, 6*1024*1024) // 6MB, exceeds default 5MB
		for i := range content {
			content[i] = 'A'
		}

		_, err := WriteFile(ctx, "large.txt", string(content), nil)
		if err != nil {
			t.Errorf("expected success with large limit, got %v", err)
		}
	})

	t.Run("different limits in different contexts", func(t *testing.T) {
		fs1 := services.NewMockFileSystem(smallLimit)
		checksumManager1 := services.NewChecksumManager()

		fs2 := services.NewMockFileSystem(largeLimit)
		checksumManager2 := services.NewChecksumManager()

		ctx1 := &models.WorkspaceContext{
			FS:              fs1,
			BinaryDetector:  services.NewMockBinaryDetector(),
			ChecksumManager: checksumManager1,
			MaxFileSize:      smallLimit,
			WorkspaceRoot:    workspaceRoot,
		}

		ctx2 := &models.WorkspaceContext{
			FS:              fs2,
			BinaryDetector:  services.NewMockBinaryDetector(),
			ChecksumManager: checksumManager2,
			MaxFileSize:      largeLimit,
			WorkspaceRoot:    workspaceRoot,
		}

		// Content that fits in ctx2 but not ctx1
		content := make([]byte, smallLimit+50)
		for i := range content {
			content[i] = 'A'
		}

		// Should fail in ctx1
		_, err := WriteFile(ctx1, "file.txt", string(content), nil)
		if err != models.ErrTooLarge {
			t.Errorf("expected ErrTooLarge in ctx1, got %v", err)
		}

		// Should succeed in ctx2
		_, err = WriteFile(ctx2, "file.txt", string(content), nil)
		if err != nil {
			t.Errorf("expected success in ctx2, got %v", err)
		}
	})
}

func TestNewWorkspaceContext(t *testing.T) {
	t.Run("creates context with default max file size", func(t *testing.T) {
		// Use a temporary directory for testing
		tmpDir := t.TempDir()

		ctx, err := NewWorkspaceContext(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx == nil {
			t.Fatal("expected non-nil context")
		}

		// Verify workspace root is set
		if ctx.WorkspaceRoot == "" {
			t.Error("expected non-empty workspace root")
		}

		// Verify default max file size
		if ctx.MaxFileSize != models.DefaultMaxFileSize {
			t.Errorf("expected default max file size %d, got %d", models.DefaultMaxFileSize, ctx.MaxFileSize)
		}

		// Verify all dependencies are set
		if ctx.FS == nil {
			t.Error("expected non-nil FileSystem")
		}
		if ctx.BinaryDetector == nil {
			t.Error("expected non-nil BinaryDetector")
		}
		if ctx.ChecksumManager == nil {
			t.Error("expected non-nil ChecksumManager")
		}
	})

	t.Run("rejects non-existent directory", func(t *testing.T) {
		nonExistent := "/nonexistent/path/that/does/not/exist"

		_, err := NewWorkspaceContext(nonExistent)
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})

	t.Run("rejects file instead of directory", func(t *testing.T) {
		tmpFile := t.TempDir() + "/testfile.txt"
		// Create a file
		if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		_, err := NewWorkspaceContext(tmpFile)
		if err == nil {
			t.Error("expected error for file instead of directory")
		}
	})
}

func TestNewWorkspaceContextWithOptions(t *testing.T) {
	t.Run("creates context with custom max file size", func(t *testing.T) {
		tmpDir := t.TempDir()
		customMaxSize := int64(10 * 1024 * 1024) // 10MB

		ctx, err := NewWorkspaceContextWithOptions(tmpDir, customMaxSize)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ctx == nil {
			t.Fatal("expected non-nil context")
		}

		// Verify custom max file size
		if ctx.MaxFileSize != customMaxSize {
			t.Errorf("expected max file size %d, got %d", customMaxSize, ctx.MaxFileSize)
		}

		// Verify workspace root is set
		if ctx.WorkspaceRoot == "" {
			t.Error("expected non-empty workspace root")
		}

		// Verify all dependencies are set
		if ctx.FS == nil {
			t.Error("expected non-nil FileSystem")
		}
		if ctx.BinaryDetector == nil {
			t.Error("expected non-nil BinaryDetector")
		}
		if ctx.ChecksumManager == nil {
			t.Error("expected non-nil ChecksumManager")
		}
	})

	t.Run("rejects invalid workspace root", func(t *testing.T) {
		nonExistent := "/invalid/path/that/does/not/exist"

		_, err := NewWorkspaceContextWithOptions(nonExistent, models.DefaultMaxFileSize)
		if err == nil {
			t.Error("expected error for invalid workspace root")
		}
	})
}
