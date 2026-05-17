// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/spf13/cobra"
)

var (
	// uninstallForce skips confirmation prompt for destructive operation
	uninstallForce bool
)

// init registers the uninstall command to the root command.
func init() {
	// Register command flags
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "skip confirmation prompt")

	// Add command to root - this will be called after rootCmd is initialized
	if rootCmd != nil {
		rootCmd.AddCommand(uninstallCmd)
	}
}

// RegisterUninstallCommand registers the uninstall command to the root command.
// This function should be called after rootCmd is initialized.
func RegisterUninstallCommand() {
	if rootCmd != nil {
		rootCmd.AddCommand(uninstallCmd)
	}
}

// uninstallCmd represents the uninstall command which removes a specific version of a tool.
var uninstallCmd = &cobra.Command{
	Use:     "uninstall <tool> [version]",
	Aliases: []string{"un"},
	Short:   "Uninstall a specific version of a development tool",
	Long: `Uninstall a specific version of a development tool.

The uninstall command removes the specified version of a tool, including:
- The tool installation directory
- Shim scripts
- Database records

This is a destructive operation and requires explicit confirmation unless
the --force flag is used.

Examples:
  # Uninstall Node.js version 20.0.0 (with confirmation)
  unirtm uninstall node 20.0.0

  # Uninstall Python version 3.11.5 without confirmation
  unirtm uninstall python 3.11.5 --force

  # Uninstall with JSON output
  unirtm uninstall go 1.21.0 --json --force`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUninstall,
}

// runUninstall executes the uninstall command.
// It validates input, requires explicit confirmation for destructive operation,
// delegates to the Installation Manager, and cleans up shims and database records.
//
// Validates: Requirements 8.2, 23.2
func runUninstall(cmd *cobra.Command, args []string) error {
	tool := args[0]
	var version string
	if len(args) > 1 {
		version = args[1]
	} else if strings.Contains(tool, "@") {
		parts := strings.SplitN(tool, "@", 2)
		tool = parts[0]
		version = parts[1]
	}

	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Validate input
	if tool == "" {
		formatter.Error("Tool name cannot be empty")
		return fmt.Errorf("tool name is required")
	}

	// Dry-run: show intent without side effects
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would uninstall %s@%s — no changes made", tool, version), map[string]interface{}{
			"tool":    tool,
			"version": version,
			"dry_run": true,
		})
		return nil
	}

	// Initialize dependencies
	ctx := context.Background()

	// Create database connection
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]any{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	// Create repositories
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create installation repository", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	if version == "" {
		// List all installations to find matches for this tool
		installations, err := installRepo.List(ctx)
		if err != nil {
			formatter.Error("Failed to list installations", map[string]any{"error": err.Error()})
			return err
		}
		var matches []*repository.Installation
		for _, inst := range installations {
			if inst.Tool == tool {
				matches = append(matches, inst)
			}
		}
		if len(matches) == 0 {
			formatter.Error(fmt.Sprintf("Tool %s is not installed", tool))
			return fmt.Errorf("tool %s is not installed", tool)
		} else if len(matches) == 1 {
			version = matches[0].Version
		} else {
			var versions []string
			for _, m := range matches {
				versions = append(versions, m.Version)
			}
			formatter.Error(fmt.Sprintf("Multiple versions installed for tool %s: %s. Please specify a version to uninstall.", tool, strings.Join(versions, ", ")))
			return fmt.Errorf("multiple versions installed for %s", tool)
		}
	}

	// Check if the tool is installed
	installation, err := installRepo.FindByToolAndVersion(ctx, tool, version)
	if err != nil {
		formatter.Error(fmt.Sprintf("Tool %s@%s is not installed", tool, version), map[string]any{
			"tool":    tool,
			"version": version,
			"error":   err.Error(),
		})
		return fmt.Errorf("tool %s@%s not found: %w", tool, version, err)
	}

	// Display information about what will be uninstalled
	formatter.Info(fmt.Sprintf("Tool to uninstall: %s@%s", tool, version), map[string]any{
		"tool":         tool,
		"version":      version,
		"install_path": installation.InstallPath,
		"backend":      installation.Backend,
	})

	// Require explicit confirmation unless --force flag is used
	if !uninstallForce && !quiet {
		confirmed, err := promptConfirmation(fmt.Sprintf("Are you sure you want to uninstall %s@%s? This action cannot be undone.", tool, version))
		if err != nil {
			formatter.Error("Failed to read confirmation", map[string]any{
				"error": err.Error(),
			})
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			formatter.Info("Uninstall cancelled by user", map[string]any{
				"tool":    tool,
				"version": version,
			})
			return nil
		}
	}

	// Create backend registry
	backendRegistry := backend.NewRegistry()

	// Create provider registry
	providerRegistry := provider.NewRegistry()

	// Create download manager
	downloadManager := download.NewManager()
	downloadManager.Register("https", download.NewHTTPDownloader())
	downloadManager.Register("http", download.NewHTTPDownloader())

	// Create transaction manager
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create installation manager
	installManager := service.NewInstallationManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
		nil,
	)

	// Perform uninstallation
	formatter.Info(fmt.Sprintf("Uninstalling %s@%s", tool, version), map[string]any{
		"tool":    tool,
		"version": version,
	})

	startTime := time.Now()
	err = installManager.Uninstall(ctx, tool, version)
	duration := time.Since(startTime)

	if err != nil {
		formatter.Error(fmt.Sprintf("Uninstallation failed: %s", err.Error()), map[string]any{
			"tool":     tool,
			"version":  version,
			"duration": duration.String(),
		})
		return fmt.Errorf("uninstall %s@%s: %w", tool, version, err)
	}

	// Display success message
	formatter.Success(fmt.Sprintf("Successfully uninstalled %s@%s", tool, version), map[string]any{
		"tool":     tool,
		"version":  version,
		"duration": duration.String(),
	})

	return nil
}

// promptConfirmation prompts the user for yes/no confirmation.
// Returns true if the user confirms (y/yes), false otherwise.
func promptConfirmation(message string) (bool, error) {
	fmt.Printf("%s [y/N]: ", message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}
