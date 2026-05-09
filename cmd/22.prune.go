// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
	Short: "Remove unused (non-latest) tool installations",
	Long: `Remove unused tool installations to free up disk space.

The prune command identifies tool versions that are installed but are not
the latest version of their tool, then removes them. The latest installed
version of each tool is always kept.

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
// It finds all non-latest tool installations and removes them.
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
	}
	var candidates []pruneCandidate
	for _, inst := range installations {
		if inst.Version != latestByTool[inst.Tool] {
			candidates = append(candidates, pruneCandidate{
				tool:        inst.Tool,
				version:     inst.Version,
				installPath: inst.InstallPath,
			})
		}
	}

	if len(candidates) == 0 {
		formatter.Info("No unused installations found — nothing to prune", nil)
		return nil
	}

	// Show what will be pruned
	formatter.Info(fmt.Sprintf("Found %d installation(s) to prune:", len(candidates)), nil)
	for _, c := range candidates {
		fmt.Printf("  - %s@%s  (%s)\n", c.tool, c.version, c.installPath)
	}

	if dryRun {
		formatter.Info("[dry-run] No changes made", nil)
		return nil
	}

	// Confirm unless --yes
	if !pruneYes {
		fmt.Print("\nProceed with pruning? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			formatter.Info("Aborted", nil)
			return nil
		}
	}

	// Create prune records
	pruned := 0
	for _, c := range candidates {
		// Remove from database
		if err := installRepo.Delete(ctx, c.tool, c.version); err != nil {
			formatter.Warning(fmt.Sprintf("Failed to remove DB record for %s@%s: %v", c.tool, c.version, err), nil)
			continue
		}
		// Remove install directory from filesystem
		if c.installPath != "" {
			if err := os.RemoveAll(c.installPath); err != nil {
				formatter.Warning(fmt.Sprintf("Failed to remove files for %s@%s: %v", c.tool, c.version, err), nil)
			}
		}
		pruned++
		formatter.Info(fmt.Sprintf("Removed %s@%s", c.tool, c.version), nil)
	}

	formatter.Success(fmt.Sprintf("Pruned %d installation(s)", pruned), map[string]interface{}{
		"pruned": pruned,
	})
	return nil
}
