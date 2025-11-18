# Deploy Agent

## Symlink Safety

The path resolution system (`internal/tools/pathutil.go`) enforces strict workspace boundary checks to prevent symlink escape attacks. All file operations (`ReadFile`, `WriteFile`, `EditFile`) rely on the `Resolve` function which:

- Resolves symlink chains component-by-component
- Validates every symlink hop stays within the workspace boundary
- Rejects any path that would escape the workspace, even temporarily
- Handles missing directories gracefully to allow directory creation
- Detects and rejects symlink loops

The `resolveSymlink` helper function fully resolves symlink chains while enforcing workspace boundaries at each hop.

### Test Contract

The test files in `internal/tools/` (specifically `*_test.go` files) contain a **TEST CONTRACT** that locks the symlink safety tests. These tests enforce the security guarantees described above. **Do not modify these tests without updating the symlink safety specification.** The locked tests are marked with comments at the top of each test file.
