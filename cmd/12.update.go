// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
 
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
	// updateAll updates all installed tools
	updateAll bool
	// updatePreview shows a preview of what will be updated without applying
	updatePreview bool
	// updateForce skips confirmation prompt
	updateForce bool
)

// init registers the update command to the root command.
func init() {
	updateCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "update all installed tools")
	updateCmd.Flags().BoolVarP(&updatePreview, "preview", "p", false, "show what would be updated without applying changes")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "skip confirmation prompt")

	if rootCmd != nil {
		rootCmd.AddCommand(updateCmd)
	}
}

// updateCmd represents the update command which updates installed tools.
var updateCmd = &cobra.Command{
	Use:   "update [tool] [version]",
	Short: "Update installed development tools",
	Long: `Update installed development tools to newer versions.

The update command checks for available updates and applies them.
It respects version constraints defined in configuration files.

Examples:
  # Update a specific tool to its latest version
  unirtm update node

  # Update a specific tool to a specific version
  unirtm update node 22.0.0

  # Update all installed tools
  unirtm update --all

  # Preview what would be updated without applying
  unirtm update --all --preview

  # Update with JSON output
  unirtm update node --json`,
	Aliases: []string{"up", "upgrade"},
	Args:    cobra.MaximumNArgs(2),
	RunE:    runUpdate,
}

// runUpdate executes the update command.
//
// Validates: Requirements 25.1, 25.2, 25.3, 25.5, 23.2
func runUpdate(cmd *cobra.Command, args []string) error {
	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Validate arguments
	if len(args) == 0 && !updateAll {
		return fmt.Errorf("specify a tool name or use --all to update all tools")
	}

	// Initialize dependencies
	ctx := context.Background()

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

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return fmt.Errorf("create installation repository: %w", err)
	}

	auditRepo, err := sqlite.NewAuditRepository(db.Conn())
	if err != nil {
		return fmt.Errorf("create audit repository: %w", err)
	}

	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()
	downloadManager.Register("https", download.NewHTTPDownloader())
	downloadManager.Register("http", download.NewHTTPDownloader())
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	updateManager := service.NewUpdateManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		auditRepo,
		txManager,
		nil, // no config manager for now
	)

	// Preview mode — show what would be updated
	if updatePreview {
		formatter.Info("Checking for available updates...", nil)

		preview, err := updateManager.PreviewUpdates(ctx)
		if err != nil {
			formatter.Error("Failed to check for updates", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("check for updates: %w", err)
		}

		if preview.TotalUpdates == 0 {
			formatter.Info("All tools are up to date", nil)
			return nil
		}

		if jsonOutput {
			formatter.Success("Available updates", map[string]interface{}{
				"count":          preview.TotalUpdates,
				"estimated_time": preview.EstimatedTime.String(),
				"updates":        preview.Updates,
			})
			return nil
		}

		tableData := pterm.TableData{
			{"TOOL", "CURRENT", "LATEST", "BACKEND"},
		}
		for _, u := range preview.Updates {
			tableData = append(tableData, []string{
				u.Tool,
				pterm.Gray(u.CurrentVersion),
				pterm.Green(u.LatestVersion),
				u.Backend,
			})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		fmt.Printf("\nEstimated time: %s\n", pterm.Cyan(preview.EstimatedTime))
		fmt.Println("\nRun without " + pterm.LightYellow("--preview") + " to apply updates.")
		return nil
	}

	// Update a specific tool
	if len(args) >= 1 {
		tool := args[0]
		targetVersion := "latest"
		if len(args) == 2 {
			targetVersion = args[1]
		}

		formatter.Info(fmt.Sprintf("Updating %s to %s...", tool, targetVersion), map[string]interface{}{
			"tool":    tool,
			"version": targetVersion,
		})

		if !updateForce && !quiet {
			confirmed, err := promptConfirmation(fmt.Sprintf("Update %s to %s?", tool, targetVersion))
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				formatter.Info("Update cancelled by user", nil)
				return nil
			}
		}

		result, err := updateManager.UpdateTool(ctx, tool, targetVersion)
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to update %s", tool), map[string]interface{}{
				"tool":  tool,
				"error": err.Error(),
			})
			return fmt.Errorf("update %s: %w", tool, err)
		}

		if jsonOutput {
			formatter.Success("Update complete", map[string]interface{}{
				"tool":        result.Tool,
				"old_version": result.OldVersion,
				"new_version": result.NewVersion,
				"duration":    result.Duration.String(),
			})
		} else {
			fmt.Printf("✓ Updated %s: %s → %s (%s)\n",
				result.Tool, result.OldVersion, result.NewVersion, result.Duration)
		}
		return nil
	}

	// Update all tools
	formatter.Info("Checking for available updates...", nil)

	preview, err := updateManager.PreviewUpdates(ctx)
	if err != nil {
		formatter.Error("Failed to check for updates", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("check for updates: %w", err)
	}

	if preview.TotalUpdates == 0 {
		formatter.Info("All tools are up to date", nil)
		return nil
	}

	// Show preview and ask for confirmation
	if !updateForce && !quiet {
		fmt.Printf("\n%d update(s) available:\n", preview.TotalUpdates)
		for _, u := range preview.Updates {
			fmt.Printf("  • %s: %s → %s\n", u.Tool, u.CurrentVersion, u.LatestVersion)
		}
		fmt.Println()

		confirmed, err := promptConfirmation("Apply all updates?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			formatter.Info("Update cancelled by user", nil)
			return nil
		}
	}

	results, err := updateManager.UpdateAll(ctx)
	if err != nil {
		formatter.Error("Update failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("update all: %w", err)
	}

	// Display results
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			if !quiet {
				fmt.Printf("✓ Updated %s: %s → %s (%s)\n", r.Tool, r.OldVersion, r.NewVersion, r.Duration)
			}
		} else {
			failCount++
			fmt.Fprintf(os.Stderr, "✗ Failed to update %s: %s\n", r.Tool, r.Error)
		}
	}

	if jsonOutput {
		formatter.Success("Update complete", map[string]interface{}{
			"success": successCount,
			"failed":  failCount,
			"results": results,
		})
	} else {
		fmt.Printf("\nDone: %d updated, %d failed\n", successCount, failCount)
	}

	if failCount > 0 {
		return fmt.Errorf("%d tool(s) failed to update", failCount)
	}
	return nil
}
