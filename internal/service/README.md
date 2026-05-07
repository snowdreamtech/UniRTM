# Service Layer

The service layer provides high-level business logic and orchestrates operations across repositories, backends, and providers. It implements the core functionality of UniRTM while maintaining clean separation from infrastructure concerns.

## Overview

The service layer sits between the CLI/API layer and the repository/backend layers. It:

- Implements business logic and workflows
- Orchestrates operations across multiple repositories
- Provides transaction management
- Handles error classification and logging
- Enforces business rules and validation

## Architecture

```
CLI/API Layer
     ↓
Service Layer (Business Logic)
     ↓
Repository Layer (Data Access)
     ↓
Database Layer (SQLite)
```

## Components

### Audit Service

The `AuditService` provides comprehensive audit logging functionality for all system operations.

**Features:**
- High-level API for creating audit log entries
- Convenience methods for common operations (install, uninstall, activate, etc.)
- Integration with the logger for immediate visibility
- Support for custom metadata (JSON-encoded)
- Query capabilities with filters (time range, operation type, tool, status)

**Usage Example:**

```go
// Initialize database and repository
db, err := database.Open(ctx, database.Config{
    Path:    "/path/to/unirtm.db",
    WALMode: true,
})
if err != nil {
    return err
}
defer db.Close()

auditRepo, err := sqlite.NewAuditRepository(db.Conn())
if err != nil {
    return err
}
defer auditRepo.Close()

// Create audit service
auditService := service.NewAuditService(auditRepo)

// Log a tool installation
err = auditService.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
if err != nil {
    return err
}

// Log a failed operation
installErr := fmt.Errorf("download failed: connection timeout")
err = auditService.LogInstall(ctx, "python", "3.11.0", 5*time.Second, installErr)
if err != nil {
    return err
}

// Query recent logs
logs, err := auditService.GetRecentLogs(ctx, 10)
if err != nil {
    return err
}

// Query logs by operation type
installLogs, err := auditService.GetLogsByOperation(ctx, service.OperationInstall, 20)
if err != nil {
    return err
}

// Query logs by tool
nodeLogs, err := auditService.GetLogsByTool(ctx, "node", 15)
if err != nil {
    return err
}

// Query failed operations
failedLogs, err := auditService.GetLogsByStatus(ctx, service.StatusFailure, 25)
if err != nil {
    return err
}
```

**Custom Metadata:**

```go
entry := &service.AuditLogEntry{
    Operation: service.OperationInstall,
    Tool:      "node",
    Version:   "20.0.0",
    Status:    service.StatusSuccess,
    Duration:  2500,
    Metadata: map[string]interface{}{
        "backend":      "github",
        "download_url": "https://github.com/nodejs/node/releases/download/v20.0.0/node-v20.0.0.tar.gz",
        "size_bytes":   45678901,
        "checksum":     "abc123def456",
    },
}

err := auditService.LogOperation(ctx, entry)
```

**Operation Types:**

- `OperationInstall` - Tool installation
- `OperationUninstall` - Tool uninstallation
- `OperationActivate` - Tool activation
- `OperationDeactivate` - Tool deactivation
- `OperationUpdate` - Tool update
- `OperationCachePurge` - Cache purge
- `OperationConfigLoad` - Configuration load
- `OperationConfigUpdate` - Configuration update
- `OperationIndexUpdate` - Index update
- `OperationVersionResolve` - Version resolution

**Operation Status:**

- `StatusSuccess` - Operation completed successfully
- `StatusFailure` - Operation failed

### Auto-Activation Manager

The `AutoActivationManager` provides automatic environment switching based on directory context.

**Features:**
- Automatic detection of project configuration files
- Shell hook generation for directory change events
- Seamless environment switching when entering/leaving projects
- Support for nested project configurations
- Integration with shell initialization scripts

**Usage Example:**

```go
// Create auto-activation manager
manager := service.NewAutoActivationManager(
    "/usr/local/unirtm/shims",
    "/var/lib/unirtm",
    configManager,
)

// Generate shell hook for bash
hook, err := manager.GenerateShellHook(ctx, service.ShellBash)
if err != nil {
    return err
}

// Add to shell initialization
fmt.Println("Add this to your ~/.bashrc:")
fmt.Println(hook.Content)
```

### Index Manager

The `IndexManager` manages tool index storage, retrieval, search, and updates.

**Features:**
- Tool index storage and retrieval
- Search functionality (name, description, tags)
- Filtering by backend type
- Incremental index updates from multiple sources
- Stale index detection and prompting
- Offline operation support

**Usage Example:**

```go
// Create index manager
indexManager, err := service.NewIndexManager(
    indexRepo,
    auditRepo,
    backends,
    service.IndexManagerConfig{
        StaleTimeout: 7 * 24 * time.Hour,
    },
)
if err != nil {
    return err
}

// Add a tool to the index
err = indexManager.UpsertTool(ctx, "node", "Node.js runtime",
    "https://nodejs.org", "MIT", "github", &service.ToolMetadata{
        AvailableVersions: []string{"20.0.0", "18.0.0"},
        Tags:              []string{"runtime", "javascript"},
    })

// Search for tools
results, err := indexManager.SearchTools(ctx, service.SearchOptions{
    Query: "javascript",
})

// Check if index is stale
shouldPrompt, message, err := indexManager.PromptForUpdate(ctx)
if shouldPrompt {
    fmt.Println(message)
}

// Filter by backend
githubTools, err := indexManager.FilterByBackend(ctx, "github")
```

**Validates Requirements:**
- **11.1**: Maintain searchable index
- **11.2**: Update from multiple sources
- **11.3**: Store tool metadata
- **11.4**: Search by name, description, tags
- **11.5**: Filter by backend type
- **11.6**: Incremental index updates
- **11.7**: Stale detection and prompting
- **11.8**: Offline operation support

## Design Principles

**Features:**
- Shell-specific script generation (bash, zsh, fish, PowerShell)
- PATH modification to include shims directory
- Environment variable setting for active tool versions
- Project-specific activation support
- Global activation support
- Automatic shell detection

**Usage Example:**

```go
// Create activation manager
manager := service.NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm")
ctx := context.Background()

// Generate global activation for bash
toolVersions := map[string]string{
    "node":   "20.0.0",
    "python": "3.11.0",
    "go":     "1.21.0",
}

script, err := manager.GenerateGlobalActivation(ctx, service.ShellBash, toolVersions)
if err != nil {
    return err
}

// Write script to file
err = os.WriteFile("activate.sh", []byte(script.Content), 0644)
if err != nil {
    return err
}

// Display instructions
fmt.Println(script.Instructions)

// Generate project-specific activation
envVars := map[string]string{
    "NODE_ENV": "development",
    "DEBUG":    "app:*",
}

projectScript, err := manager.GenerateProjectActivation(
    ctx,
    service.ShellBash,
    "/home/user/myproject",
    map[string]string{"node": "18.0.0"},
    envVars,
)
if err != nil {
    return err
}

// Detect current shell
shell, err := service.DetectShell()
if err != nil {
    return err
}

// Generate activation for detected shell
autoScript, err := manager.GenerateGlobalActivation(ctx, shell, toolVersions)
if err != nil {
    return err
}
```

**Shell Types:**

- `ShellBash` - Bash shell
- `ShellZsh` - Zsh shell
- `ShellFish` - Fish shell
- `ShellPowerShell` - PowerShell

**Activation Scopes:**

- `ScopeGlobal` - System-wide default tool versions
- `ScopeProject` - Project-specific tool versions

## Design Principles

### 1. Single Responsibility

Each service has a single, well-defined purpose. The `AuditService` is responsible only for audit logging, not for performing the operations being audited.

### 2. Dependency Inversion

Services depend on repository interfaces, not concrete implementations. This allows for easy testing and swapping of implementations.

```go
type AuditService struct {
    repo repository.AuditRepository  // Interface, not concrete type
}
```

### 3. Error Handling

Services use the error handling system from `internal/pkg/errors` to classify errors:

- **User Errors**: Invalid input, configuration errors
- **System Errors**: Database failures, disk full
- **External Errors**: Network failures, backend API errors

### 4. Logging Integration

Services integrate with the logger from `internal/pkg/logger` to provide immediate visibility into operations:

```go
logger.Info("Operation completed", map[string]interface{}{
    "operation": entry.Operation,
    "tool":      entry.Tool,
    "version":   entry.Version,
    "status":    entry.Status,
    "duration":  entry.Duration,
})
```

### 5. Context Propagation

All service methods accept `context.Context` as the first parameter for:

- Cancellation support
- Timeout enforcement
- Request-scoped values (request ID, user ID, etc.)

### 6. Transaction Support

Services that perform multiple database operations should use the transaction manager from `internal/transaction` to ensure atomicity.

## Testing

### Unit Tests

Unit tests use mock repositories to test business logic in isolation:

```go
type MockAuditRepository struct {
    logFunc   func(ctx context.Context, entry *repository.AuditEntry) error
    queryFunc func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error)
}

func TestAuditService_LogInstall(t *testing.T) {
    repo := &MockAuditRepository{
        logFunc: func(ctx context.Context, entry *repository.AuditEntry) error {
            assert.Equal(t, "install", entry.Operation)
            assert.Equal(t, "node", entry.Tool)
            return nil
        },
    }

    service := NewAuditService(repo)
    err := service.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
    require.NoError(t, err)
}
```

### Integration Tests

Integration tests use real database connections to test the full stack:

```go
func TestAuditService_Integration(t *testing.T) {
    db, err := database.Open(ctx, database.Config{
        Path:    ":memory:",
        WALMode: false,
    })
    require.NoError(t, err)
    defer db.Close()

    auditRepo, err := sqlite.NewAuditRepository(db.Conn())
    require.NoError(t, err)
    defer auditRepo.Close()

    service := NewAuditService(auditRepo)

    // Test full workflow
    err = service.LogInstall(ctx, "node", "20.0.0", 2*time.Second, nil)
    require.NoError(t, err)

    logs, err := service.GetRecentLogs(ctx, 10)
    require.NoError(t, err)
    assert.Len(t, logs, 1)
}
```

## Future Services

The service layer will be expanded to include:

- **InstallationService**: Tool installation and management ✅ (Implemented)
- **ConfigService**: Configuration management and validation
- **CacheService**: Cache management and purging ✅ (Implemented)
- **IndexService**: Tool index management and search ✅ (Implemented)
- **VersionService**: Version resolution and management ✅ (Implemented)
- **BackendService**: Backend coordination and selection
- **ProviderService**: Provider management and delegation
- **ActivationService**: Environment activation management ✅ (Implemented)
- **AutoActivationService**: Automatic environment switching ✅ (Implemented)

## Requirements Validation

The audit service validates the following requirements:

- **Requirement 7.8**: Audit logging to database
- **Requirement 8.1**: Log all operations to audit log before execution
- **Requirement 8.5**: Record operation type, timestamp, user, affected tools, success/failure status, error messages

## References

- [Design Document](../../.kiro/specs/unirtm/design.md)
- [Requirements Document](../../.kiro/specs/unirtm/requirements.md)
- [Repository Layer](../repository/README.md)
- [Error Handling](../pkg/errors/errors.go)
- [Logger](../pkg/logger/logger.go)
