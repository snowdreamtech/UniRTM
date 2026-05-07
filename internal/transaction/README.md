# Transaction Manager

This package provides transaction management for UniRTM's database operations, ensuring atomicity and consistency across multiple repository operations.

## Overview

The transaction manager implements the transaction pattern for SQLite database operations, providing:

- **Atomic Operations**: All operations within a transaction either succeed completely or fail completely
- **Transaction-Scoped Repositories**: Each transaction provides its own repository instances that operate within the transaction scope
- **Explicit Commit/Rollback**: Transactions must be explicitly committed or rolled back
- **Automatic Rollback on Error**: Failed operations can trigger rollback to maintain consistency

## Architecture

### Interfaces

#### TransactionManager

```go
type TransactionManager interface {
    Begin(ctx context.Context) (Transaction, error)
}
```

The `TransactionManager` is responsible for creating new transactions.

#### Transaction

```go
type Transaction interface {
    Commit() error
    Rollback() error
    InstallationRepo() repository.InstallationRepository
    CacheRepo() repository.CacheRepository
    AuditRepo() repository.AuditRepository
    IndexRepo() repository.IndexRepository
}
```

The `Transaction` interface provides:
- Transaction control methods (`Commit`, `Rollback`)
- Access to transaction-scoped repositories

### Implementation

The package provides a SQLite-specific implementation:

- `sqliteTransactionManager`: Implements `TransactionManager` using `*sql.DB`
- `sqliteTransaction`: Implements `Transaction` using `*sql.Tx`

## Usage

### Basic Transaction

```go
// Create transaction manager
tm := transaction.NewSQLiteTransactionManager(db)

// Begin transaction
tx, err := tm.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}

// Always ensure rollback on error
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// Perform operations
installation := &repository.Installation{
    Tool:        "node",
    Version:     "20.0.0",
    Backend:     "github",
    Provider:    "node",
    InstallPath: "/opt/unirtm/node/20.0.0",
    Checksum:    "abc123",
    Metadata:    "{}",
}

err = tx.InstallationRepo().Create(ctx, installation)
if err != nil {
    return fmt.Errorf("create installation: %w", err)
}

// Commit transaction
err = tx.Commit()
if err != nil {
    return fmt.Errorf("commit transaction: %w", err)
}
```

### Multi-Repository Transaction

```go
tx, err := tm.Begin(ctx)
if err != nil {
    return err
}
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// Create installation
installation := &repository.Installation{...}
err = tx.InstallationRepo().Create(ctx, installation)
if err != nil {
    return err
}

// Log audit entry
auditEntry := &repository.AuditEntry{
    Operation: "install",
    Tool:      installation.Tool,
    Version:   installation.Version,
    Status:    "success",
}
err = tx.AuditRepo().Log(ctx, auditEntry)
if err != nil {
    return err
}

// Update index
indexEntry := &repository.IndexEntry{
    Tool:        installation.Tool,
    Description: "Tool description",
    Backend:     installation.Backend,
}
err = tx.IndexRepo().Upsert(ctx, indexEntry)
if err != nil {
    return err
}

// Commit all operations atomically
return tx.Commit()
```

### Error Handling with Rollback

```go
tx, err := tm.Begin(ctx)
if err != nil {
    return err
}

// Automatic rollback on any error
defer func() {
    if err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            log.Printf("rollback failed: %v", rbErr)
        }
    }
}()

// Perform operations...
err = performOperations(tx)
if err != nil {
    return err // Deferred rollback will execute
}

// Explicit commit
return tx.Commit()
```

## Requirements Validation

This implementation validates the following requirements:

- **Requirement 2.8**: Use transactions for all write operations to ensure atomicity
- **Requirement 3.3**: Support explicit commit operations for multi-step workflows

## Design Decisions

### Repository Interface Abstraction

The repositories use a `DBExecutor` interface that is implemented by both `*sql.DB` and `*sql.Tx`. This allows:

- Repositories to work with both regular connections and transactions
- Transaction-scoped repository instances without code duplication
- Type-safe transaction boundaries

### Explicit Transaction Control

Transactions require explicit `Commit()` or `Rollback()` calls:

- **Pros**: Clear transaction boundaries, explicit error handling
- **Cons**: Requires careful cleanup (use `defer` for rollback)

This design follows Go best practices and makes transaction boundaries explicit in the code.

### Transaction-Scoped Repositories

Each transaction creates its own repository instances:

- Ensures all operations use the same transaction
- Prevents accidental mixing of transactional and non-transactional operations
- Clear ownership of transaction scope

## Testing

The package includes comprehensive unit tests covering:

- Transaction creation and basic operations
- Commit and rollback behavior
- Multi-repository atomic operations
- Error handling and automatic rollback
- Transaction isolation
- Context cancellation
- Edge cases (double commit, double rollback)

Run tests:

```bash
go test ./internal/transaction/...
```

## Performance Considerations

- **Prepared Statements**: Repositories use prepared statements for performance
- **Connection Pooling**: SQLite connection pooling is managed by the database layer
- **WAL Mode**: Write-Ahead Logging mode is recommended for better concurrent read performance
- **Transaction Scope**: Keep transactions short to minimize lock contention

## Future Enhancements

Potential improvements for future iterations:

1. **Savepoints**: Support for nested transactions using SQLite savepoints
2. **Read-Only Transactions**: Optimize read-only transaction paths
3. **Transaction Retry**: Automatic retry logic for transient failures
4. **Transaction Metrics**: Instrumentation for transaction duration and success rates
5. **Distributed Transactions**: Support for coordinating transactions across multiple databases (if needed)
