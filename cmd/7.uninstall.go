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

	"github.com/pterm/pterm"
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

// uninstallCmd represents the uninstall command which removes specified versions of tools.
var uninstallCmd = &cobra.Command{
	Use:     "uninstall [tool[@version]...]",
	Aliases: []string{"un"},
	Short:   "Uninstall development tools and package specifications",
	Long: `Uninstall development tools and package specifications.

The uninstall command removes specified versions of tools, including:
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

  # Uninstall multiple tools concurrently
  unirtm uninstall node@20.11.0 go@1.22.1 python@3.12.0 --force

  # Uninstall with JSON output
  unirtm uninstall go 1.21.0 --json --force`,
	Args: cobra.ArbitraryArgs,
	RunE: runUninstall,
}

// runUninstall executes the uninstall command.
// It validates input, requires explicit confirmation for destructive operation,
// delegates to the Installation Manager, and cleans up shims and database records.
//
// Validates: Requirements 8.2, 23.2
func runUninstall(cmd *cobra.Command, args []string) error {
	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	if len(args) == 0 {
		return fmt.Errorf("tool specification is required")
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
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create repositories
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return fmt.Errorf("failed to create installation repository: %w", err)
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

	type uninstallTarget struct {
		toolName    string
		version     string
		backendName string
		original    string
	}
	var targets []uninstallTarget

	// Determine if we should treat it as a single tool + version (backward compatibility)
	isLegacySingleTool := false
	if len(args) == 2 {
		firstArg := args[0]
		secondArg := args[1]
		if !strings.Contains(firstArg, "@") && !strings.Contains(firstArg, ":") &&
			!strings.Contains(secondArg, "@") {
			isLegacySingleTool = true
		}
	}

	if isLegacySingleTool {
		_, tool, _, _ := installManager.ParseToolSpec(args[0])
		version := args[1]

		if tool == "" {
			return fmt.Errorf("tool name is required")
		}
		if version == "" {
			return fmt.Errorf("version is required")
		}

		// Verify the exact tool/version is installed
		installation, err := installRepo.FindByToolAndVersion(ctx, tool, version)
		if err != nil {
			return fmt.Errorf("Tool %s@%s is not installed", tool, version)
		}

		targets = append(targets, uninstallTarget{
			toolName:    tool,
			version:     version,
			backendName: installation.Backend,
			original:    args[0] + "@" + version,
		})
	} else {
		for _, arg := range args {
			backendName, tool, version, explicitVersion := installManager.ParseToolSpec(arg)

			if tool == "" {
				return fmt.Errorf("tool name is required")
			}

			// If version is not explicit, resolve it
			if !explicitVersion || version == "latest" {
				installations, err := installRepo.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list installations: %w", err)
				}
				var matches []*repository.Installation
				for _, inst := range installations {
					if inst.Tool == tool {
						matches = append(matches, inst)
					}
				}
				if len(matches) == 0 {
					return fmt.Errorf("Tool %s is not installed", tool)
				} else if len(matches) == 1 {
					version = matches[0].Version
					backendName = matches[0].Backend
				} else {
					var versions []string
					for _, m := range matches {
						versions = append(versions, m.Version)
					}
					return fmt.Errorf("Multiple versions installed for tool %s: %s. Please specify a version to uninstall.", tool, strings.Join(versions, ", "))
				}
			} else {
				// Verify exact tool/version is installed
				installation, err := installRepo.FindByToolAndVersion(ctx, tool, version)
				if err != nil {
					return fmt.Errorf("Tool %s@%s is not installed", tool, version)
				}
				backendName = installation.Backend
			}

			targets = append(targets, uninstallTarget{
				toolName:    tool,
				version:     version,
				backendName: backendName,
				original:    arg,
			})
		}
	}

	// Dry-run: show intent without side effects
	if dryRun {
		for _, t := range targets {
			formatter.Info(fmt.Sprintf("[dry-run] Would uninstall %s@%s — no changes made", t.toolName, t.version), map[string]interface{}{
				"tool":    t.toolName,
				"version": t.version,
				"dry_run": true,
			})
		}
		return nil
	}

	// Require explicit confirmation unless --force flag is used
	if !uninstallForce && !quiet {
		var confirmMsg string
		if len(targets) == 1 {
			confirmMsg = fmt.Sprintf("Are you sure you want to uninstall %s@%s? This action cannot be undone.", targets[0].toolName, targets[0].version)
		} else {
			confirmMsg = "The following tool(s) will be uninstalled:\n"
			for _, t := range targets {
				confirmMsg += fmt.Sprintf("  • %s@%s\n", t.toolName, t.version)
			}
			confirmMsg += "\nAre you sure you want to uninstall these tools? This action cannot be undone."
		}

		confirmed, err := promptConfirmation(confirmMsg)
		if err != nil {
			formatter.Error("Failed to read confirmation", map[string]any{
				"error": err.Error(),
			})
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			formatter.Info("Uninstall cancelled by user")
			return nil
		}
	}

	// Perform uninstallations
	if !quiet {
		formatter.Info(fmt.Sprintf("Uninstalling %d tool(s)...", len(targets)))
	}

	startTime := time.Now()
	for _, t := range targets {
		if !quiet {
			formatter.Info(fmt.Sprintf("Uninstalling %s@%s...", t.toolName, t.version))
		}
		err = installManager.Uninstall(ctx, t.toolName, t.version)
		if err != nil {
			formatter.Error(fmt.Sprintf("Uninstallation failed for %s@%s: %s", t.toolName, t.version, err.Error()))
			return fmt.Errorf("uninstall %s@%s: %w", t.toolName, t.version, err)
		}
		pterm.FgGreen.Printf("✓ Successfully uninstalled %s@%s\n", t.toolName, t.version)
	}
	duration := time.Since(startTime)

	if len(targets) > 1 && !quiet {
		pterm.FgGreen.Printf("✓ All tools uninstalled (took %s)\n", duration.Round(time.Millisecond).String())
	}

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
