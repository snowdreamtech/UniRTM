# Cache Manager Implementation

## Overview

The Cache Manager is a service layer component that provides high-level caching operations with TTL support, checksum verification, automatic cleanup, and performance monitoring. It implements Requirements 10.1-10.8 from the UniRTM specification.

## Architecture

The Cache Manager sits in the service layer and uses the `CacheRepository` interface from the repository layer to persist cache entries in the SQLite database. It also integrates with the `AuditRepository` for logging cache operations.

```
┌─────────────────────────────────────────┐
│         Service Layer                   │
│  ┌───────────────────────────────────┐  │
│  │      CacheManager                 │  │
│  │  - Set/Get/Delete operations      │  │
│  │  - Checksum verification          │  │
│  │  - TTL management                 │  │
│  │  - Automatic cleanup              │  │
│  │  - Hit/Miss tracking              │  │
│  └───────────────────────────────────┘  │
│           │                    │         │
└───────────┼────────────────────┼─────────┘
            │                    │
            ▼                    ▼
┌─────────────────────┐  ┌──────────────────┐
│  CacheRepository    │  │  AuditRepository │
│  (Repository Layer) │  │  (Repository     │
│                     │  │   Layer)         │
└─────────────────────┘  └──────────────────┘
            │                    │
            ▼                    ▼
┌─────────────────────────────────────────┐
│         SQLite Database                 │
│  - cache table (key, value, expires_at) │
│  - audit_log table                      │
└─────────────────────────────────────────┘
```

## Features

### 1. Cache Storage with TTL (Requirements 10.1, 10.2, 10.3)

The Cache Manager stores cache entries with configurable Time-To-Live (TTL):

```go
// Store GitHub release metadata with 24-hour TTL
err := cacheManager.Set(ctx, "github:releases:node", metadata, 24*time.Hour)

// Store version resolution results with 1-hour TTL
err := cacheManager.Set(ctx, "version:node:latest", versionData, 1*time.Hour)
```

### 2. Checksum Verification (Requirements 10.4, 10.5)

The Cache Manager verifies checksums before returning cached data:

```go
// Get with automatic checksum verification
data, err := cacheManager.GetWithChecksum(ctx, key, expectedChecksum)
if err != nil {
    // Checksum mismatch - corrupted entry was automatically deleted
    // Re-download the artifact
}
```

If a checksum mismatch is detected:

1. The corrupted entry is automatically deleted
2. An audit log entry is created
3. An error is returned to trigger re-download

### 3. Cache Purging (Requirement 10.6)

The Cache Manager supports multiple purging strategies:

```go
// Purge all cache entries
err := cacheManager.PurgeAll(ctx)

// Purge only expired entries
err := cacheManager.PurgeExpired(ctx)

// Purge tool-specific entries (by prefix)
err := cacheManager.PurgeByPrefix(ctx, "github:releases:node")
```

### 4. Cache Size Tracking and Automatic Cleanup (Requirement 10.7)

The Cache Manager tracks cache size and performs automatic cleanup:

```go
// Get current cache size
size, err := cacheManager.GetCacheSize()

// Automatic cleanup when size exceeds threshold
err := cacheManager.AutoCleanup(ctx)
```

The automatic cleanup process:

1. Calculates current cache size by walking the cache directory
2. If size exceeds the configured threshold (default 5GB):
   - Purges expired entries first
   - Logs a warning if still over threshold (LRU cleanup would be implemented here)

### 5. Cache Hit/Miss Tracking (Requirement 10.8)

The Cache Manager tracks cache hits and misses for performance monitoring:

```go
// Get cache statistics
stats := cacheManager.GetStats()
fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)

// Reset statistics
cacheManager.ResetStats()
```

Statistics are tracked in a thread-safe manner using `sync.RWMutex`.

## Configuration

The Cache Manager is configured using `CacheManagerConfig`:

```go
config := CacheManagerConfig{
    CacheDir:     "/path/to/cache",  // Required
    MaxCacheSize: 5 * 1024 * 1024 * 1024, // 5GB (default)
}

cacheManager, err := NewCacheManager(cacheRepo, auditRepo, config)
```

## Thread Safety

The Cache Manager is thread-safe:

- All operations use `sync.RWMutex` for concurrent access
- Read operations use `RLock()` for concurrent reads
- Write operations use `Lock()` for exclusive access
- Statistics tracking uses a separate mutex for fine-grained locking

## Error Handling

The Cache Manager follows the project's error handling conventions:

```go
// Wrap errors with context
if err := cm.repo.Set(ctx, key, value, ttl); err != nil {
    return fmt.Errorf("store cache entry in database: %w", err)
}

// Use sentinel errors for common cases
if errors.Is(err, repository.ErrNotFound) {
    // Handle cache miss
}
```

## Audit Logging

The Cache Manager logs important operations to the audit log:

- Checksum verification failures
- Cache purge operations (all, expired, by-prefix)
- Automatic cleanup operations

Example audit log entry:

```go
&repository.AuditEntry{
    Timestamp: time.Now(),
    Operation: "cache_checksum_failure",
    Status:    "failure",
    Error:     "checksum mismatch for key X: expected Y, got Z",
}
```

## Testing

The Cache Manager has comprehensive unit tests covering:

- Constructor validation
- Set/Get/Delete operations
- Checksum verification (success and failure cases)
- Purge operations (all, expired, by-prefix)
- Statistics tracking (hits, misses, reset)
- Cache size calculation
- Automatic cleanup (triggered and not triggered)

Run tests:

```bash
go test -v ./internal/service -run TestCacheManager
go test -v ./internal/service -run TestNewCacheManager
```

## Future Enhancements

### LRU Eviction

Currently, when the cache exceeds the size threshold after purging expired entries, only a warning is logged. A future enhancement would implement Least Recently Used (LRU) eviction:

1. Track access time for each cache entry
2. When cleanup is needed, remove least recently used entries
3. Continue until cache size is below threshold

### Compression

Large cache entries could be compressed to save disk space:

```go
// Compress before storing
compressed := compress(value)
err := cm.Set(ctx, key, compressed, ttl)

// Decompress after retrieving
compressed, err := cm.Get(ctx, key)
value := decompress(compressed)
```

### Distributed Caching

For multi-node deployments, the cache could be backed by Redis or Memcached:

```go
type DistributedCacheRepository struct {
    redis *redis.Client
}

func (r *DistributedCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return r.redis.Set(ctx, key, value, ttl).Err()
}
```

## Requirements Validation

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| 10.1 | `Set()` stores downloaded tarballs in cache directory | ✅ |
| 10.2 | `Set()` stores metadata with configurable TTL | ✅ |
| 10.3 | `Set()` stores version resolution results with TTL | ✅ |
| 10.4 | `GetWithChecksum()` verifies checksum before use | ✅ |
| 10.5 | `GetWithChecksum()` deletes corrupted entries | ✅ |
| 10.6 | `PurgeAll()`, `PurgeExpired()`, `PurgeByPrefix()` | ✅ |
| 10.7 | `GetCacheSize()`, `AutoCleanup()` track and cleanup | ✅ |
| 10.8 | `GetStats()` records hits and misses | ✅ |

## Related Components

- **CacheRepository** (`internal/repository/repository.go`): Interface for cache persistence
- **SQLiteCacheRepository** (`internal/repository/sqlite/cache.go`): SQLite implementation
- **AuditRepository** (`internal/repository/repository.go`): Interface for audit logging
- **TransactionManager** (`internal/transaction/transaction.go`): Transaction support

## Example Usage

```go
package main

import (
    "context"
    "time"

    "github.com/snowdreamtech/unirtm/internal/service"
    "github.com/snowdreamtech/unirtm/internal/repository/sqlite"
)

func main() {
    // Initialize repositories
    db, _ := sqlite.Open("unirtm.db")
    cacheRepo := sqlite.NewCacheRepository(db)
    auditRepo := sqlite.NewAuditRepository(db)

    // Create cache manager
    config := service.CacheManagerConfig{
        CacheDir:     "/var/cache/unirtm",
        MaxCacheSize: 5 * 1024 * 1024 * 1024, // 5GB
    }

    cacheManager, _ := service.NewCacheManager(cacheRepo, auditRepo, config)

    ctx := context.Background()

    // Store metadata with 24-hour TTL
    metadata := []byte(`{"version": "20.0.0", "url": "..."}`)
    _ = cacheManager.Set(ctx, "github:releases:node", metadata, 24*time.Hour)

    // Retrieve with checksum verification
    expectedChecksum := "abc123..."
    data, err := cacheManager.GetWithChecksum(ctx, "github:releases:node", expectedChecksum)
    if err != nil {
        // Handle cache miss or checksum mismatch
    }

    // Get statistics
    stats := cacheManager.GetStats()
    println("Cache hit rate:", float64(stats.Hits)/float64(stats.Hits+stats.Misses))

    // Automatic cleanup
    _ = cacheManager.AutoCleanup(ctx)
}
```
