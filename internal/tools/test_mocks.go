package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// mockFileInfo implements FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (f *mockFileInfo) Name() string       { return f.name }
func (f *mockFileInfo) Size() int64        { return f.size }
func (f *mockFileInfo) Mode() os.FileMode  { return f.mode }
func (f *mockFileInfo) ModTime() time.Time { return f.modTime }
func (f *mockFileInfo) IsDir() bool        { return f.isDir }

// MockFileSystem implements FileSystem with in-memory storage
type MockFileSystem struct {
	mu          sync.RWMutex
	files       map[string][]byte        // path -> content
	fileInfos   map[string]*mockFileInfo // path -> metadata
	symlinks    map[string]string        // symlink path -> target path
	dirs        map[string]bool          // path -> is directory
	errors      map[string]error         // path -> error to return
	maxFileSize int64
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem(maxFileSize int64) *MockFileSystem {
	return &MockFileSystem{
		files:       make(map[string][]byte),
		fileInfos:   make(map[string]*mockFileInfo),
		symlinks:    make(map[string]string),
		dirs:        make(map[string]bool),
		errors:      make(map[string]error),
		maxFileSize: maxFileSize,
	}
}

// SetError sets an error to return for a specific path
func (f *MockFileSystem) SetError(path string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errors[path] = err
}

// CreateFile creates a file with content
func (f *MockFileSystem) CreateFile(path string, content []byte, modTime time.Time, perm os.FileMode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[path] = content
	f.fileInfos[path] = &mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    perm,
		modTime: modTime,
		isDir:   false,
	}
	f.dirs[path] = false
}

// CreateDir creates a directory
func (f *MockFileSystem) CreateDir(path string, modTime time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dirs[path] = true
	f.fileInfos[path] = &mockFileInfo{
		name:    filepath.Base(path),
		size:    0,
		mode:    os.ModeDir | 0755,
		modTime: modTime,
		isDir:   true,
	}
}

// CreateSymlink creates a symlink
func (f *MockFileSystem) CreateSymlink(symlinkPath, targetPath string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.symlinks[symlinkPath] = targetPath
	f.fileInfos[symlinkPath] = &mockFileInfo{
		name:    filepath.Base(symlinkPath),
		size:    0,
		mode:    os.ModeSymlink | 0777,
		modTime: time.Now(),
		isDir:   false,
	}
}

func (f *MockFileSystem) Stat(path string) (FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err, ok := f.errors[path]; ok {
		return nil, err
	}

	if info, ok := f.fileInfos[path]; ok {
		return info, nil
	}

	return nil, os.ErrNotExist
}

func (f *MockFileSystem) Lstat(path string) (FileInfo, error) {
	// For mock, Lstat is the same as Stat since we track symlinks explicitly
	return f.Stat(path)
}

func (f *MockFileSystem) ReadFileRange(path string, offset, limit int64) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err, ok := f.errors[path]; ok {
		return nil, err
	}

	content, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	fileSize := int64(len(content))

	if fileSize > f.maxFileSize {
		return nil, ErrTooLarge
	}

	if offset == 0 && limit == 0 {
		return content, nil
	}

	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	if offset >= fileSize {
		return []byte{}, nil
	}

	remaining := fileSize - offset
	var readSize int64

	if limit == 0 {
		readSize = remaining
	} else {
		if remaining < limit {
			readSize = remaining
		} else {
			readSize = limit
		}
	}

	end := offset + readSize
	if end > fileSize {
		end = fileSize
	}

	return content[offset:end], nil
}

func (f *MockFileSystem) WriteFile(path string, content []byte, perm os.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err, ok := f.errors[path]; ok {
		return err
	}

	if int64(len(content)) > f.maxFileSize {
		return ErrTooLarge
	}

	f.files[path] = content
	f.fileInfos[path] = &mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    perm,
		modTime: time.Now(),
		isDir:   false,
	}
	f.dirs[path] = false

	return nil
}

func (f *MockFileSystem) EnsureDirs(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	cleaned := filepath.Clean(path)
	parts := strings.Split(cleaned, string(filepath.Separator))

	// Handle absolute paths
	var current string
	startIdx := 0
	if filepath.IsAbs(cleaned) {
		if len(parts) > 0 && parts[0] == "" {
			startIdx = 1
			current = "/"
		}
	}

	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		if part == "" {
			continue
		}
		if current == "" {
			current = part
		} else if current == "/" {
			current = "/" + part
		} else {
			current = filepath.Join(current, part)
		}
		if !f.dirs[current] {
			f.dirs[current] = true
			f.fileInfos[current] = &mockFileInfo{
				name:    part,
				size:    0,
				mode:    os.ModeDir | 0755,
				modTime: time.Now(),
				isDir:   true,
			}
		}
	}
	return nil
}

func (f *MockFileSystem) IsDir(path string) (bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err, ok := f.errors[path]; ok {
		return false, err
	}

	if info, ok := f.fileInfos[path]; ok {
		return info.IsDir(), nil
	}

	return false, os.ErrNotExist
}

func (f *MockFileSystem) Readlink(path string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if err, ok := f.errors[path]; ok {
		return "", err
	}

	target, ok := f.symlinks[path]
	if !ok {
		return "", fmt.Errorf("not a symlink: %s", path)
	}

	return target, nil
}

func (f *MockFileSystem) EvalSymlinks(path string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Follow symlink chain
	visited := make(map[string]bool)
	current := path

	for {
		if visited[current] {
			return "", fmt.Errorf("symlink loop detected")
		}
		visited[current] = true

		target, ok := f.symlinks[current]
		if !ok {
			// Not a symlink, return as-is
			return current, nil
		}

		if filepath.IsAbs(target) {
			current = target
		} else {
			current = filepath.Join(filepath.Dir(current), target)
		}
		current = filepath.Clean(current)
	}
}

func (f *MockFileSystem) Abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(path)
}

func (f *MockFileSystem) UserHomeDir() (string, error) {
	return "/home/user", nil
}

// MockBinaryDetector implements BinaryDetector with configurable behaviour
type MockBinaryDetector struct {
	binaryPaths   map[string]bool
	binaryContent map[string]bool // content hash -> is binary
}

// NewMockBinaryDetector creates a new mock binary detector
func NewMockBinaryDetector() *MockBinaryDetector {
	return &MockBinaryDetector{
		binaryPaths:   make(map[string]bool),
		binaryContent: make(map[string]bool),
	}
}

// SetBinaryPath marks a path as binary
func (f *MockBinaryDetector) SetBinaryPath(path string, isBinary bool) {
	f.binaryPaths[path] = isBinary
}

func (f *MockBinaryDetector) IsBinary(path string) (bool, error) {
	if isBinary, ok := f.binaryPaths[path]; ok {
		return isBinary, nil
	}
	// Default: check for NUL bytes
	return false, nil
}

func (f *MockBinaryDetector) IsBinaryContent(content []byte) bool {
	sampleSize := BinaryDetectionSampleSize
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	for i := 0; i < sampleSize; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}

// MockChecksumComputer implements ChecksumComputer with deterministic behaviour
type MockChecksumComputer struct {
	checksums map[string]string // content -> checksum
	mu        sync.RWMutex
}

// NewMockChecksumComputer creates a new mock checksum computer
func NewMockChecksumComputer() *MockChecksumComputer {
	return &MockChecksumComputer{
		checksums: make(map[string]string),
	}
}

func (f *MockChecksumComputer) ComputeChecksum(data []byte) string {
	// Use real SHA-256 implementation for deterministic behaviour
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// MockClock implements Clock with controllable time
type MockClock struct {
	mu    sync.RWMutex
	times []time.Time
	index int
}

// NewMockClock creates a new mock clock
func NewMockClock() *MockClock {
	return &MockClock{
		times: []time.Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		index: 0,
	}
}

// SetTime sets the current time
func (f *MockClock) SetTime(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.times = []time.Time{t}
	f.index = 0
}

// AdvanceTime adds a time to the sequence
func (f *MockClock) AdvanceTime(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.times = append(f.times, t)
}

func (f *MockClock) Now() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.index < len(f.times) {
		return f.times[f.index]
	}
	return f.times[len(f.times)-1]
}

// Tick advances to the next time
func (f *MockClock) Tick() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.index < len(f.times)-1 {
		f.index++
	}
}

// MockChecksumStore implements ChecksumStore with in-memory storage
type MockChecksumStore struct {
	mu    sync.RWMutex
	store map[string]FileMetadata
}

// NewMockChecksumStore creates a new mock checksum store
func NewMockChecksumStore() *MockChecksumStore {
	return &MockChecksumStore{
		store: make(map[string]FileMetadata),
	}
}

func (f *MockChecksumStore) Get(path string) (checksum string, ok bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	meta, ok := f.store[path]
	if !ok {
		return "", false
	}
	return meta.Checksum, true
}

func (f *MockChecksumStore) Update(path string, checksum string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[path] = FileMetadata{
		Checksum: checksum,
	}
}

func (f *MockChecksumStore) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store = make(map[string]FileMetadata)
}

// MockRootCanonicalizer implements RootCanonicalizer for testing
type MockRootCanonicalizer struct {
	mu           sync.RWMutex
	canonicalMap map[string]string // input -> canonical output
	errors       map[string]error  // input -> error
}

// NewMockRootCanonicalizer creates a new mock root canonicalizer
func NewMockRootCanonicalizer() *MockRootCanonicalizer {
	return &MockRootCanonicalizer{
		canonicalMap: make(map[string]string),
		errors:       make(map[string]error),
	}
}

// SetCanonical sets the canonical path for an input path
func (m *MockRootCanonicalizer) SetCanonical(input, canonical string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.canonicalMap[input] = canonical
}

// SetError sets an error to return for a given input path
func (m *MockRootCanonicalizer) SetError(input string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[input] = err
}

func (m *MockRootCanonicalizer) CanonicalizeRoot(root string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.errors[root]; ok {
		return "", err
	}

	if canonical, ok := m.canonicalMap[root]; ok {
		return canonical, nil
	}

	// Default: return as-is if not configured
	return root, nil
}
