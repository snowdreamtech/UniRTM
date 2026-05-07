# Update Manager

## Overview

The Update Manager provides comprehensive tool update functionality for UniRTM, including version checking, single and bulk updates, version constraint respect, update preview, automatic updates, and rollback on failure.

**Validates Requirements:** 25.1, 25.2, 25.3, 25.4, 25.5, 25.6, 25.7

## Architecture

### Core Components

```
UpdateManager
├── CheckForUpdates()      - Check for newer versions of installed tools
├── UpdateTool()           - Update a specific tool to a specific version
├── UpdateAll()            - Update all tools to their latest versions
├── PreviewUpdates()       - Show preview of what will be updated
├── EnableAutomaticUpdates()  - Enable automatic updates (opt-in)
└── DisableAutomaticUpdates() - Disable automatic updates
```

### Dependencies

- **Backend Registry**: For version resolution and download information
- **Provider Registry**: For tool-specific installation logic
- **Download Manager**: For downloading tool artifacts
- **Installation Repository**: For tracking installed tools
- **Audit Repository**: For logging update operations
- **Transaction Manager**: For atomic updates with rollback
- **Config Manager**: For respecting version constraints

## Features

### 1. Version Checking (Requirement 25.1)

The Update Manager checks for newer versions of installed tools by:

1. Listing all installed tools from the database
2. Querying each tool's backend for the latest version
3. Comparing installed version with latest version
4. Respecting version constraints from configuration files

```go
updates, err := updateManager.CheckForUpdates(ctx)
for _, update := range updates {
    if update.UpdateRequired {
        fmt.Printf("%s: %s → %s\n",
            update.Tool,
            update.CurrentVersion,
            update.LatestVersion)
    }
}
```

### 2. Single Tool Update (Requirement 25.2)

Update a specific tool to a specific version with atomic transaction support:

```go
result, err := updateManager.UpdateTool(ctx, "node", "20.0.0")
if err != nil {
    log.Fatalf("Update failed: %v", err)
}

if result.Success {
    fmt.Printf("Updated %s from %s to %s in %v\n",
        result.Tool,
        result.OldVersion,
        result.NewVersion,
        result.Duration)
}
```

**Update Workflow:**

1. Verify tool is installed
2. Check if already at target version (skip if yes)
3. Start database transaction
4. Download new version
5. Verify checksum
6. Install new version
7. Update database record
8. Commit transaction
9. Remove old version (after successful commit)

### 3. Bulk Update (Requirement 25.3)

Update all tools to their latest versions:

```go
results, err := updateManager.UpdateAll(ctx)
for _, result := range results {
    if result.Success {
        fmt.Printf("✓ %s: %s → %s\n",
            result.Tool,
            result.OldVersion,
            result.NewVersion)
    } else {
        fmt.Printf("✗ %s: %s\n", result.Tool, result.Error)
    }
}
```

**Behavior:**

- Continues with other updates even if one fails
- Returns results for all attempted updates
- Each update is atomic (all-or-nothing)

### 4. Version Constraint Respect (Requirement 25.4)

The Update Manager respects version constraints defined in configuration files:

```toml
# unirtm.toml
[tools]
node = { version = "18.5.0" }  # Pin to specific version
python = { version = "^3.11" }  # Use semver range
go = { version = "latest" }     # Always use latest
```

When checking for updates:

- If a specific version is pinned, that version is used as the target
- If a range is specified, the latest version within that range is used
- If "latest" is specified, the absolute latest version is used

### 5. Update Preview (Requirement 25.5)

Preview what will be updated before applying changes:

```go
preview, err := updateManager.PreviewUpdates(ctx)

fmt.Printf("Updates available: %d\n", preview.TotalUpdates)
fmt.Printf("Estimated time: %v\n", preview.EstimatedTime)

for _, update := range preview.Updates {
    fmt.Printf("  %s: %s → %s\n",
        update.Tool,
        update.CurrentVersion,
        update.LatestVersion)
}
```

**Preview Information:**

- List of tools that will be updated
- Current and target versions
- Total number of updates
- Estimated time for all updates (30 seconds per tool)

### 6. Automatic Updates (Requirement 25.6)

Enable automatic updates with a configurable schedule:

```go
// Enable automatic updates (runs daily at 2 AM)
err := updateManager.EnableAutomaticUpdates(ctx, "0 2 * * *")

// Disable automatic updates
err := updateManager.DisableAutomaticUpdates(ctx)
```

**Implementation Notes:**

- The Update Manager stores the automatic update configuration
- Actual scheduling is handled by a separate scheduler service or cron job
- The scheduler calls `UpdateAll()` periodically based on the schedule
- Configuration is logged in the audit log for tracking

### 7. Rollback on Failure (Requirement 25.7)

The Update Manager automatically rolls back failed updates:

**Rollback Triggers:**

- Download failure
- Checksum verification failure
- Installation failure
- Post-installation failure
- Database update failure
- Transaction commit failure

**Rollback Process:**

1. Transaction is automatically rolled back (database changes reverted)
2. New installation directory is removed
3. Old installation remains intact
4. Error is logged with full context
5. UpdateResult indicates rollback occurred

```go
result, err := updateManager.UpdateTool(ctx, "node", "20.0.0")
if err != nil {
    if result.RolledBack {
        fmt.Printf("Update failed and was rolled back: %s\n",
            result.RollbackReason)
    }
}
```

## Data Models

### UpdateInfo

```go
type UpdateInfo struct {
    Tool           string // Tool name
    CurrentVersion string // Currently installed version
    LatestVersion  string // Latest available version
    Backend        string // Backend used for the tool
    UpdateRequired bool   // Whether an update is available
}
```

### UpdatePreview

```go
type UpdatePreview struct {
    Updates       []UpdateInfo  // List of tools that will be updated
    TotalUpdates  int           // Total number of updates
    EstimatedTime time.Duration // Estimated time for all updates
}
```

### UpdateResult

```go
type UpdateResult struct {
    Tool           string        // Tool name
    OldVersion     string        // Previous version
    NewVersion     string        // New version after update
    Success        bool          // Whether the update succeeded
    Error          string        // Error message if update failed
    Duration       time.Duration // Time taken for the update
    RolledBack     bool          // Whether a rollback was performed
    RollbackReason string        // Reason for rollback if applicable
}
```

## Error Handling

The Update Manager provides comprehensive error handling:

### Error Types

1. **Tool Not Installed**: Returned when trying to update a tool that isn't installed
2. **Backend Not Found**: Returned when the tool's backend is unavailable
3. **Version Not Found**: Returned when the target version doesn't exist
4. **Download Failure**: Network errors, timeouts, connection issues
5. **Checksum Mismatch**: Downloaded file doesn't match expected checksum
6. **Installation Failure**: Provider-specific installation errors
7. **Database Failure**: Transaction or database operation errors

### Error Context

All errors include:

- Operation being performed
- Tool name and versions involved
- Underlying error cause
- Suggested resolution (where applicable)

## Audit Logging

All update operations are logged to the audit log:

```go
// Audit entry for update operation
{
    Timestamp: "2024-01-15T10:30:00Z",
    Operation: "update",
    Tool:      "node",
    Version:   "20.0.0",
    Status:    "success",
    Duration:  45000,  // milliseconds
    Metadata:  `{"old_version":"18.0.0","new_version":"20.0.0"}`
}
```

**Logged Operations:**

- `update` - Single tool update
- `update_all` - Bulk update operation
- `enable_automatic_updates` - Automatic updates enabled
- `disable_automatic_updates` - Automatic updates disabled

## Testing

The Update Manager includes comprehensive unit tests covering:

### Test Coverage

1. **CheckForUpdates**
   - Single tool with update available
   - Tool already at latest version
   - Respect version constraints from config

2. **UpdateTool**
   - Successful update
   - Already at target version (no-op)
   - Download failure triggers rollback
   - Checksum verification failure

3. **UpdateAll**
   - Multiple tools updated successfully
   - Continue on individual failures

4. **PreviewUpdates**
   - Correct update count and information
   - Estimated time calculation

5. **Automatic Updates**
   - Enable automatic updates
   - Disable automatic updates
   - Configuration logging

### Running Tests

```bash
# Run all Update Manager tests
go test -v ./internal/service -run TestUpdateManager

# Run specific test
go test -v ./internal/service -run TestUpdateManager_UpdateTool

# Run with race detector
go test -race ./internal/service -run TestUpdateManager
```

## Usage Examples

### Example 1: Check and Apply Updates

```go
// Check for updates
updates, err := updateManager.CheckForUpdates(ctx)
if err != nil {
    log.Fatalf("Failed to check for updates: %v", err)
}

// Show available updates
fmt.Printf("Found %d updates:\n", len(updates))
for _, update := range updates {
    if update.UpdateRequired {
        fmt.Printf("  %s: %s → %s\n",
            update.Tool,
            update.CurrentVersion,
            update.LatestVersion)
    }
}

// Apply updates
results, err := updateManager.UpdateAll(ctx)
if err != nil {
    log.Fatalf("Failed to update tools: %v", err)
}

// Report results
for _, result := range results {
    if result.Success {
        fmt.Printf("✓ %s updated in %v\n", result.Tool, result.Duration)
    } else {
        fmt.Printf("✗ %s failed: %s\n", result.Tool, result.Error)
    }
}
```

### Example 2: Preview Before Updating

```go
// Preview updates
preview, err := updateManager.PreviewUpdates(ctx)
if err != nil {
    log.Fatalf("Failed to preview updates: %v", err)
}

// Show preview
fmt.Printf("Updates available: %d\n", preview.TotalUpdates)
fmt.Printf("Estimated time: %v\n", preview.EstimatedTime)
fmt.Println("\nTools to be updated:")
for _, update := range preview.Updates {
    fmt.Printf("  %s: %s → %s\n",
        update.Tool,
        update.CurrentVersion,
        update.LatestVersion)
}

// Confirm with user
fmt.Print("\nProceed with updates? (y/n): ")
var response string
fmt.Scanln(&response)

if response == "y" {
    results, err := updateManager.UpdateAll(ctx)
    // ... handle results
}
```

### Example 3: Update Specific Tool

```go
// Update Node.js to version 20.0.0
result, err := updateManager.UpdateTool(ctx, "node", "20.0.0")
if err != nil {
    log.Fatalf("Update failed: %v", err)
}

if result.Success {
    fmt.Printf("Successfully updated %s from %s to %s\n",
        result.Tool,
        result.OldVersion,
        result.NewVersion)
    fmt.Printf("Update took %v\n", result.Duration)
} else {
    fmt.Printf("Update failed: %s\n", result.Error)
    if result.RolledBack {
        fmt.Printf("Changes were rolled back: %s\n", result.RollbackReason)
    }
}
```

## Integration

### CLI Integration

The Update Manager is designed to be integrated with the CLI:

```go
// cmd/update.go
func updateCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "update [tool] [version]",
        Short: "Update tools to newer versions",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()

            if len(args) == 0 {
                // Update all tools
                return updateAll(ctx)
            } else if len(args) == 2 {
                // Update specific tool to specific version
                return updateTool(ctx, args[0], args[1])
            } else {
                return fmt.Errorf("invalid arguments")
            }
        },
    }

    cmd.Flags().Bool("preview", false, "Preview updates without applying")
    cmd.Flags().Bool("auto", false, "Enable automatic updates")

    return cmd
}
```

### Scheduler Integration

For automatic updates, integrate with a scheduler:

```go
// scheduler/scheduler.go
func (s *Scheduler) scheduleAutomaticUpdates(schedule string) error {
    // Parse cron schedule
    cronSchedule, err := cron.Parse(schedule)
    if err != nil {
        return fmt.Errorf("invalid schedule: %w", err)
    }

    // Schedule update job
    s.cron.Schedule(cronSchedule, cron.FuncJob(func() {
        ctx := context.Background()
        results, err := s.updateManager.UpdateAll(ctx)
        if err != nil {
            log.Printf("Automatic update failed: %v", err)
            return
        }

        // Log results
        for _, result := range results {
            if result.Success {
                log.Printf("Auto-updated %s to %s", result.Tool, result.NewVersion)
            } else {
                log.Printf("Failed to auto-update %s: %s", result.Tool, result.Error)
            }
        }
    }))

    return nil
}
```

## Future Enhancements

Potential future enhancements for the Update Manager:

1. **Parallel Updates**: Update multiple tools concurrently for faster bulk updates
2. **Update Notifications**: Notify users when updates are available
3. **Update History**: Track update history with before/after snapshots
4. **Selective Updates**: Update only specific categories of tools (e.g., runtimes, CLI tools)
5. **Update Policies**: Define update policies (e.g., only security updates, only minor versions)
6. **Rollback Command**: Manual rollback to previous version
7. **Update Channels**: Support for stable, beta, and nightly update channels
8. **Bandwidth Throttling**: Limit download bandwidth for updates
9. **Update Verification**: Run tests after updates to verify functionality
10. **Update Scheduling**: More sophisticated scheduling options (maintenance windows, etc.)
