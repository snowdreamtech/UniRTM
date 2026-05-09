// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/spf13/cobra"
)

var (
	// installBackend specifies the backend to use for installation
	installBackend string
)

// init registers the install command to the root command.
func init() {
	// Register command flags
	installCmd.Flags().StringVarP(&installBackend, "backend", "b", "", "backend to use for installation (default: auto-detect)")

	// Add command to root - this will be called after rootCmd is initialized
	if rootCmd != nil {
		rootCmd.AddCommand(installCmd)
	}
}

// RegisterInstallCommand registers the install command to the root command.
// This function should be called after rootCmd is initialized.
func RegisterInstallCommand() {
	if rootCmd != nil {
		rootCmd.AddCommand(installCmd)
	}
}

// installCmd represents the install command which installs a specific version of a tool.
var installCmd = &cobra.Command{
	Use:   "install <tool> <version>",
	Short: "Install a specific version of a development tool",
	Long: `Install a specific version of a development tool.

The install command downloads and installs the specified version of a tool
using the appropriate backend. It validates the installation, records it in
the database, and generates shim scripts.

Examples:
  # Install Node.js version 20.0.0
  unirtm install node 20.0.0

  # Install Python version 3.11.5 using a specific backend
  unirtm install python 3.11.5 --backend github

  # Install with JSON output
  unirtm install go 1.21.0 --json`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runInstall,
}

// runInstall executes the install command.
// It validates input, delegates to the Installation Manager, and displays progress and results.
//
// Validates: Requirements 9.1, 9.2, 9.3, 23.2
func runInstall(cmd *cobra.Command, args []string) error {
	tool := args[0]
	version := "latest"

	// Parse "tool@version" syntax (like mise)
	if strings.Contains(tool, "@") {
		parts := strings.SplitN(tool, "@", 2)
		tool = parts[0]
		if parts[1] != "" {
			version = parts[1]
		}
	} else if len(args) == 2 {
		version = args[1]
	} else if len(args) == 1 {
		// Just 'unirtm install tool' means latest
		version = "latest"
	}

	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Parse backend prefix (e.g., "npm:typescript" -> backend: "npm", tool: "typescript")
	backendName := getBackendName()
	if strings.Contains(tool, ":") {
		parts := strings.SplitN(tool, ":", 2)
		backendName = parts[0]
		tool = parts[1]
	}

	// Validate input
	if tool == "" {
		formatter.Error("Tool name cannot be empty")
		return fmt.Errorf("tool name is required")
	}

	if version == "" {
		formatter.Error("Version cannot be empty")
		return fmt.Errorf("version is required")
	}

	// Display start message
	formatter.Info(fmt.Sprintf("Installing %s@%s", tool, version), map[string]interface{}{
		"tool":    tool,
		"version": version,
		"backend": backendName,
	})

	// Dry-run: show intent without side effects
	if dryRun {
		formatter.Info("[dry-run] Would install "+tool+"@"+version+" — no changes made", map[string]interface{}{
			"tool":    tool,
			"version": version,
			"backend": backendName,
			"dry_run": true,
		})
		return nil
	}

	// Initialize dependencies
	ctx := context.Background()

	// Create backend registry
	backendRegistry := backend.NewRegistry()

	// Create provider registry
	providerRegistry := provider.NewRegistry()

	// Create download manager
	downloadManager := download.NewManager()
	downloadManager.Register("https", download.NewHTTPDownloader())
	downloadManager.Register("http", download.NewHTTPDownloader())

	// Create database connection
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	// Create repositories
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create installation repository", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	// Create transaction manager
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create installation manager
	installManager := service.NewInstallationManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
	)

	// Display start message
	formatter.Info(fmt.Sprintf("Initializing installation for %s@%s", tool, version), map[string]interface{}{
		"tool":    tool,
		"version": version,
		"backend": backendName,
	})

	// Start gorgeous loading animation for initialization
	spinner, _ := pterm.DefaultSpinner.Start("Resolving backend and initializing...")

	// Since native tools (npm, cargo, asdf) have their own progress bars that require a raw TTY,
	// we stop our spinner right before the heavy lifting to avoid clashing with their native progress output.
	spinner.Success("Initialization complete. Handing over to native provider...")

	// Perform installation
	startTime := time.Now()
	err = installManager.Install(ctx, tool, version, backendName)
	duration := time.Since(startTime)

	if err != nil {
		pterm.Error.Printf("Installation failed: %v\n", err)
		formatter.Error(fmt.Sprintf("Installation failed: %s", err.Error()), map[string]interface{}{
			"tool":     tool,
			"version":  version,
			"duration": duration.String(),
		})
		return fmt.Errorf("install %s@%s: %w", tool, version, err)
	}

	// Display success message
	pterm.Success.Printf("Successfully installed %s@%s (took %s)\n", tool, version, duration.Round(time.Millisecond).String())

	return nil
}

// getOutputFormat returns the output format based on the jsonOutput flag.
func getOutputFormat() output.OutputFormat {
	if jsonOutput {
		return output.FormatJSON
	}
	return output.FormatHuman
}

// getBackendName returns the backend name to use for installation.
// If installBackend flag is set, it returns that value.
// Otherwise, it returns "github" as the default backend.
func getBackendName() string {
	if installBackend != "" {
		return installBackend
	}
	return "github" // Default backend
}

// getDefaultDatabasePath returns the default path for the SQLite database.
// TODO: This should be configurable via configuration file or environment variable.
func getDefaultDatabasePath() string {
	// Use XDG_DATA_HOME if set, otherwise use ~/.local/share
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory
			return "./unirtm.db"
		}
		dataHome = homeDir + "/.local/share"
	}

	// Create unirtm data directory
	unirtmDataDir := dataHome + "/unirtm"
	if err := os.MkdirAll(unirtmDataDir, 0755); err != nil {
		// Fallback to current directory
		return "./unirtm.db"
	}

	return unirtmDataDir + "/unirtm.db"
}
