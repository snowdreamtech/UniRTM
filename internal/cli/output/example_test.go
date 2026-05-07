package output_test

import (
	"fmt"
	"time"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
)

// ExampleNewFormatter demonstrates creating different formatter types
func ExampleNewFormatter() {
	// Create a human-readable formatter
	humanFormatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true, // Disable colors for example output
	})

	humanFormatter.Info("Starting installation")
	humanFormatter.Success("Installation completed")

	// Create a JSON formatter
	jsonFormatter := output.NewFormatter(output.FormatterOptions{
		Format: output.FormatJSON,
	})

	jsonFormatter.Info("Starting installation")
	// Output will be JSON formatted
}

// ExampleFormatter_Info demonstrates outputting informational messages
func ExampleFormatter_Info() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	formatter.Info("Starting tool installation")

	// With structured fields
	formatter.Info("Tool installed", map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
	})
}

// ExampleFormatter_Success demonstrates outputting success messages
func ExampleFormatter_Success() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	formatter.Success("Installation completed successfully")

	// With structured fields
	formatter.Success("Tool activated", map[string]interface{}{
		"tool": "node",
		"path": "/usr/local/bin/node",
	})
}

// ExampleFormatter_Warning demonstrates outputting warning messages
func ExampleFormatter_Warning() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	formatter.Warning("Configuration file not found, using defaults")

	// With structured fields
	formatter.Warning("Tool version mismatch", map[string]interface{}{
		"expected": "20.0.0",
		"found":    "18.0.0",
	})
}

// ExampleFormatter_Error demonstrates outputting error messages
func ExampleFormatter_Error() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	formatter.Error("Failed to download tool")

	// With structured fields
	formatter.Error("Installation failed", map[string]interface{}{
		"tool":  "node",
		"error": "checksum mismatch",
	})
}

// ExampleFormatter_Table demonstrates outputting tabular data
func ExampleFormatter_Table() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	headers := []string{"Tool", "Version", "Status"}
	rows := [][]string{
		{"node", "20.0.0", "installed"},
		{"python", "3.11.0", "installed"},
		{"go", "1.21.0", "not installed"},
	}

	formatter.Table(headers, rows)
}

// ExampleNewProgressIndicator demonstrates creating and using a progress indicator
func ExampleNewProgressIndicator() {
	progress := output.NewProgressIndicator(output.ProgressOptions{
		Message:        "Downloading node-v20.0.0.tar.gz",
		ShowPercentage: true,
		ShowBytes:      true,
		ShowSpeed:      true,
		NoColor:        true,
	})

	progress.Start()

	// Simulate download progress
	total := int64(10 * 1024 * 1024) // 10 MB
	for downloaded := int64(0); downloaded < total; downloaded += 1024 * 1024 {
		progress.Update(downloaded, total)
		time.Sleep(100 * time.Millisecond)
	}

	progress.Finish()
}

// ExampleProgressIndicator_Fail demonstrates handling failed operations
func ExampleProgressIndicator_Fail() {
	progress := output.NewProgressIndicator(output.ProgressOptions{
		Message: "Installing tool",
		NoColor: true,
	})

	progress.Start()

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Operation failed
	progress.Fail(fmt.Errorf("checksum verification failed"))
}

// ExampleSetGlobalFormatter demonstrates using the global formatter
func ExampleSetGlobalFormatter() {
	// Set up global formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})
	output.SetGlobalFormatter(formatter)

	// Use global formatter functions
	output.Info("Starting operation")
	output.Success("Operation completed")
	output.Warning("Potential issue detected")
	output.Error("Operation failed")
}

// ExampleFormatter_Data demonstrates outputting structured data
func ExampleFormatter_Data() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  output.FormatHuman,
		NoColor: true,
	})

	// Output simple data
	formatter.Data("node@20.0.0")

	// Output structured data
	data := map[string]interface{}{
		"tool":         "node",
		"version":      "20.0.0",
		"install_path": "/usr/local/bin/node",
		"installed_at": time.Now().Format(time.RFC3339),
	}
	formatter.Data(data)
}

// ExampleJSONFormatter demonstrates JSON output format
func ExampleJSONFormatter() {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format: output.FormatJSON,
	})

	formatter.Info("Starting installation", map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
	})

	formatter.Success("Installation completed", map[string]interface{}{
		"tool":         "node",
		"version":      "20.0.0",
		"install_path": "/usr/local/bin/node",
	})

	// Output will be JSON formatted with timestamps
}

// ExampleIsColorSupported demonstrates checking color support
func ExampleIsColorSupported() {
	if output.IsColorSupported() {
		fmt.Println("Terminal supports colors")
	} else {
		fmt.Println("Terminal does not support colors")
	}
}

// ExampleNoOpProgressIndicator demonstrates a no-op progress indicator
func ExampleNoOpProgressIndicator() {
	// Useful for testing or when progress output is not desired
	progress := output.NewNoOpProgressIndicator()

	progress.Start()
	progress.Update(50, 100)
	progress.Finish()

	// No output is produced
}
