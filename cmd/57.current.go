// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(currentCmd)
	}
}

// currentCmd represents the current command which shows active tool versions.
var currentCmd = &cobra.Command{
	Use:   "current [tool]",
	Short: "Display the active version for each tool",
	Long: `Display the active version for each tool.

If no tool is specified, it shows the active version for all installed tools.
A version is considered active if its shim currently points to it or if it is
the resolved version from the active configuration.

Examples:
  # Show all active versions
  unirtm current

  # Show active version for node
  unirtm current node`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
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
		return err
	}
	defer db.Close()

	installRepo, _ := sqlite.NewInstallationRepository(db.Conn())
	installations, err := installRepo.List(ctx)
	if err != nil {
		return err
	}

	if len(installations) == 0 {
		if !quiet {
			formatter.Info("No tools installed", nil)
		}
		return nil
	}

	// Resolve active versions (logic borrowed from list command)
	shimsDir := env.GetShimsDir()
	activeVersions := resolveActiveVersions(shimsDir, installations)

	// Filter by tool if specified
	if len(args) == 1 {
		toolName := args[0]
		version, ok := activeVersions[toolName]
		if !ok {
			// Check if tool exists at all
			found := false
			for _, inst := range installations {
				if inst.Tool == toolName {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("tool not found: %s", toolName)
			}
			if !quiet {
				fmt.Printf("%s: (none active)\n", toolName)
			}
			return nil
		}
		if jsonOutput {
			formatter.Success(toolName, map[string]interface{}{"tool": toolName, "version": version})
		} else {
			fmt.Println(version)
		}
		return nil
	}

	// Show all
	tools := make([]string, 0, len(activeVersions))
	for t := range activeVersions {
		tools = append(tools, t)
	}
	sort.Strings(tools)

	if jsonOutput {
		results := make(map[string]interface{})
		for _, t := range tools {
			results[t] = activeVersions[t]
		}
		formatter.Success("Current active versions", results)
		return nil
	}

	for _, t := range tools {
		fmt.Printf("%-20s %s\n", t, activeVersions[t])
	}

	return nil
}

// NOTE: resolveActiveVersions is already defined in 8.list.go.
// In Go, since both files are in the 'cmd' package, they can share it.
// If not exported, they must be in the same package, which they are.
