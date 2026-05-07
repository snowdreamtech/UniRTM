# UniRTM Logger

Enhanced zerolog-based logging system for UniRTM with rotating file writers, structured logging, and automatic stack trace capture.

## Features

- **Multiple Log Levels**: Trace, Debug, Info, Warn, Error, Fatal, Panic
- **Rotating File Writers**: Automatic log rotation for both operation and error logs
  - Max size: 500MB per file
  - Max backups: 10 files
  - Max age: 30 days
  - Automatic compression of rotated files
- **Structured Logging**: Support for context fields and key-value pairs
- **Stack Trace Capture**: Automatic stack trace capture for error-level logs
- **Console Output**: Color-coded log levels for better readability
- **Dual Output**: Logs written to both files and console (stdout/stderr)

## Configuration

### Log Files

- **unirtm.log**: All operation logs (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- **error.log**: Error-level logs only (Error, Fatal, Panic) with stack traces

### Rotation Settings

```go
MaxSize:    500  // megabytes
MaxBackups: 10   // number of old log files to retain
MaxAge:     30   // days
Compress:   true // compress rotated files
```

## Usage

### Initialization

```go
import "github.com/snowdreamtech/unirtm/internal/pkg/logger"

// Initialize with default paths (unirtm.log and error.log)
opWriter, errWriter := logger.InitUniRTMLogger("", "")

// Initialize with custom paths
opWriter, errWriter := logger.InitUniRTMLogger("/var/log/unirtm/unirtm.log", "/var/log/unirtm/error.log")
```

### Basic Logging

```go
// Simple log messages
logger.Trace("trace message")
logger.Debug("debug message")
logger.Info("info message")
logger.Warn("warning message")
logger.Error("error message")  // Automatically captures stack trace
logger.Fatal("fatal message")  // Logs and exits
logger.Panic("panic message")  // Logs and panics
```

### Structured Logging with Context

```go
// Log with context fields
logger.Info("Tool installed", map[string]interface{}{
    "tool":    "node",
    "version": "20.0.0",
    "path":    "/usr/local/bin/node",
})

logger.Error("Installation failed", map[string]interface{}{
    "tool":    "python",
    "version": "3.11.0",
    "reason":  "checksum mismatch",
})
```

### Logging Errors with Error Objects

```go
err := errors.New("network timeout")

// Log error with message
logger.ErrorWithErr(err, "Failed to download artifact")

// Log error with context fields
logger.ErrorWithErr(err, "Download failed", map[string]interface{}{
    "url":     "https://example.com/artifact.tar.gz",
    "retries": 5,
    "timeout": "30s",
})
```

### Creating Loggers with Context

```go
// Create a logger with context fields
contextLogger := logger.WithContext(map[string]interface{}{
    "request_id": "req-123456",
    "user_id":    "user-789",
    "operation":  "install",
})

// All logs from this logger will include the context fields
contextLogger.Info().Msg("Starting installation")
contextLogger.Debug().Msg("Downloading artifact")
contextLogger.Info().Msg("Installation completed")
```

### Creating Loggers with Variadic Fields

```go
// Create a logger with key-value pairs
fieldsLogger := logger.WithFields(
    "tool", "go",
    "version", "1.21.0",
    "backend", "github",
)

// All logs from this logger will include these fields
fieldsLogger.Info().Msg("Tool resolved")
fieldsLogger.Debug().Msg("Downloading from GitHub")
fieldsLogger.Info().Msg("Installation successful")
```

### Using the Global Logger Directly

```go
// Access the global logger for advanced usage
logger.Logger.Info().
    Str("tool", "node").
    Str("version", "20.0.0").
    Int("size_mb", 45).
    Msg("Download completed")

logger.Logger.Error().
    Err(err).
    Str("operation", "install").
    Msg("Operation failed")
```

## Integration with Error Handling

The logger integrates seamlessly with the UniRTM error handling system:

```go
import (
    "github.com/snowdreamtech/unirtm/internal/pkg/errors"
    "github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// Create a categorized error
err := errors.NewUserError("invalid version specification", nil)

// Log the error with its category
logger.ErrorWithErr(err, "Configuration validation failed", map[string]interface{}{
    "category": errors.GetCategory(err).String(),
    "file":     ".unirtm.toml",
})
```

## Log Output Format

### Console Output (Human-Readable)

```
2026-05-06T13:00:00+08:00 INF logger.go:456 > Tool installed tool=node version=20.0.0 path=/usr/local/bin/node
2026-05-06T13:00:01+08:00 ERR logger.go:477 > Installation failed tool=python version=3.11.0 reason="checksum mismatch" stack_trace="goroutine 1 [running]:\n..."
```

### File Output (JSON)

```json
{
  "level": "info",
  "time": "2026-05-06T13:00:00+08:00",
  "caller": "/path/to/file.go:123",
  "message": "Tool installed",
  "tool": "node",
  "version": "20.0.0",
  "path": "/usr/local/bin/node"
}

{
  "level": "error",
  "time": "2026-05-06T13:00:01+08:00",
  "caller": "/path/to/file.go:456",
  "message": "Installation failed",
  "tool": "python",
  "version": "3.11.0",
  "reason": "checksum mismatch",
  "stack_trace": "goroutine 1 [running]:\n..."
}
```

## Requirements Validation

This implementation satisfies the following UniRTM requirements:

- **Requirement 7.1**: Multiple log levels (Trace, Debug, Info, Warn, Error, Fatal, Panic)
- **Requirement 7.2**: Rotating file writers for unirtm.log (max 500MB, 10 backups, 30 days retention)
- **Requirement 7.3**: Rotating file writers for error.log (max 500MB, 10 backups, 30 days retention)
- **Requirement 7.4**: Structured logging with context fields
- **Requirement 7.5**: Timestamps, log levels, and structured context in all log entries
- **Requirement 7.7**: Stack trace capture for errors
- **Requirement 7.8**: Integration with audit logging system (via structured fields)

## Testing

Run the test suite:

```bash
go test -v ./internal/pkg/logger/...
```

Run specific tests:

```bash
go test -v -run TestInitUniRTMLogger ./internal/pkg/logger/...
go test -v -run TestLogLevels ./internal/pkg/logger/...
go test -v -run TestErrorWithErr ./internal/pkg/logger/...
```

## Performance Considerations

- **Buffered I/O**: Lumberjack uses buffered I/O for efficient file writes
- **Lazy File Creation**: Log files are created only when first written to
- **Compression**: Rotated log files are automatically compressed to save disk space
- **Minimal Overhead**: Zerolog is one of the fastest structured logging libraries for Go
- **Stack Trace Caching**: Stack traces are captured only for error-level logs

## Best Practices

1. **Initialize Early**: Call `InitUniRTMLogger()` at application startup
2. **Use Structured Logging**: Always include relevant context fields
3. **Avoid Sensitive Data**: Never log passwords, tokens, or PII
4. **Use Appropriate Levels**:
   - Trace: Very detailed debugging information
   - Debug: Debugging information
   - Info: General informational messages
   - Warn: Warning messages for potentially harmful situations
   - Error: Error messages for failures that don't stop the application
   - Fatal: Critical errors that cause application exit
   - Panic: Critical errors that cause panic
5. **Context Propagation**: Use `WithContext()` or `WithFields()` to create loggers with persistent context
6. **Error Wrapping**: Use `ErrorWithErr()` to log errors with their full context

## Migration from Existing Logger

The enhanced logger maintains backward compatibility with the existing `InitLogger()` function for Gin integration. New code should use `InitUniRTMLogger()` for full UniRTM functionality.

```go
// Old (Gin-specific)
errorWriter, ginWriter := logger.InitLogger("error.log", "gin.log")

// New (UniRTM-specific)
opWriter, errWriter := logger.InitUniRTMLogger("unirtm.log", "error.log")
```
