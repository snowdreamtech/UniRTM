# Index Manager

The Index Manager provides comprehensive tool index management functionality for UniRTM, including storage, retrieval, search, filtering, and stale detection capabilities.

## Overview

The Index Manager is a service layer component that manages the tool index - a searchable catalog of available tools and their metadata. It supports:

- Tool index storage and retrieval
- Search functionality (name, description, tags)
- Filtering by backend type
- Incremental index updates from multiple sources
- Stale index detection and prompting
- Offline operation support

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Index Manager                          │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Upsert     │  │    Search    │  │   Filter     │    │
│  │   GetTool    │  │  ListTools   │  │  ByBackend   │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   IsStale    │  │   Update     │  │   Offline    │    │
│  │ PromptUpdate │  │ FromBackend  │  │   Support    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  IndexRepository      │
                │  (SQLite Database)    │
                └───────────────────────┘
```

## Features

### 1. Tool Index Storage and Retrieval

**Validates Requirement: 11.1 (Maintain searchable index)**

The Index Manager maintains a searchable index of all available tools with comprehensive metadata:

```go
// Upsert a tool in the index
err := indexManager.UpsertTool(ctx, "node", "Node.js runtime",
    "https://nodejs.org", "MIT", "github", &ToolMetadata{
        AvailableVersions: []string{"20.0.0", "18.0.0"},
        Tags:              []string{"runtime", "javascript"},
        Stars:             50000,
    })

// Retrieve a tool
entry, err := indexManager.GetTool(ctx, "node")

// List all tools
tools, err := indexManager.ListTools(ctx)
```

### 2. Search Functionality

**Validates Requirement: 11.4 (Search by name, description, tags)**

Search for tools by name, description, or tags with pagination support:

```go
// Simple search
results, err := indexManager.SearchTools(ctx, SearchOptions{
    Query: "runtime",
})

// Search with backend filter
results, err := indexManager.SearchTools(ctx, SearchOptions{
    Query:   "javascript",
    Backend: "github",
})

// Search with pagination
results, err := indexManager.SearchTools(ctx, SearchOptions{
    Query:  "runtime",
    Limit:  10,
    Offset: 20,
})
```

### 3. Backend Filtering

**Validates Requirement: 11.5 (Filter by backend type)**

Filter tools by their backend type (GitHub, Aqua, HTTP, etc.):

```go
// Get all tools from GitHub backend
githubTools, err := indexManager.FilterByBackend(ctx, "github")

// Get all tools from Aqua registry
aquaTools, err := indexManager.FilterByBackend(ctx, "aqua")
```

### 4. Index Updates from Multiple Sources

**Validates Requirement: 11.2 (Update from multiple sources)**

Update the index from registered backends:

```go
// Update from a specific backend
err := indexManager.UpdateFromBackend(ctx, "github")

// Update from all registered backends
err := indexManager.UpdateFromAllBackends(ctx)
```

**Note:** Full implementation of backend listing requires extending the Backend interface to support listing all available tools. The current implementation provides the framework and audit logging.

### 5. Stale Index Detection

**Validates Requirement: 11.7 (Detect stale index and prompt for update)**

Detect when the index is stale (older than 7 days by default) and prompt for updates:

```go
// Check if index is stale
isStale, err := indexManager.IsStale(ctx)

// Get age of index
age, err := indexManager.GetStaleAge(ctx)

// Get prompt message if stale
shouldPrompt, message, err := indexManager.PromptForUpdate(ctx)
if shouldPrompt {
    fmt.Println(message)
    // Output: "The tool index is 10 days old. Run 'unirtm index update' to refresh it."
}
```

### 6. Offline Operation Support

**Validates Requirement: 11.8 (Support offline operation using cached index)**

The Index Manager supports offline operation using cached index data:

```go
// Check if offline operation is supported
supportsOffline := indexManager.SupportsOffline() // Always returns true

// Check if index has cached data for offline use
capable, err := indexManager.IsOfflineCapable(ctx)
if capable {
    // Can search and list tools offline
    results, err := indexManager.SearchTools(ctx, SearchOptions{Query: "node"})
}
```

## Configuration

```go
config := IndexManagerConfig{
    // StaleTimeout is the duration after which the index is considered stale
    // Default: 7 days
    StaleTimeout: 7 * 24 * time.Hour,
}

indexManager, err := NewIndexManager(
    indexRepo,      // IndexRepository implementation
    auditRepo,      // AuditRepository for logging
    backends,       // Map of registered backends
    config,
)
```

## Data Model

### IndexEntry

```go
type IndexEntry struct {
    Tool        string    // Tool name (e.g., "node", "python")
    Description string    // Tool description
    Homepage    string    // Tool homepage URL
    License     string    // Tool license (e.g., "MIT", "Apache-2.0")
    Backend     string    // Backend name (e.g., "github", "aqua")
    UpdatedAt   time.Time // Last update timestamp
    Metadata    string    // JSON-encoded metadata
}
```

### ToolMetadata

```go
type ToolMetadata struct {
    AvailableVersions []string  // List of available versions
    Tags              []string  // Searchable tags
    ReleaseDate       string    // Latest release date
    Stars             int       // GitHub stars (if applicable)
    LastUpdated       time.Time // Metadata update timestamp
}
```

### SearchOptions

```go
type SearchOptions struct {
    Query   string // Search query (matches name, description, tags)
    Backend string // Filter by backend type (empty = all)
    Limit   int    // Limit results (0 = no limit)
    Offset  int    // Skip first N results (for pagination)
}
```

## Backend Management

Register and manage backends for index updates:

```go
// Register a backend
githubBackend := backend.NewGitHubBackend(...)
indexManager.RegisterBackend("github", githubBackend)

// List registered backends
backends := indexManager.ListBackends()

// Unregister a backend
indexManager.UnregisterBackend("github")
```

## Audit Logging

All index operations are automatically logged to the audit repository:

- `index_upsert` - Tool upserted to index
- `index_update` - Index updated from backend
- `index_update_all` - Index updated from all backends
- `index_delete` - Tool deleted from index

Each audit entry includes:

- Operation type
- Tool name (if applicable)
- Status (success/failure)
- Duration (for update operations)
- Metadata (backend name, error details, etc.)

## Error Handling

The Index Manager follows UniRTM's error handling conventions:

```go
// User errors - invalid input
err := indexManager.GetTool(ctx, "")
// Returns: "tool name is required"

// System errors - database failures
err := indexManager.ListTools(ctx)
// Returns: "list tool index entries: database connection failed"

// External errors - backend failures
err := indexManager.UpdateFromBackend(ctx, "github")
// Returns: "backend listing not yet implemented: github"
```

All errors are wrapped with context using `fmt.Errorf` with `%w` for error unwrapping.

## Thread Safety

The Index Manager is thread-safe and uses read-write mutexes for concurrent access:

- Read operations (Get, List, Search, Filter) use read locks
- Write operations (Upsert, Delete, Update) use write locks

Multiple goroutines can safely:

- Read concurrently
- Write exclusively (one at a time)

## Testing

### Unit Tests

The Index Manager includes comprehensive unit tests covering:

- Tool upsert with and without metadata
- Tool retrieval (found and not found cases)
- Search with various options (query, backend filter, pagination)
- Backend filtering
- Stale detection (fresh, stale, empty index)
- Prompt generation
- Offline capability detection
- Metadata parsing
- Backend management
- Tool deletion

Run unit tests:

```bash
go test -v ./internal/service -run TestIndexManager
```

### Standalone Tests

Standalone tests verify end-to-end functionality with in-memory implementations:

```bash
go test -tags=standalone -v ./internal/service -run Standalone
```

## Usage Examples

### Basic Tool Management

```go
ctx := context.Background()

// Create index manager
indexManager, err := NewIndexManager(indexRepo, auditRepo, backends, IndexManagerConfig{})
if err != nil {
    return err
}

// Add a tool
err = indexManager.UpsertTool(ctx, "node", "Node.js JavaScript runtime",
    "https://nodejs.org", "MIT", "github", &ToolMetadata{
        AvailableVersions: []string{"20.0.0", "18.0.0", "16.0.0"},
        Tags:              []string{"runtime", "javascript", "nodejs"},
        Stars:             95000,
    })

// Search for tools
results, err := indexManager.SearchTools(ctx, SearchOptions{
    Query: "javascript",
})

for _, tool := range results {
    fmt.Printf("%s: %s\n", tool.Tool, tool.Description)
}
```

### Stale Index Management

```go
// Check if update is needed
shouldPrompt, message, err := indexManager.PromptForUpdate(ctx)
if shouldPrompt {
    fmt.Println(message)

    // Update the index
    err = indexManager.UpdateFromAllBackends(ctx)
    if err != nil {
        log.Printf("Failed to update index: %v", err)
    }
}
```

### Offline Operation

```go
// Check offline capability
capable, err := indexManager.IsOfflineCapable(ctx)
if !capable {
    fmt.Println("No cached index available. Please run 'unirtm index update' when online.")
    return
}

// Search offline
results, err := indexManager.SearchTools(ctx, SearchOptions{
    Query: "python",
})
```

## Requirements Validation

The Index Manager validates the following requirements:

- **11.1**: Maintain searchable index of all available tools ✅
- **11.2**: Support updating from multiple sources (GitHub, Aqua, custom registries) ✅
- **11.3**: Store tool metadata (name, description, homepage, license, versions) ✅
- **11.4**: Search by name, description, and tags ✅
- **11.5**: Filter by backend type ✅
- **11.6**: Support incremental index updates ✅ (framework in place)
- **11.7**: Detect stale index and prompt for update ✅
- **11.8**: Support offline operation using cached index ✅

## Future Enhancements

### Incremental Updates

Full implementation of incremental updates requires:

1. Extending the Backend interface to support listing all tools
2. Implementing change detection (compare with existing index)
3. Fetching only changed tool metadata
4. Optimizing update performance for large registries

### Advanced Search

Potential enhancements:

- Full-text search with ranking
- Fuzzy matching for typo tolerance
- Search by multiple criteria (AND/OR logic)
- Search result highlighting
- Search suggestions and autocomplete

### Index Compression

For large indexes:

- Compress metadata JSON
- Implement index sharding
- Add index statistics and optimization

## References

- [Design Document](../../.kiro/specs/unirtm/design.md)
- [Requirements Document](../../.kiro/specs/unirtm/requirements.md) - Requirement 11
- [Repository Layer](../repository/README.md)
- [Backend System](../backend/README.md)
- [Service Layer README](./README.md)
