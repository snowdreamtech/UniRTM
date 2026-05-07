# SQLite Repository Implementations

This package provides SQLite-based implementations of the repository interfaces defined in `internal/repository/repository.go`.

## Overview

The SQLite repositories provide persistent storage for UniRTM's core data:

- **InstallationRepository**: Manages tool installation records
- **CacheRepository**: Manages cached data with TTL support
- **AuditRepository**: Manages audit log entries with filtering and pagination
- **IndexRepository**: Manages tool index entries with search capabilities

## Features

### Performance Optimizations

- **Prepared Statements**: All repositories use prepared statements for repeated queries to improve performance
- **Database Indexes**: The schema includes indexes on frequently queried columns:
  - `installations`: indexed on `tool`, `installed_at`, and unique constraint on `(tool, version)`
  - `cache`: indexed on `expires_at` and primary key on `key`
  - `audit_log`: indexed on `timestamp`, `operation`, and `tool`
  - `tool_index`: indexed on `backend`, `updated_at`, and primary key on `tool`

### Error Handling

All repositories implement proper error handling with context wrapping:

- `ErrNotFound`: Returned when a resource is not found
- `ErrAlreadyExists`: Returned when attempting to create a duplicate resource
- All errors are wrapped with context using `fmt.Errorf` with `%w`

### Expiration Logic

The `CacheRepository` implements automatic expiration:

- Entries are stored with a TTL (time-to-live)
- `Get()` only returns non-expired entries
- `Purge()` removes all expired entries

### Filtering and Pagination

The `AuditRepository` supports flexible querying:

- Filter by time range (`StartTime`, `EndTime`)
- Filter by operation type (`install`, `uninstall`, etc.)
- Filter by tool name
- Filter by status (`success`, `failure`)
- Pagination with `Limit` and `Offset`

### Search Capabilities

The `IndexRepository` provides search functionality:

- Case-insensitive search across tool name, description, and metadata
- Uses SQL `LIKE` for pattern matching
- Returns results in alphabetical order

## Usage

### Creating Repositories

```go
import (
    "github.com/snowdreamtech/unirtm/internal/database"
    "github.com/snowdreamtech/unirtm/internal/repository/sqlite"
)

// Open database
db, err := database.Open(ctx, database.Config{
    Path:    "/path/to/unirtm.db",
    WALMode: true,
})
if err != nil {
    return err
}
defer db.Close()

// Create repositories
installRepo, err := sqlite.NewInstallationRepository(db.Conn())
if err != nil {
    return err
}
defer installRepo.Close()

cacheRepo, err := sqlite.NewCacheRepository(db.Conn())
if err != nil {
    return err
}
defer cacheRepo.Close()

auditRepo, err := sqlite.NewAuditRepository(db.Conn())
if err != nil {
    return err
}
defer auditRepo.Close()

indexRepo, err := sqlite.NewIndexRepository(db.Conn())
if err != nil {
    return err
}
defer indexRepo.Close()
```

### Installation Repository

```go
// Create installation
installation := &repository.Installation{
    Tool:        "node",
    Version:     "20.0.0",
    Backend:     "github",
    Provider:    "node",
    InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
    Checksum:    "abc123def456",
    Metadata:    `{"arch":"x64","os":"linux"}`,
}
err := installRepo.Create(ctx, installation)

// Find installation
found, err := installRepo.FindByToolAndVersion(ctx, "node", "20.0.0")

// List all installations
installations, err := installRepo.List(ctx)

// Delete installation
err := installRepo.Delete(ctx, "node", "20.0.0")
```

### Cache Repository

```go
// Set cache entry with TTL
err := cacheRepo.Set(ctx, "key", []byte("value"), 24*time.Hour)

// Get cache entry (returns nil if expired or not found)
value, err := cacheRepo.Get(ctx, "key")

// Delete cache entry
err := cacheRepo.Delete(ctx, "key")

// Purge expired entries
err := cacheRepo.Purge(ctx)
```

### Audit Repository

```go
// Log audit entry
entry := &repository.AuditEntry{
    Operation: "install",
    Tool:      "node",
    Version:   "20.0.0",
    Status:    "success",
    Duration:  1500,
    Metadata:  `{"backend":"github"}`,
}
err := auditRepo.Log(ctx, entry)

// Query with filters
filter := repository.AuditFilter{
    Operation: "install",
    Tool:      "node",
    Status:    "success",
    Limit:     10,
    Offset:    0,
}
entries, err := auditRepo.Query(ctx, filter)
```

### Index Repository

```go
// Upsert tool index entry
entry := &repository.IndexEntry{
    Tool:        "node",
    Description: "Node.js JavaScript runtime",
    Homepage:    "https://nodejs.org",
    License:     "MIT",
    Backend:     "github",
    Metadata:    `{"repo":"nodejs/node"}`,
}
err := indexRepo.Upsert(ctx, entry)

// Find by tool name
found, err := indexRepo.FindByTool(ctx, "node")

// List all tools
tools, err := indexRepo.List(ctx)

// Search tools
results, err := indexRepo.Search(ctx, "javascript")

// Delete tool
err := indexRepo.Delete(ctx, "node")
```

## Testing

The package includes comprehensive unit tests and integration tests:

```bash
# Run all tests
go test -v ./internal/repository/sqlite/...

# Run specific test
go test -v ./internal/repository/sqlite/... -run TestInstallationRepository

# Run with race detector
go test -race ./internal/repository/sqlite/...
```

### Test Coverage

- **Unit Tests**: Test each repository method individually
- **Integration Tests**: Test repositories working together in realistic workflows
- **Index Usage Tests**: Verify that database indexes are being used by queries

## Requirements Validation

This implementation validates the following requirements:

- **Requirement 2.2**: Store installation cache data (downloaded tarballs, extracted paths, checksums)
- **Requirement 2.3**: Store runtime state (active tool versions, environment resolution results)
- **Requirement 2.4**: Store tool indexes (available tools, GitHub releases, version lists)
- **Requirement 2.5**: Store audit logs (installation logs, execution logs, error stacks)

## Design Principles

- **Prepared Statements**: All repeated queries use prepared statements for performance
- **Error Wrapping**: All errors include context using `fmt.Errorf` with `%w`
- **Resource Cleanup**: All repositories implement `Close()` to clean up prepared statements
- **Idiomatic Go**: Follows Go best practices and conventions
- **Testability**: Designed for easy testing with temporary databases
