# Repository Layer

This package implements the data access layer using SQLite.

## Responsibilities

- Define repository interfaces for data access
- Implement SQLite repository implementations
- Manage database schema and migrations
- Handle transactions and concurrent access
- Store installation records, cache, audit logs, and tool index
- Implement query filters and pagination

## Key Components

- `repository.go` - Repository interface definitions
- `installation.go` - Installation repository implementation
- `cache.go` - Cache repository implementation
- `audit.go` - Audit log repository implementation
- `index.go` - Tool index repository implementation
- `transaction.go` - Transaction manager implementation
- `schema.go` - Database schema definitions
- `migration.go` - Database migration system

## Usage

```go
import "github.com/snowdreamtech/unirtm/internal/repository"

// Open database
db, err := repository.Open("unirtm.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Use repositories
installRepo := repository.NewInstallationRepository(db)
installations, err := installRepo.List(ctx)
if err != nil {
    log.Fatal(err)
}
```

## Requirements

Implements requirements: 2.1-2.9
