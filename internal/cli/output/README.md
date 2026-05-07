# CLI Output Package

This package provides comprehensive CLI output formatting capabilities for UniRTM, including progress indicators, JSON output, and color-coded human-readable output.

## Features

- **Multiple Output Formats**: Human-readable (with colors) and JSON (for scripting)
- **Progress Indicators**: Visual progress bars for long-running operations
- **Color Support**: Automatic color detection with NO_COLOR support
- **Structured Output**: Tables, key-value pairs, and structured data
- **Global Formatter**: Thread-safe global formatter for consistent output

## Usage

### Basic Output

```go
import "github.com/snowdreamtech/unirtm/internal/cli/output"

// Using global formatter
output.Info("Starting installation")
output.Success("Installation completed")
output.Warning("Configuration file not found, using defaults")
output.Error("Failed to connect to backend")

// With structured fields
output.Info("Tool installed", map[string]interface{}{
    "tool":    "node",
    "version": "20.0.0",
    "path":    "/usr/local/bin/node",
})
```

### Creating a Formatter

```go
// Human-readable formatter
formatter := output.NewFormatter(output.FormatterOptions{
    Format:  output.FormatHuman,
    NoColor: false,
    Verbose: true,
})

// JSON formatter for scripting
jsonFormatter := output.NewFormatter(output.FormatterOptions{
    Format: output.FormatJSON,
})

// Set as global formatter
output.SetGlobalFormatter(formatter)
```

### Progress Indicators

```go
// Create progress indicator
progress := output.NewProgressIndicator(output.ProgressOptions{
    Message:        "Downloading node-v20.0.0.tar.gz",
    ShowPercentage: true,
    ShowBytes:      true,
    ShowSpeed:      true,
})

// Start progress
progress.Start()

// Update progress
for downloaded := int64(0); downloaded < total; downloaded += chunkSize {
    progress.Update(downloaded, total)
}

// Finish successfully
progress.Finish()

// Or mark as failed
// progress.Fail(err)
```

### Table Output

```go
headers := []string{"Tool", "Version", "Status"}
rows := [][]string{
    {"node", "20.0.0", "installed"},
    {"python", "3.11.0", "installed"},
    {"go", "1.21.0", "not installed"},
}

output.Table(headers, rows)
```

### Structured Data

```go
data := map[string]interface{}{
    "tool":         "node",
    "version":      "20.0.0",
    "install_path": "/usr/local/bin/node",
    "installed_at": time.Now(),
}

output.Data(data)
```

## Integration with CLI Commands

### Setting Up Formatter Based on Flags

```go
func setupFormatter(jsonOutput, noColor, quiet, verbose bool) {
    format := output.FormatHuman
    if jsonOutput {
        format = output.FormatJSON
    }

    formatter := output.NewFormatter(output.FormatterOptions{
        Format:  format,
        NoColor: noColor || !output.IsColorSupported(),
        Quiet:   quiet,
        Verbose: verbose,
    })

    output.SetGlobalFormatter(formatter)
}
```

### Using in Commands

```go
func installCommand(cmd *cobra.Command, args []string) error {
    tool := args[0]
    version := args[1]

    output.Info("Installing tool", map[string]interface{}{
        "tool":    tool,
        "version": version,
    })

    // Create progress indicator
    progress := output.NewProgressIndicator(output.ProgressOptions{
        Message:        fmt.Sprintf("Downloading %s@%s", tool, version),
        ShowPercentage: true,
        ShowBytes:      true,
        ShowSpeed:      true,
    })

    progress.Start()

    // Perform installation with progress updates
    err := installTool(tool, version, progress)
    if err != nil {
        progress.Fail(err)
        output.Error("Installation failed", map[string]interface{}{
            "tool":    tool,
            "version": version,
            "error":   err.Error(),
        })
        return err
    }

    progress.Finish()
    output.Success("Installation completed", map[string]interface{}{
        "tool":    tool,
        "version": version,
    })

    return nil
}
```

## Output Formats

### Human Format

Human-readable output with colors and symbols:

```
ℹ Starting installation tool=node version=20.0.0
Downloading node-v20.0.0.tar.gz [████████████████████░░░░░░░░] 75.0% 7.5 MB / 10.0 MB 1.2 MB/s
✓ Installation completed tool=node version=20.0.0
```

### JSON Format

Structured JSON output for scripting:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Starting installation",
  "fields": {
    "tool": "node",
    "version": "20.0.0"
  }
}
{
  "timestamp": "2024-01-15T10:30:15Z",
  "level": "success",
  "message": "Installation completed",
  "fields": {
    "tool": "node",
    "version": "20.0.0"
  }
}
```

## Color Support

The package automatically detects color support:

- Respects `NO_COLOR` environment variable
- Checks `TERM` environment variable
- Detects if stdout is a terminal
- Can be explicitly disabled with `NoColor` option

## Thread Safety

The global formatter is thread-safe and can be safely accessed from multiple goroutines:

```go
// Safe to call from multiple goroutines
go func() {
    output.Info("Background task started")
}()

go func() {
    output.Success("Background task completed")
}()
```

## Testing

The package includes comprehensive unit tests:

```bash
go test ./internal/cli/output/...
```

Run tests with coverage:

```bash
go test -cover ./internal/cli/output/...
```

## Requirements Satisfied

This implementation satisfies the following requirements:

- **Requirement 23.5**: Progress indicators for long-running operations (downloads, installations)
- **Requirement 23.6**: JSON output format (--json flag) and color-coded output

## Design Principles

1. **Separation of Concerns**: Formatters are independent and can be easily extended
2. **Flexibility**: Support for multiple output formats and customization options
3. **User Experience**: Clear, informative output with visual feedback
4. **Automation-Friendly**: JSON output for scripting and CI/CD integration
5. **Performance**: Efficient rendering with minimal overhead
6. **Accessibility**: Respects user preferences (NO_COLOR, quiet mode)
