# Database Module

This module provides SQLite database management for UniRTM, including schema initialization, automatic migrations, and connection management.

## Features

- **Automatic Schema Initialization**: Creates all required tables and indexes on first run
- **Migration System**: Tracks and applies schema changes automatically
- **WAL Mode Support**: Enables Write-Ahead Logging for better concurrent read performance
- **Transaction Support**: Provides atomic operations with automatic rollback on failure
- **Idempotent Operations**: Safe to run initialization multiple times

## Usage

### Opening a Database

```go
import (
    "context"
    "github.com/snowdreamtech/unirtm/internal/database"
)

ctx := context.Background()
db, err := database.Open(ctx, database.Config{
    Path:    "/path/to/database.db",
    WALMode: true,
})
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Using Transactions

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Perform database operations
_, err = tx.ExecContext(ctx, `
    INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
    VALUES (?, ?, ?, ?, ?, ?)
`, "node", "20.0.0", "github", "generic", "/path/to/node", "abc123")
if err != nil {
    return err
}

// Commit the transaction
return tx.Commit()
```

### Checking Schema Version

```go
version, err := db.GetSchemaVersion(ctx)
if err != nil {
    return err
}
fmt.Printf("Current schema version: %d\n", version)
```

## Database Schema

### Tables

#### installations

Stores information about installed tools.

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| tool | TEXT | Tool name (e.g., "node", "python") |
| version | TEXT | Tool version (e.g., "20.0.0") |
| backend | TEXT | Backend used for installation |
| provider | TEXT | Provider used for installation |
| install_path | TEXT | Path to installation directory |
| checksum | TEXT | SHA-256 checksum of the artifact |
| installed_at | TIMESTAMP | Installation timestamp |
| metadata | TEXT | JSON-encoded metadata |

**Indexes:**

- `idx_installations_tool` on `tool`
- `idx_installations_installed_at` on `installed_at`

**Constraints:**

- `UNIQUE(tool, version)` - prevents duplicate installations

#### cache

Stores cached data with TTL support.

| Column | Type | Description |
|--------|------|-------------|
| key | TEXT | Cache key (primary key) |
| value | BLOB | Cached value |
| expires_at | TIMESTAMP | Expiration timestamp |
| created_at | TIMESTAMP | Creation timestamp |

**Indexes:**

- `idx_cache_expires_at` on `expires_at`

#### audit_log

Records all operations for audit purposes.

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| timestamp | TIMESTAMP | Operation timestamp |
| operation | TEXT | Operation type (install, uninstall, etc.) |
| tool | TEXT | Tool name |
| version | TEXT | Tool version |
| status | TEXT | Operation status (success, failure) |
| error | TEXT | Error message if failed |
| duration_ms | INTEGER | Operation duration in milliseconds |
| metadata | TEXT | JSON-encoded metadata |

**Indexes:**

- `idx_audit_log_timestamp` on `timestamp`
- `idx_audit_log_operation` on `operation`
- `idx_audit_log_tool` on `tool`

#### tool_index

Stores tool metadata and index information.

| Column | Type | Description |
|--------|------|-------------|
| tool | TEXT | Tool name (primary key) |
| description | TEXT | Tool description |
| homepage | TEXT | Tool homepage URL |
| license | TEXT | Tool license |
| backend | TEXT | Backend for this tool |
| updated_at | TIMESTAMP | Last update timestamp |
| metadata | TEXT | JSON-encoded metadata |

**Indexes:**

- `idx_tool_index_backend` on `backend`
- `idx_tool_index_updated_at` on `updated_at`

#### schema_migrations

Tracks applied database migrations.

| Column | Type | Description |
|--------|------|-------------|
| version | INTEGER | Migration version (primary key) |
| applied_at | TIMESTAMP | Application timestamp |
| description | TEXT | Migration description |

## Migration System

The migration system automatically applies schema changes when the database is opened. Migrations are applied in order and tracked in the `schema_migrations` table.

### Adding a New Migration

1. Add a new `Migration` to the `migrations` slice in `schema.go`:

```go
var migrations = []Migration{
    {
        Version:     1,
        Description: "Initial schema",
        Up:          Schema,
        Down:        "",
    },
    {
        Version:     2,
        Description: "Add new_column to installations",
        Up: `ALTER TABLE installations ADD COLUMN new_column TEXT;`,
        Down: `ALTER TABLE installations DROP COLUMN new_column;`,
    },
}
```

1. Update `CurrentSchemaVersion` in `schema.go`:

```go
const CurrentSchemaVersion = 2
```

1. The migration will be automatically applied on next database open.

### Migration Safety

- All migrations run in transactions
- Failed migrations are automatically rolled back
- Migrations are idempotent (safe to run multiple times)
- Schema version is tracked to prevent duplicate applications

## WAL Mode

Write-Ahead Logging (WAL) mode is enabled by default for better concurrent read performance:

- Multiple readers can access the database simultaneously
- Readers don't block writers
- Better performance for read-heavy workloads

## Testing

Run tests with:

```bash
go test ./internal/database/...
```

Run tests with race detector:

```bash
go test -race ./internal/database/...
```

## Design Decisions

### SQLite Choice

- Embedded database (no separate server process)
- Zero configuration
- ACID transactions
- Cross-platform support
- Sufficient performance for local tool management

### WAL Mode

- Enables concurrent reads without blocking
- Better performance for read-heavy workloads
- Minimal overhead for write operations

### Migration System

- Automatic schema evolution
- Version tracking prevents duplicate applications
- Transaction-based for safety
- Rollback support for development

### Connection Pooling

- Single writer connection (SQLite limitation)
- Prevents write contention
- Readers can still access concurrently in WAL mode

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [SQLite WAL Mode](https://www.sqlite.org/wal.html)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
