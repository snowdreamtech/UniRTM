# Integration Guide

This guide shows how to integrate the output package with Cobra CLI commands.

## Setting Up the Global Formatter

In your root command's `PersistentPreRun` hook, set up the global formatter based on CLI flags:

```go
// cmd/1.main.go
func setupLogging(cmd *cobra.Command, args []string) {
 // Determine output format
 format := output.FormatHuman
 if jsonOutput {
  format = output.FormatJSON
 }

 // Create formatter with options from flags
 formatter := output.NewFormatter(output.FormatterOptions{
  Format:  format,
  NoColor: !output.IsColorSupported() || zerolog.GlobalLevel() == zerolog.Disabled,
  Quiet:   quiet,
  Verbose: verbose,
 })

 // Set as global formatter
 output.SetGlobalFormatter(formatter)

 // Set log level from CLI flags
 if quiet {
  zerolog.SetGlobalLevel(zerolog.Disabled)
 } else if verbose {
  zerolog.SetGlobalLevel(zerolog.DebugLevel)
 } else {
  zerolog.SetGlobalLevel(zerolog.InfoLevel)
 }
}
```

## Using in Commands

### Simple Command with Output

```go
// cmd/list.go
package cmd

import (
 "github.com/snowdreamtech/unirtm/internal/cli/output"
 "github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
 Use:   "list",
 Short: "List installed tools",
 RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
 output.Info("Fetching installed tools")

 // Get installed tools from service
 tools, err := getInstalledTools()
 if err != nil {
  output.Error("Failed to fetch tools", map[string]interface{}{
   "error": err.Error(),
  })
  return err
 }

 // Display as table
 headers := []string{"Tool", "Version", "Install Path"}
 rows := make([][]string, len(tools))
 for i, tool := range tools {
  rows[i] = []string{tool.Name, tool.Version, tool.InstallPath}
 }

 output.Table(headers, rows)
 output.Success("Listed %d installed tools", len(tools))

 return nil
}
```

### Command with Progress Indicator

```go
// cmd/install.go
package cmd

import (
 "fmt"

 "github.com/snowdreamtech/unirtm/internal/cli/output"
 "github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
 Use:   "install <tool> <version>",
 Short: "Install a tool",
 Args:  cobra.ExactArgs(2),
 RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
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

 // Perform installation with progress callback
 err := installService.Install(tool, version, func(current, total int64) {
  progress.Update(current, total)
 })

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

### Command with Structured Data Output

```go
// cmd/info.go
package cmd

import (
 "github.com/snowdreamtech/unirtm/internal/cli/output"
 "github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
 Use:   "info <tool>",
 Short: "Show tool information",
 Args:  cobra.ExactArgs(1),
 RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
 tool := args[0]

 output.Info("Fetching tool information", map[string]interface{}{
  "tool": tool,
 })

 // Get tool info from service
 info, err := getToolInfo(tool)
 if err != nil {
  output.Error("Failed to fetch tool information", map[string]interface{}{
   "tool":  tool,
   "error": err.Error(),
  })
  return err
 }

 // Output structured data
 // In human format: pretty-printed
 // In JSON format: structured JSON
 output.Data(map[string]interface{}{
  "name":         info.Name,
  "version":      info.Version,
  "description":  info.Description,
  "homepage":     info.Homepage,
  "license":      info.License,
  "install_path": info.InstallPath,
  "installed_at": info.InstalledAt,
 })

 return nil
}
```

## Service Layer Integration

### Download Service with Progress

```go
// internal/service/download.go
package service

import (
 "context"
 "io"
 "net/http"
 "os"

 "github.com/snowdreamtech/unirtm/internal/cli/output"
)

type DownloadService struct {
 client *http.Client
}

// DownloadWithProgress downloads a file with progress reporting
func (s *DownloadService) DownloadWithProgress(
 ctx context.Context,
 url string,
 destination string,
 progressCallback func(current, total int64),
) error {
 // Create HTTP request
 req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
 if err != nil {
  return fmt.Errorf("create request: %w", err)
 }

 // Execute request
 resp, err := s.client.Do(req)
 if err != nil {
  return fmt.Errorf("execute request: %w", err)
 }
 defer resp.Body.Close()

 if resp.StatusCode != http.StatusOK {
  return fmt.Errorf("unexpected status: %s", resp.Status)
 }

 // Create destination file
 out, err := os.Create(destination)
 if err != nil {
  return fmt.Errorf("create file: %w", err)
 }
 defer out.Close()

 // Copy with progress
 total := resp.ContentLength
 var current int64

 buf := make([]byte, 32*1024) // 32KB buffer
 for {
  n, err := resp.Body.Read(buf)
  if n > 0 {
   _, writeErr := out.Write(buf[:n])
   if writeErr != nil {
    return fmt.Errorf("write file: %w", writeErr)
   }

   current += int64(n)
   if progressCallback != nil {
    progressCallback(current, total)
   }
  }

  if err == io.EOF {
   break
  }
  if err != nil {
   return fmt.Errorf("read response: %w", err)
  }
 }

 return nil
}
```

## Testing with Output Package

### Testing Commands

```go
// cmd/install_test.go
package cmd

import (
 "bytes"
 "testing"

 "github.com/snowdreamtech/unirtm/internal/cli/output"
 "github.com/stretchr/testify/assert"
)

func TestInstallCommand(t *testing.T) {
 // Capture output
 buf := &bytes.Buffer{}
 formatter := output.NewFormatter(output.FormatterOptions{
  Format:  output.FormatHuman,
  Writer:  buf,
  NoColor: true,
 })
 output.SetGlobalFormatter(formatter)

 // Run command
 err := runInstall(nil, []string{"node", "20.0.0"})
 assert.NoError(t, err)

 // Verify output
 outputStr := buf.String()
 assert.Contains(t, outputStr, "Installing tool")
 assert.Contains(t, outputStr, "Installation completed")
}

func TestInstallCommand_JSON(t *testing.T) {
 // Capture JSON output
 buf := &bytes.Buffer{}
 formatter := output.NewFormatter(output.FormatterOptions{
  Format: output.FormatJSON,
  Writer: buf,
 })
 output.SetGlobalFormatter(formatter)

 // Run command
 err := runInstall(nil, []string{"node", "20.0.0"})
 assert.NoError(t, err)

 // Verify JSON output
 outputStr := buf.String()
 assert.Contains(t, outputStr, `"level"`)
 assert.Contains(t, outputStr, `"message"`)
}
```

## Best Practices

### 1. Use Global Formatter for Consistency

```go
// Good: Use global formatter
output.Info("Starting operation")
output.Success("Operation completed")

// Avoid: Creating formatters in commands
formatter := output.NewFormatter(...)
formatter.Info("Starting operation")
```

### 2. Provide Structured Fields for Context

```go
// Good: Include relevant context
output.Error("Installation failed", map[string]interface{}{
 "tool":    tool,
 "version": version,
 "error":   err.Error(),
})

// Avoid: Plain messages without context
output.Error("Installation failed")
```

### 3. Use Progress Indicators for Long Operations

```go
// Good: Show progress for downloads
progress := output.NewProgressIndicator(...)
progress.Start()
// ... perform operation with updates
progress.Finish()

// Avoid: Silent long-running operations
downloadFile(url, dest) // No feedback
```

### 4. Handle Both Human and JSON Output

```go
// Good: Use output.Data for structured output
output.Data(map[string]interface{}{
 "tool":    tool,
 "version": version,
 "status":  "installed",
})

// This works in both human and JSON formats
```

### 5. Clean Up Progress Indicators

```go
// Good: Always finish or fail
progress := output.NewProgressIndicator(...)
progress.Start()
defer func() {
 if err != nil {
  progress.Fail(err)
 } else {
  progress.Finish()
 }
}()
```

## Environment Variables

The output package respects standard environment variables:

- `NO_COLOR`: Disables color output when set
- `TERM`: Checks terminal capabilities (e.g., `dumb` disables colors)

## CLI Flags

Standard flags that should be supported:

- `--json, -j`: Enable JSON output format
- `--quiet, -q`: Suppress non-essential output
- `--verbose, -v`: Enable verbose output with detailed fields
- `--no-color`: Explicitly disable color output (if needed)
