// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

// init registers the where command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(whereCmd)
	}
}

// whereCmd represents the where command which prints the installation directory
// of a specific tool version. Equivalent to `mise where <tool> [version]`.
var whereCmd = &cobra.Command{
	Use:   "where <tool> [version]",
	Short: "Display the installation path of a tool",
	Long: `Display the installation path of a specific tool version.

The where command prints the directory where a tool version is installed.
If no version is specified, it returns the path of the latest installed version.

Examples:
  # Show installation path of node (latest installed)
  unirtm where node

  # Show installation path of a specific version
  unirtm where node 20.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runWhere,
}

// runWhere executes the where command.
// It queries the database for the installation path of the specified tool.
func runWhere(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()
	tool := args[0]
	var version string
	if len(args) == 2 {
		version = args[1]
	}

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

	if version != "" {
		// Find specific version
		inst, err := installRepo.FindByToolAndVersion(ctx, tool, version)
		if err != nil {
			formatter.Error(fmt.Sprintf("Tool %s@%s is not installed", tool, version), map[string]interface{}{
				"tool":    tool,
				"version": version,
				"error":   err.Error(),
			})
			return fmt.Errorf("tool %s@%s not found: %w", tool, version, err)
		}
		fmt.Println(inst.InstallPath)
		return nil
	}

	// Find latest installed version of tool
	installations, err := installRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list installations: %w", err)
	}

	var found *struct{ installPath string }
	for _, inst := range installations {
		if inst.Tool == tool {
			found = &struct{ installPath string }{inst.InstallPath}
			// Continue to get the last (latest) entry
		}
	}

	if found == nil {
		formatter.Error(fmt.Sprintf("Tool %s is not installed", tool), map[string]interface{}{
			"tool": tool,
		})
		return fmt.Errorf("tool %s is not installed", tool)
	}

	fmt.Println(found.installPath)
	return nil
}
