// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var (
	// pruneYes skips confirmation prompt
	pruneYes bool
	// pruneTool limits prune to a specific tool
	pruneTool string
)

// init registers the prune command to the root command.
func init() {
	pruneCmd.Flags().BoolVarP(&pruneYes, "yes", "y", false, "skip confirmation prompt")
	pruneCmd.Flags().StringVarP(&pruneTool, "tool", "t", "", "limit prune to a specific tool")

	if rootCmd != nil {
		rootCmd.AddCommand(pruneCmd)
	}
}

// pruneCmd represents the prune command which removes unused tool installations.
// Unused means: not the latest installed version of that tool.
// Equivalent to `mise prune`.
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unused (non-latest) tool installations and show freed disk space",
	Long: `Remove unused tool installations to free up disk space.

The prune command identifies tool versions that are installed but are not
the latest version of their tool, then removes them. The latest installed
version of each tool is always kept.

After pruning, the total amount of freed disk space is reported.

Examples:
  # Preview what would be pruned (dry-run)
  unirtm prune --dry-run

  # Prune all unused versions (with confirmation)
  unirtm prune

  # Prune without confirmation
  unirtm prune --yes

  # Prune only node versions
  unirtm prune --tool node`,
	Args: cobra.NoArgs,
	RunE: runPrune,
}

// runPrune executes the prune command.
// It finds all non-latest tool installations and removes them,
// then reports the total disk space freed.
func runPrune(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()

	// Open database
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return fmt.Errorf("create installation repository: %w", err)
	}

	// List all installations
	installations, err := installRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list installations: %w", err)
	}

	// Filter by tool if specified
	if pruneTool != "" {
		filtered := installations[:0]
		for _, inst := range installations {
			if inst.Tool == pruneTool {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	// Group by tool; keep only the latest (last in list) per tool
	latestByTool := make(map[string]string) // tool → version
	for _, inst := range installations {
		latestByTool[inst.Tool] = inst.Version
	}

	// Collect candidates for pruning (not the latest)
	type pruneCandidate struct {
		tool        string
		version     string
		installPath string
		sizeBytes   int64
	}
	var candidates []pruneCandidate
	for _, inst := range installations {
		if inst.Version != latestByTool[inst.Tool] {
			sz, _ := dirSize(inst.InstallPath)
			candidates = append(candidates, pruneCandidate{
				tool:        inst.Tool,
				version:     inst.Version,
				installPath: inst.InstallPath,
				sizeBytes:   sz,
			})
		}
	}

	if len(candidates) == 0 {
		formatter.Info("No unused installations found — nothing to prune", nil)
		return nil
	}

	// Calculate total size to be freed
	var totalBytes int64
	for _, c := range candidates {
		totalBytes += c.sizeBytes
	}

	// Show a rich table of what will be pruned
	pterm.DefaultSection.Printfln("Found %d installation(s) to prune  (~%s to free)", len(candidates), formatBytes(totalBytes))

	tableData := pterm.TableData{
		{"Tool", "Version", "Install Path", "Size"},
	}
	for _, c := range candidates {
		tableData = append(tableData, []string{
			pterm.FgYellow.Sprint(c.tool),
			c.version,
			c.installPath,
			pterm.FgCyan.Sprint(formatBytes(c.sizeBytes)),
		})
	}
	_ = pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("  ").
		WithData(tableData).
		Render()

	fmt.Printf("\n  Total to free: %s\n\n", pterm.FgRed.Sprint(formatBytes(totalBytes)))

	if dryRun {
		formatter.Info("[dry-run] No changes made", nil)
		return nil
	}

	// Confirm unless --yes
	if !pruneYes {
		fmt.Print("Proceed with pruning? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			formatter.Info("Aborted", nil)
			return nil
		}
	}

	// Prune with a progress bar
	progressBar, _ := pterm.DefaultProgressbar.
		WithTotal(len(candidates)).
		WithTitle("Pruning").
		Start()

	pruned := 0
	var freedBytes int64
	for _, c := range candidates {
		progressBar.UpdateTitle(fmt.Sprintf("Removing %s@%s", c.tool, c.version))

		// Remove from database
		if err := installRepo.Delete(ctx, c.tool, c.version); err != nil {
			formatter.Warning(fmt.Sprintf("Failed to remove DB record for %s@%s: %v", c.tool, c.version, err), nil)
			progressBar.Increment()
			continue
		}
		// Remove install directory from filesystem
		if c.installPath != "" {
			if err := os.RemoveAll(c.installPath); err != nil {
				formatter.Warning(fmt.Sprintf("Failed to remove files for %s@%s: %v", c.tool, c.version, err), nil)
			} else {
				freedBytes += c.sizeBytes
			}
		}
		pruned++
		progressBar.Increment()
	}

	progressBar.Stop()

	formatter.Success(
		fmt.Sprintf("Pruned %d installation(s)  ·  freed %s", pruned, pterm.FgGreen.Sprint(formatBytes(freedBytes))),
		map[string]interface{}{
			"pruned":       pruned,
			"freed_bytes":  freedBytes,
			"freed_human":  formatBytes(freedBytes),
		},
	)
	return nil
}
