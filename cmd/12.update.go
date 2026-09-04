// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
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

func init() {
	updateCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "update all installed tools")
	updateCmd.Flags().BoolVarP(&updatePreview, "preview", "p", false, "show what would be updated without applying changes")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "skip confirmation prompt")

	upgradeCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "update all installed tools")
	upgradeCmd.Flags().BoolVarP(&updatePreview, "preview", "p", false, "show what would be updated without applying changes")
	upgradeCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "skip confirmation prompt")

	if rootCmd != nil {
		rootCmd.AddCommand(updateCmd)
		rootCmd.AddCommand(upgradeCmd)
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
  # Preview all available updates (no changes applied)
  unirtm update --preview

  # Update a specific tool to its latest version
  unirtm update node

  # Update a specific tool to a specific version
  unirtm update node 22.0.0

  # Update all installed tools
  unirtm update --all

  # Preview what would be updated for all tools (explicit)
  unirtm update --all --preview

  # Update with JSON output
  unirtm update node --json`,
	Aliases: []string{"up"},
	Args:    cobra.MaximumNArgs(2),
	RunE:    runUpdate,
}

// upgradeCmd represents the upgrade command, which is an exact alias for update.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade [tool] [version]",
	Short: "Upgrade installed development tools (alias for update)",
	Long: `Upgrade installed development tools to newer versions.

This command is exactly the same as 'unirtm update' and is provided for compatibility.`,
	Aliases: []string{"upg"},
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

	// Validate arguments: require a tool name OR --all OR --preview (preview implies all tools)
	if len(args) == 0 && !updateAll && !updatePreview {
		return fmt.Errorf("specify a tool name or use --all to update all tools")
	}

	// Initialize dependencies
	ctx := context.Background()

	// Load config and apply environment variables
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		cfg.ApplyEnvironment()
	}

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
		cfg, // Pass config manager for GPG keys and settings
	)

	// Preview mode — show what would be updated
	if updatePreview {
		var spinner *pterm.SpinnerPrinter
		if !quiet {
			spinner, _ = output.StartSpinner("Checking for available updates...")
		}

		preview, err := updateManager.PreviewUpdates(ctx)
		if err != nil {
			if spinner != nil {
				spinner.Fail("Failed to check for updates: " + err.Error())
			}
			formatter.Error("Failed to check for updates", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("check for updates: %w", err)
		}

		if spinner != nil {
			spinner.Success("Update check complete!")
		}

		if preview.TotalUpdates == 0 {
			if preview.TotalInstalled == 0 {
				formatter.Warning("No tools are currently installed. Please run 'unirtm install' first.", nil)
			} else {
				formatter.Info("All tools are up to date", nil)
			}
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

		pterm.DefaultSection.Println("Available Updates")
		tableData := pterm.TableData{
			{"TOOL", "CURRENT", "LATEST", "BACKEND"},
		}
		for _, u := range preview.Updates {
			tableData = append(tableData, []string{
				pterm.FgCyan.Sprint(u.Tool),
				pterm.FgGray.Sprint(u.CurrentVersion),
				pterm.FgGreen.Sprint(u.LatestVersion),
				pterm.FgMagenta.Sprint(u.Backend),
			})
		}

		_ = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		output.Infof("Estimated time: %s", pterm.LightCyan(preview.EstimatedTime))
		pterm.Printf("\nRun without %s to apply updates.\n", pterm.LightYellow("--preview"))
		return nil
	}

	// Update a specific tool
	if len(args) >= 1 {
		tool := args[0]
		targetVersion := "latest"
		if len(args) == 2 {
			targetVersion = args[1]
		}

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

		var spinner *pterm.SpinnerPrinter
		if !quiet {
			spinner, _ = output.StartSpinner(fmt.Sprintf("Updating %s to %s...", tool, targetVersion))
		}

		result, err := updateManager.UpdateTool(ctx, tool, "", targetVersion)
		if err != nil {
			if spinner != nil {
				spinner.Fail(fmt.Sprintf("Failed to update %s: %s", tool, err.Error()))
			}
			formatter.Error(fmt.Sprintf("Failed to update %s", tool), map[string]interface{}{
				"tool":  tool,
				"error": err.Error(),
			})
			return fmt.Errorf("update %s: %w", tool, err)
		}

		if spinner != nil {
			spinner.Success(fmt.Sprintf("Successfully updated %s", tool))
		}

		updateConfigAndLockfile(ctx, cfg, backendRegistry, []service.UpdateResult{*result})

		if jsonOutput {
			formatter.Success("Update complete", map[string]interface{}{
				"tool":        result.Tool,
				"old_version": result.OldVersion,
				"new_version": result.NewVersion,
				"duration":    result.Duration.String(),
			})
		} else {
			output.Successf("Updated %s: %s → %s (%s)",
				pterm.LightGreen(result.Tool),
				pterm.FgGray.Sprint(result.OldVersion),
				pterm.LightGreen(result.NewVersion),
				pterm.FgCyan.Sprint(result.Duration))
		}
		return nil
	}

	// Update all tools
	var spinner *pterm.SpinnerPrinter
	if !quiet {
		spinner, _ = output.StartSpinner("Checking for available updates...")
	}

	preview, err := updateManager.PreviewUpdates(ctx)
	if err != nil {
		if spinner != nil {
			spinner.Fail("Failed to check for updates: " + err.Error())
		}
		formatter.Error("Failed to check for updates", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("check for updates: %w", err)
	}

	if spinner != nil {
		spinner.Success("Update check complete!")
	}

	if preview.TotalUpdates == 0 {
		if preview.TotalInstalled == 0 {
			formatter.Warning("No tools are currently installed. Please run 'unirtm install' first.", nil)
		} else {
			formatter.Info("All tools are up to date", nil)
		}
		return nil
	}

	// Show preview and ask for confirmation
	if !updateForce && !quiet {
		pterm.DefaultSection.Printfln("%d update(s) available:", preview.TotalUpdates)
		for _, u := range preview.Updates {
			pterm.Println(fmt.Sprintf("  • %s: %s → %s", pterm.FgCyan.Sprint(u.Tool), pterm.FgGray.Sprint(u.CurrentVersion), pterm.FgGreen.Sprint(u.LatestVersion)))
		}
		pterm.Println()

		confirmed, err := promptConfirmation("Apply all updates?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			formatter.Info("Update cancelled by user", nil)
			return nil
		}
	}

	var updateSpinner *pterm.SpinnerPrinter
	if !quiet {
		updateSpinner, _ = output.StartSpinner("Applying all updates...")
	}

	results, err := updateManager.UpdateAll(ctx)
	if err != nil {
		if updateSpinner != nil {
			updateSpinner.Fail("Update failed: " + err.Error())
		}
		formatter.Error("Update failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("update all: %w", err)
	}

	if updateSpinner != nil {
		updateSpinner.Success("All updates processed!")
	}

	// Display results
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			if !quiet {
				output.Successf("Updated %s: %s → %s (%s)",
					pterm.LightGreen(r.Tool),
					pterm.FgGray.Sprint(r.OldVersion),
					pterm.LightGreen(r.NewVersion),
					pterm.FgCyan.Sprint(r.Duration))
			}
		} else {
			failCount++
			output.Errorf("Failed to update %s: %s", pterm.LightRed(r.Tool), pterm.LightRed(r.Error))
		}
	}

	updateConfigAndLockfile(ctx, cfg, backendRegistry, results)

	if jsonOutput {
		formatter.Success("Update complete", map[string]interface{}{
			"success": successCount,
			"failed":  failCount,
			"results": results,
		})
	} else {
		pterm.Println()
		if failCount == 0 {
			output.Successf("All updates complete: %d tools updated successfully.", successCount)
		} else {
			output.Warningf("Update finished: %d succeeded, %d failed.", successCount, failCount)
		}
	}

	if failCount > 0 {
		return fmt.Errorf("%d tool(s) failed to update", failCount)
	}
	return nil
}

func updateConfigAndLockfile(ctx context.Context, cfg *config.Config, backendRegistry *backend.Registry, results []service.UpdateResult) {
	if len(results) == 0 {
		return
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	if successCount == 0 {
		return
	}

	// 1. Update .unirtm.toml
	if cfg != nil && cfg.ToolsRaw != nil {
		cwd, err := os.Getwd()
		if err == nil {
			configFile := findOrCreateConfigFile(cwd, "")
			content, readErr := config.ReadFileOrEmpty(configFile)
			if readErr == nil {
				modified := false
				for _, r := range results {
					if r.Success {
						// Find matching key in config
						keyToUpdate := ""
						if _, ok := cfg.ToolsRaw[r.Tool]; ok {
							keyToUpdate = r.Tool
						} else {
							for k := range cfg.ToolsRaw {
								if strings.HasSuffix(k, ":"+r.Tool) || strings.HasSuffix(k, "/"+r.Tool) {
									keyToUpdate = k
									break
								}
							}
						}

						if keyToUpdate != "" {
							content = config.UpsertToolVersion(content, keyToUpdate, r.NewVersion)
							modified = true
						}
					}
				}

				if modified {
					if writeErr := os.WriteFile(configFile, []byte(content), 0644); writeErr == nil {
						_, _ = config.FormatFile(configFile, false)
					}
				}
			}
		}
	}

	// 2. Update unirtm.lock
	lockPath := env.GetLockFilePath()
	if _, err := os.Stat(lockPath); err == nil {
		lockSvc, err := service.NewLockService(service.LockServiceOptions{
			LockfilePath: lockPath,
		})
		if err == nil {
			lockSvc.SetBackendRegistry(backendRegistry)

			updatedTools := make(map[string]service.ToolSpec)
			for _, r := range results {
				if r.Success {
					backendName := ""
					if cfg != nil {
						if toolCfg, ok := cfg.Tools[r.Tool]; ok {
							backendName = toolCfg.Backend
						} else {
							for k, v := range cfg.Tools {
								if strings.HasSuffix(k, ":"+r.Tool) || strings.HasSuffix(k, "/"+r.Tool) {
									backendName = v.Backend
									// When backend is not explicitly set in the ToolConfig,
									// infer it from the config key prefix (e.g. "npm:eslint" → "npm").
									if backendName == "" {
										if idx := strings.Index(k, ":"); idx != -1 {
											backendName = k[:idx]
										}
									}
									break
								}
							}
						}
					}
					if backendName == "" {
						if strings.Contains(r.Tool, "/") {
							backendName = "github"
						} else {
							backendName = "asdf"
						}
					}
					updatedTools[r.Tool] = service.ToolSpec{
						Name:        r.Tool,
						Version:     r.NewVersion,
						BackendName: backendName,
					}
				}
			}

			if len(updatedTools) > 0 {
				_, _ = lockSvc.Generate(ctx, updatedTools, service.GenerateOptions{
					Platforms: lockfile.StandardPlatforms,
				})
			}
		}
	}
}
