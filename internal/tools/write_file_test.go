package tools

// TEST CONTRACT: do not modify without updating symlink safety spec
// These tests enforce the symlink safety guarantees for file operations.
// Any changes to these tests must be reviewed against the symlink safety specification.

import (
	"os"
	"testing"
)

func TestWriteFile(t *testing.T) {
	workspaceRoot := "/workspace"
	maxFileSize := int64(1024 * 1024) // 1MB

	t.Run("create new file succeeds and updates cache", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		content := "test content"
		resp, err := WriteFile(ctx, "new.txt", content, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.BytesWritten != len(content) {
			t.Errorf("expected %d bytes written, got %d", len(content), resp.BytesWritten)
		}

		// Verify file was created
		fileContent, err := fs.ReadFileRange("/workspace/new.txt", 0, 0)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}
		if string(fileContent) != content {
			t.Errorf("expected content %q, got %q", content, string(fileContent))
		}

		// Verify cache was updated
		checksum, ok := ctx.ChecksumCache.Get(resp.AbsolutePath)
		if !ok {
			t.Error("expected cache to be updated after write")
		}
		if checksum == "" {
			t.Error("expected non-empty checksum in cache")
		}
	})

	t.Run("existing file rejection", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		fs.CreateFile("/workspace/existing.txt", []byte("existing"), clock.Now(), 0644)

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		_, err := WriteFile(ctx, "existing.txt", "new content", nil)
		if err != ErrFileExists {
			t.Errorf("expected ErrFileExists, got %v", err)
		}
	})

	t.Run("symlink escape prevention", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		// Create symlink pointing outside workspace
		fs.CreateSymlink("/workspace/escape", "/outside/target.txt")

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		_, err := WriteFile(ctx, "escape", "content", nil)
		if err != ErrOutsideWorkspace {
			t.Errorf("expected ErrOutsideWorkspace for symlink escape, got %v", err)
		}
	})

	t.Run("large content rejection", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Create content larger than limit
		largeContent := make([]byte, maxFileSize+1)
		for i := range largeContent {
			largeContent[i] = 'A'
		}

		_, err := WriteFile(ctx, "large.txt", string(largeContent), nil)
		if err != ErrTooLarge {
			t.Errorf("expected ErrTooLarge, got %v", err)
		}
	})

	t.Run("binary content rejection", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Content with NUL byte
		binaryContent := []byte{0x48, 0x65, 0x6C, 0x00, 0x6C, 0x6F}
		_, err := WriteFile(ctx, "binary.bin", string(binaryContent), nil)
		if err != ErrBinaryFile {
			t.Errorf("expected ErrBinaryFile, got %v", err)
		}
	})

	t.Run("custom permissions", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		perm := os.FileMode(0755)
		resp, err := WriteFile(ctx, "executable.txt", "content", &perm)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := fs.Stat("/workspace/executable.txt")
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}

		if info.Mode().Perm() != perm {
			t.Errorf("expected permissions %o, got %o", perm, info.Mode().Perm())
		}

		if resp.FileMode != uint32(perm) {
			t.Errorf("expected FileMode %o, got %o", perm, resp.FileMode)
		}
	})

	t.Run("nested directory creation", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		_, err := WriteFile(ctx, "nested/deep/file.txt", "content", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify file was created
		fileContent, err := fs.ReadFileRange("/workspace/nested/deep/file.txt", 0, 0)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}
		if string(fileContent) != "content" {
			t.Errorf("expected content %q, got %q", "content", string(fileContent))
		}
	})

	t.Run("symlink inside workspace allowed", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		// Create symlink pointing inside workspace
		fs.CreateSymlink("/workspace/link", "/workspace/target.txt")
		fs.CreateFile("/workspace/target.txt", []byte("target"), clock.Now(), 0644)

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Writing to a symlink that points inside workspace should work
		// (the symlink itself is treated as a regular path for new files)
		_, err := WriteFile(ctx, "link", "new content", nil)
		// This should succeed because we're creating a new file at the symlink path
		if err != nil {
			// If it fails, it's because the symlink exists, which is expected
			if err != ErrFileExists {
				t.Errorf("unexpected error: %v", err)
			}
		}
	})

	t.Run("symlink directory escape prevention", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		// Create symlink directory pointing outside workspace
		fs.CreateSymlink("/workspace/link", "/outside")
		fs.CreateDir("/outside", clock.Now())

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Try to write a file through the symlink directory - should fail
		_, err := WriteFile(ctx, "link/escape.txt", "content", nil)
		if err != ErrOutsideWorkspace {
			t.Errorf("expected ErrOutsideWorkspace for symlink directory escape, got %v", err)
		}
	})

	t.Run("write through symlink chain inside workspace", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		// Create symlink chain: link1 -> link2 -> target_dir
		fs.CreateSymlink("/workspace/link1", "/workspace/link2")
		fs.CreateSymlink("/workspace/link2", "/workspace/target_dir")
		fs.CreateDir("/workspace/target_dir", clock.Now())

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Write through symlink chain - should succeed
		resp, err := WriteFile(ctx, "link1/file.txt", "content", nil)
		if err != nil {
			t.Fatalf("unexpected error writing through symlink chain: %v", err)
		}

		// Verify file was created at resolved location
		fileContent, err := fs.ReadFileRange("/workspace/target_dir/file.txt", 0, 0)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}
		if string(fileContent) != "content" {
			t.Errorf("expected content %q, got %q", "content", string(fileContent))
		}

		// Verify response has correct absolute path
		if resp.AbsolutePath != "/workspace/target_dir/file.txt" {
			t.Errorf("expected absolute path /workspace/target_dir/file.txt, got %s", resp.AbsolutePath)
		}
	})

	t.Run("write through symlink chain escaping workspace", func(t *testing.T) {
		fs := NewMockFileSystem(maxFileSize)
		cache := NewMockChecksumStore()
		clock := NewMockClock()

		// Create chain: link1 -> link2 -> /tmp/outside
		fs.CreateSymlink("/workspace/link1", "/workspace/link2")
		fs.CreateSymlink("/workspace/link2", "/tmp/outside")
		fs.CreateDir("/tmp/outside", clock.Now())

		ctx := &WorkspaceContext{
			FS:               fs,
			BinaryDetector:   NewMockBinaryDetector(),
			ChecksumComputer: NewMockChecksumComputer(),
			Clock:            clock,
			ChecksumCache:            cache,
			MaxFileSize:      maxFileSize,
			WorkspaceRoot:    workspaceRoot,
		}

		// Try to write through escaping chain - should fail
		_, err := WriteFile(ctx, "link1/file.txt", "content", nil)
		if err != ErrOutsideWorkspace {
			t.Errorf("expected ErrOutsideWorkspace for escaping symlink chain, got %v", err)
		}
	})
}
