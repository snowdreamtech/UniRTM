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

func init() {
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

Examples:
  # Check all installed tools
  unirtm outdated

  # Check specific tool(s)
  unirtm outdated cli/cli

  # JSON output
  unirtm outdated --json`,
	RunE: runOutdated,
}

// outdatedResult holds the comparison result for one installed tool.
type outdatedResult struct {
	Tool        string `json:"tool"`
	Backend     string `json:"backend"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	Outdated    bool   `json:"outdated"`
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

	spinner, _ := pterm.DefaultSpinner.Start("Checking for newer versions...")

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
