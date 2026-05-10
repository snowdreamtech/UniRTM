// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var (
	// listToolFilter filters list output by tool name.
	listToolFilter string
)

// init registers the list command to the root command.
func init() {
	listCmd.Flags().StringVarP(&listToolFilter, "tool", "t", "", "filter by tool name")
	if rootCmd != nil {
		rootCmd.AddCommand(listCmd)
	}
}

// listCmd represents the list command which shows all installed tools.
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all installed development tools",
	Long: `List all installed development tools.

Shows all tools installed with UniRTM, their version, backend, activation
status, and installation path. The STATUS column shows whether a version
is currently active (its shim points to it).

Examples:
  # List all installed tools
  unirtm list

  # Filter by tool name
  unirtm list --tool node

  # JSON output
  unirtm list --json`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    runList,
}

// runList executes the list command.
func runList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
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
		formatter.Error("Failed to create installation repository", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	installations, err := installRepo.List(ctx)
	if err != nil {
		formatter.Error("Failed to list installations", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("list installations: %w", err)
	}

	// Apply --tool filter.
	if listToolFilter != "" {
		filtered := installations[:0]
		for _, inst := range installations {
			if inst.Tool == listToolFilter {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	if len(installations) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			formatter.Info("No tools installed", nil)
		}
		return nil
	}

	// Resolve active version per tool via shim symlinks.
	shimsDir := env.GetShimsDir()
	activeVersions := resolveActiveVersions(shimsDir, installations)

	// JSON output.
	if jsonOutput {
		type jsonEntry struct {
			Tool        string    `json:"tool"`
			Version     string    `json:"version"`
			Backend     string    `json:"backend"`
			Status      string    `json:"status"`
			InstallPath string    `json:"install_path"`
			InstalledAt time.Time `json:"installed_at"`
		}
		results := make([]jsonEntry, 0, len(installations))
		for _, inst := range installations {
			status := "installed"
			if activeVersions[inst.Tool] == inst.Version {
				status = "active"
			}
			results = append(results, jsonEntry{
				Tool:        inst.Tool,
				Version:     inst.Version,
				Backend:     inst.Backend,
				Status:      status,
				InstallPath: inst.InstallPath,
				InstalledAt: inst.InstalledAt,
			})
		}
		formatter.Success("Installed tools", map[string]interface{}{
			"count": len(results),
			"tools": results,
		})
		return nil
	}

	// Human-readable table with STATUS column.
	tableData := pterm.TableData{
		{"TOOL", "VERSION", "STATUS", "BACKEND", "INSTALLED AT"},
	}

	for _, inst := range installations {
		statusColored := pterm.FgDefault.Sprint("─")
		if activeVersions[inst.Tool] == inst.Version {
			statusColored = pterm.FgGreen.Sprint("active ✓")
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(inst.Tool),
			pterm.FgYellow.Sprint(inst.Version),
			statusColored,
			pterm.FgMagenta.Sprint(inst.Backend),
			inst.InstalledAt.Format("2006-01-02"),
		})
	}

	fmt.Println()
	pterm.EnableColor()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	return nil
}

// resolveActiveVersions inspects shim symlinks to determine which version of
// each tool is currently active.  If a shim exists and is a symlink that
// points into a particular version's install directory, that version is active.
// Falls back to the most-recently-installed version when no symlink evidence.
func resolveActiveVersions(shimsDir string, installations []*repository.Installation) map[string]string {
	active := make(map[string]string)
	installsDir := env.GetInstallsDir()

	for _, inst := range installations {
		if inst == nil {
			continue
		}
		// Find any binary in the install bin dir and check if its shim resolves here.
		binDir := filepath.Join(installsDir, inst.Tool, inst.Version, "bin")
		entries, err := os.ReadDir(binDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			shimPath := filepath.Join(shimsDir, e.Name())
			target, err := os.Readlink(shimPath)
			if err != nil {
				continue
			}
			// If the symlink target contains this version's path, mark it active.
			if filepath.IsAbs(target) {
				if isPathUnder(target, filepath.Join(installsDir, inst.Tool, inst.Version)) {
					active[inst.Tool] = inst.Version
					break
				}
			}
		}
	}
	return active
}

// isPathUnder reports whether path is inside or equal to dir.
func isPathUnder(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return len(rel) > 0 && rel[0] != '.'
}
