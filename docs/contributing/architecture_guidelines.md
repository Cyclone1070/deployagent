# Go Architecture & Design Guidelines

> [!IMPORTANT]
> **Core Principle**: Make invalid states unrepresentable. Design your code so that it is impossible to misuse.
>
> These principles are strict guidelines to ensure maintainable, testable, and robust Go code. Any violations, no matter how small, will be rejected immediately during code review.

## Pre-Commit Checklist

Before submitting code, verify **every** item. A single unchecked box = rejection.

### Package Design
- [ ] No generic subdirectories (`model/`, `service/`, `utils/`, `types/`)
- [ ] No package exceeds 10-15 files (excluding `*_test.go`)
- [ ] No circular dependencies
- [ ] No junk drawer packages mixing unrelated logic

### Interfaces
- [ ] All interfaces defined in the **consumer** package, not the implementer
- [ ] All interfaces are minimal (≤5 methods)

### Structs & Encapsulation
- [ ] All domain entity fields are **private**
- [ ] All domain entities have `New*` constructors
- [ ] No direct struct initialization with `{}`
- [ ] DTOs have **no methods** attached
- [ ] No `json:`/`yaml:`/ORM tags on domain entities

### Validation & DI
- [ ] Static validation inside constructors
- [ ] Dynamic validation at start of method body (clearly commented)
- [ ] All dependencies injected via constructor
- [ ] No global state for dependencies

---

## 1. Package Design principles
**Goal**: Small, focused, reusable components.

*   **Small & Focused**: Packages should do one thing and do it well.

*   **File Organization**:
    *   **Prefer Files over Directories**: Do not create generic sub-directories like `model/`, `service/`, or `types/` inside your package.
    *   **Correct**: `feature/types.go`, `feature/service.go`.
    *   **Incorrect**: `feature/models/types.go`, `feature/services/service.go`.

> [!CAUTION]
> **ANTI-PATTERN 1**: Grouping by layer (Global)
> *   `controllers/`, `models/`, `services/`
>
> **ANTI-PATTERN 2**: Internal Layering (Inside Package)
> *   `feature/user/models/user.go`
> *   `feature/user/services/register.go`
>
> **CORRECT**: Grouping by feature/domain & Flat Package Files
> *   `feature/order/`, `feature/payment/`
> *   `feature/user/types.go`
> *   `feature/user/register.go`
> *   `feature/user/registration/register.go` (if complex)

*   **Splitting Rule (The "Too Big" Test)**:
    *   If a package grows large enough (around 10-15 files) that you feel the need to create a `models/` directory to organize multiple model files, **the package is too big**.
    *   **Action**: Break it down into smaller, focused sub-packages (e.g., `feature/subfeature1`, `feature/subfeature2`).

> [!CAUTION]
> **ANTI-PATTERN**: The "Flatten and Bloat" Wrong Fix
> 
> When removing generic subdirectories like `model/` or `service/`, do NOT blindly flatten all files into the parent package if this creates a bloated package.
>
> ```text
> # BEFORE: Internal Layering (WRONG)
> feature/
>   ├── models/     (8 files)
>   ├── services/   (12 files)
>   └── handlers/   (5 files)
>
> # WRONG FIX: Flatten Everything (STILL WRONG - now 25 files!)
> feature/
>   ├── user.go
>   ├── order.go
>   ├── ... (25 files total)
>
> # CORRECT FIX: Split by Domain
> feature/
>   ├── user/       (types.go, service.go, handler.go)
>   ├── order/      (types.go, service.go, handler.go)
>   └── payment/    (types.go, service.go, handler.go)
> ```
>
> *   **Why**: You've traded one anti-pattern (internal layering) for another (bloated package). Both violate "small and focused."
> *   **Rule**: If flattening would exceed 10-15 files, you MUST split into domain sub-packages instead.

*   **Hierarchy**: Nested packages are permitted and encouraged for grouping related sub-features, as long as they adhere to the circular dependency rule.

*   **Strict Rule**: **NO Circular Dependencies**. If you hit a circular dependency, your design is wrong. Refactor by extracting common definitions to a third package or decoupling via interfaces.

```text
# Bad Design (Layered) - ANTI-PATTERN
internal/
  ├── controllers/
  ├── services/
  └── models/

# Good Design (Domain) - RECOMMENDED
internal/
  ├── order/
  ├── payment/
  └── customer/
```

> [!CAUTION]
> **ANTI-PATTERN**: The "Junk Drawer" (Utils/Common)
> *   **Bad**: `feature/utils` or `feature/common` containing mixed logic (strings, encryption, formatting).
> *   **Why**: Violates cohesion. It becomes a dumping ground where dependencies tangle.
> *   **Solution**: Group by **what it operates on** or **domain**.
>     *   String helpers -> `feature/text` or `internal/strutil`
>     *   Time helpers -> `feature/timeext`
>     *   Domain logic -> `feature/auth/hashing` (NOT `feature/auth/utils`)
> *   **Exception**: A single `feature/utils` is permissible **ONLY IF** it relies strictly on the **Standard Library**. Once a function imports a 3rd party dependency, it MUST be moved to a specific package to prevent polluting the dependency tree.

## 2. Interfaces: Consumer-Defined
**Goal**: Decoupling and testability.

*   **Define where used**: Do NOT define interfaces in the implementing package. Define them in the consumer package.
    *   **Why**: The consumer knows what it needs. The implementer should not dictate the contract.
    *   **Benefit**: You can swap implementations without touching the consumer. You can mock easily in tests.

*   **Small Interfaces**: Keep interfaces minimal (`Reader` vs `ReadWriteCloser`).
    *   **Why**: Large interfaces force implementers to provide methods they don't use. This creates coupling and bloat.
    *   **Rule of Thumb**: If an interface has more than 5 methods, it's probably too big. Split it by use case.

*   **No Shared Interfaces**: Interfaces are **local** to the package that uses them. They are NOT shared across packages, even sibling packages.
    *   **Why**: Sibling packages should not know each other exist. If `file/` and `directory/` both need filesystem access, each defines its own interface with only the methods IT needs.
    *   **Trade-off**: This creates duplication. You accept a small amount of duplication in exchange for massive decoupling and testability. This is the correct trade-off.

> [!CAUTION]
> **ANTI-PATTERN**: Shared Interface Library
> *   **Bad**: Creating `internal/interfaces/filesystem.go` with a 10-method `FileSystem` interface that everyone imports.
> *   **Why**: This is just `model/` in disguise. It creates a central dependency, couples all consumers, and forces implementers to satisfy methods they don't need.
> *   **Solution**: Each consumer defines its own minimal interface. Duplication is acceptable. Coupling is not.


**Example**:
If `service` uses a database, `service` defines the `Repository` interface. The `database` package implements it.

```go
// package service
type UserRepository interface {
    Find(id string) (*User, error) // Defined here, where it's used
}

type Service struct {
    repo UserRepository
}
```

**Sibling Isolation Example**:
Both `file/` and `directory/` need filesystem operations, but each defines only what it needs:

```go
// package file
type fileSystem interface {
    Stat(path string) (FileInfo, error)
    ReadFileRange(path string, offset, limit int64) ([]byte, error)
}

// package directory (different package, defines its own interface)
type fileSystem interface {
    Stat(path string) (FileInfo, error)
    ListDir(path string) ([]FileInfo, error)
}

// Both are satisfied by the same concrete OSFileSystem,
// but neither package knows about the other's interface.
```

> [!CAUTION]
> **ANTI-PATTERN**: Leaky Interfaces
> *   **Bad**: `Save(u *User) (sql.Result, error)`
>     *   *Why*: `sql.Result` ties your interface to a SQL database. You can't implement this for a file system or memory store.
> *   **Good**: `Save(u *User) (string, error)`
>     *   *Why*: Returns the data you actually need (the new ID) in a generic way, without leaking implementation details.

## 3. Structs & Encapsulation
**Goal**: Control state and enforce invariants.

*   **Private Fields**: All Domain Entity fields MUST be private.

*   **Public Constructor**: You MUST provide a public `New...` constructor for every Domain Entity.
    *   **Strict Rule**: Direct initialization with `{}` is **forbidden** for Domain Entities, even if they currently have no validation logic.
    *   **Reason**: Future-proofing. If you add validation later, you shouldn't have to refactor every usage.

> [!CAUTION]
> **ANTI-PATTERN**: Constructor Bypass
> *   **Bad**: `user := &User{email: "..."}`
> *   **Good**: `user := NewUser("...")`
> *   **Why**: Bypassing the constructor skips validation and makes it impossible to guarantee invariants. It also breaks encapsulation.

*   **DTOs**: Use separate **Data Transfer Objects** (DTOs) with public fields for simple data passing (JSON, API) where no behavior/validation is attached.

```go
// Domain Entity
type User struct {
    id    string // Private: immutable once created
    email string // Private: validated format
}

// DTO
type UserDTO struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}
```

> [!CAUTION]
> **ANTI-PATTERN**: Tag Pollution on Domain Entities
> *   **Rule**: NEVER add `json:"..."`, `yaml:"..."`, or ORM tags (`gorm:"..."`) to **Domain Entities** (private structs).
> *   **Reason**: This couples your pure business logic to specific external interfaces or database implementations.
> *   **Solution**: Always use dedicated DTOs for serialization and separate DB Models for persistence.

## 4. Validation Strategy
**Goal**: Trust your objects.

### Type A: Static Validation (Invariants)
*   **Where**: Inside the Constructor (`New...`).
*   **Guarantee**: It is **impossible** to create an instance of the struct if these pass fail.
*   **Scope**: Internal consistency (e.g., "ID cannot be empty", "Email must have @").

```go
func NewUser(id, email string) (*User, error) {
    if id == "" {
        return nil, errors.New("id is required")
    }
    if !strings.Contains(email, "@") {
        return nil, errors.New("invalid email") // Static check
    }
    return &User{id: id, email: email}, nil
}
```

### Type B: Dynamic Validation (Business Rule / External)
*   **Where**: First thing in the method body.
*   **Scope**: Depends on external state (e.g., "File exists", "User is unique in DB").
*   **Requirement**: You MUST strictly separate and comment the validation section from the actual implementation logic to ensure readability.

```go
func (u *User) Save(repo UserRepository) error {
    // 1. Dynamic Validation
    if exists := repo.Exists(u.id); exists {
        return errors.New("user already exists")
    }

    // 2. Implementation
    return repo.Save(u)
}
```

## 5. Dependency Injection (DI) & Testing
**Goal**: Deterministic, isolated tests.

*   **Strict DI**: dependencies MUST be passed via constructor.
*   **No Globals**: Never use global state for dependencies.
*   **Testing**:
    *   **Mocks**: Use mocks for all dependencies.
    *   **No Temp Files**: Do not touch the filesystem in unit tests. Mock the `FileSystem` interface.
    *   **No Temp Dirs**: Do not use `os.TempDir` or `t.TempDir()` in unit logic tests.

**Correct Usage**:

```go
// Service definition
type FileProcessor struct {
    fs FileSystem // Interface
}

// Construction
func NewFileProcessor(fs FileSystem) *FileProcessor {
    return &FileProcessor{fs: fs}
}

// Testing
func TestFileProcessor(t *testing.T) {
    mockFS := new(MockFileSystem) // Mock implementation
    processor := NewFileProcessor(mockFS)
    // Test logic without touching disk
}
```
