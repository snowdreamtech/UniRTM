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
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var (
	outdatedInteractive bool
)

func init() {
	outdatedCmd.Flags().BoolVarP(&outdatedInteractive, "interactive", "i", false, "interactively select tools to upgrade")
	if rootCmd != nil {
		rootCmd.AddCommand(outdatedCmd)
	}
}

// outdatedCmd lists tools that have a newer version available.
var outdatedCmd = &cobra.Command{
	Use:   "outdated [tool...]",
	Short: "Show installed tools that have newer versions available",
	Long: `Show installed tools that have newer versions available.

Queries each backend for the latest available version and compares it to
what is currently installed. Useful before running 'unirtm update'.

Use --interactive / -i to select which outdated tools to upgrade immediately.

Examples:
  # Check all installed tools
  unirtm outdated

  # Check specific tool(s)
  unirtm outdated cli/cli

  # Interactively select tools to upgrade
  unirtm outdated --interactive

  # JSON output
  unirtm outdated --json`,
	RunE: runOutdated,
}

// outdatedResult holds the comparison result for one installed tool.
type outdatedResult struct {
	Tool     string `json:"tool"`
	Backend  string `json:"backend"`
	Current  string `json:"current"`
	Latest   string `json:"latest"`
	Outdated bool   `json:"outdated"`
}

func runOutdated(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	// Open DB and load installed tools.
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	installations, err := installRepo.List(ctx)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to list installations: %v", err))
		return err
	}

	// Filter by CLI args if provided.
	if len(args) > 0 {
		want := make(map[string]bool, len(args))
		for _, a := range args {
			want[strings.TrimSpace(a)] = true
		}
		filtered := installations[:0]
		for _, inst := range installations {
			if want[inst.Tool] {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	if len(installations) == 0 {
		formatter.Info("No installed tools found.", nil)
		return nil
	}

	backendRegistry := backend.NewRegistry()
	platform := backend.CurrentPlatform()

	spinner, _ := output.StartSpinner("Checking for newer versions...")

	var results []outdatedResult
	for _, inst := range installations {
		b, err := backendRegistry.Get(inst.Backend)
		if err != nil {
			// Backend not found — skip silently in non-verbose mode.
			if verbose {
				formatter.Warning(fmt.Sprintf("Skipping %s: backend %q not found", inst.Tool, inst.Backend))
			}
			continue
		}

		latest, err := resolveLatestVersion(ctx, b, inst.Tool, platform)
		if err != nil {
			if verbose {
				formatter.Warning(fmt.Sprintf("Could not resolve latest for %s: %v", inst.Tool, err))
			}
			continue
		}

		outdated := latest != "" && latest != inst.Version
		results = append(results, outdatedResult{
			Tool:     inst.Tool,
			Backend:  inst.Backend,
			Current:  inst.Version,
			Latest:   latest,
			Outdated: outdated,
		})
	}

	spinner.Stop()

	// Filter to only outdated tools for display (unless verbose shows all).
	display := results[:0]
	for _, r := range results {
		if r.Outdated || verbose {
			display = append(display, r)
		}
	}

	if len(display) == 0 {
		formatter.Info("All tools are up to date. ✓", nil)
		return nil
	}

	// JSON output.
	if jsonOutput {
		formatter.Success("Outdated tools", map[string]interface{}{
			"count": len(display),
			"tools": display,
		})
		return nil
	}

	// Human-readable table.
	tableData := pterm.TableData{
		{"TOOL", "CURRENT", "LATEST", "BACKEND"},
	}
	for _, r := range display {
		latestStr := pterm.FgGreen.Sprint(r.Latest)
		if !r.Outdated {
			latestStr = pterm.FgDefault.Sprint(r.Latest + " (up to date)")
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(r.Tool),
			pterm.FgYellow.Sprint(r.Current),
			latestStr,
			r.Backend,
		})
	}

	fmt.Println()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	fmt.Println()

	// Interactive upgrade mode.
	if outdatedInteractive {
		// Build options list: only truly outdated tools.
		outdatedOnly := make([]outdatedResult, 0)
		for _, r := range display {
			if r.Outdated {
				outdatedOnly = append(outdatedOnly, r)
			}
		}

		if len(outdatedOnly) == 0 {
			formatter.Info("No outdated tools to upgrade.", nil)
			return nil
		}

		options := make([]string, 0, len(outdatedOnly))
		for _, r := range outdatedOnly {
			options = append(options, fmt.Sprintf("%s  %s → %s",
				pterm.FgCyan.Sprint(r.Tool),
				pterm.FgYellow.Sprint(r.Current),
				pterm.FgGreen.Sprint(r.Latest),
			))
		}

		selectedLabels, err := pterm.DefaultInteractiveMultiselect.
			WithOptions(options).
			WithDefaultOptions(options). // Pre-select all by default.
			WithFilter(false).
			Show("Select tools to upgrade (Space to toggle, Enter to confirm)")
		if err != nil || len(selectedLabels) == 0 {
			formatter.Info("No tools selected for upgrade.", nil)
			return nil
		}

		// Build a set of selected tool names.
		selectedSet := make(map[string]bool)
		for _, label := range selectedLabels {
			for _, r := range outdatedOnly {
				optionLabel := fmt.Sprintf("%s  %s → %s",
					pterm.FgCyan.Sprint(r.Tool),
					pterm.FgYellow.Sprint(r.Current),
					pterm.FgGreen.Sprint(r.Latest),
				)
				if label == optionLabel {
					selectedSet[r.Tool] = true
				}
			}
		}

		// Run install for each selected tool.
		fmt.Println()
		for _, r := range outdatedOnly {
			if !selectedSet[r.Tool] {
				continue
			}
			upgradeSpinner, _ := output.StartSpinner(
				fmt.Sprintf("Upgrading %s %s → %s ...",
					pterm.FgCyan.Sprint(r.Tool),
					pterm.FgYellow.Sprint(r.Current),
					pterm.FgGreen.Sprint(r.Latest),
				),
			)
			// Delegate to the install command runner with tool@latest.
			installErr := runInstall(cmd, []string{fmt.Sprintf("%s@%s", r.Tool, r.Latest)})
			if installErr != nil {
				upgradeSpinner.Fail(fmt.Sprintf("Failed to upgrade %s: %v", r.Tool, installErr))
			} else {
				upgradeSpinner.Success(fmt.Sprintf("Upgraded %s to %s", pterm.FgCyan.Sprint(r.Tool), pterm.FgGreen.Sprint(r.Latest)))
			}
		}
	}

	return nil
}

// resolveLatestVersion calls ResolveVersion("latest") on the backend.
func resolveLatestVersion(ctx context.Context, b backend.Backend, tool string, platform backend.Platform) (string, error) {
	info, err := b.ResolveVersion(ctx, tool, "latest", platform)
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", nil
	}
	return info.Version, nil
}
