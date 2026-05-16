// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/config"
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
	ctx := context.Background()
	cfg, _ := config.LoadFull()
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		return err
	}

	// 1. Parse input: could be "node" or "node@20"
	input := args[0]
	var backendName, toolName, version string
	
	if len(args) == 2 {
		// Old style: unirtm where node 20
		toolName = args[0]
		version = args[1]
	} else {
		// New style: unirtm where node@20
		_, toolName, version, _ = im.ParseToolSpec(input)
	}

	// 2. Resolve the version to a concrete installation path
	// If version is "latest" or empty, try to resolve from current context
	if version == "" || version == "latest" {
		// Try to find what's active in the current directory first
		if cfg != nil {
			if toolSpec, ok := cfg.Tools[toolName]; ok {
				version = toolSpec.Version
				if toolSpec.Backend != "" {
					backendName = toolSpec.Backend
				}
			}
		}
	}

	// 3. Fallback to database if version is still missing
	if version == "" || version == "latest" {
		dbPath := env.GetDatabasePath()
		db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
		if err == nil {
			defer db.Close()
			installRepo, err := sqlite.NewInstallationRepository(db.Conn())
			if err == nil {
				installations, err := installRepo.List(ctx)
				if err == nil {
					for _, inst := range installations {
						if inst.Tool == toolName {
							version = inst.Version
						}
					}
				}
			}
		}
	}

	if version == "" {
		version = "latest"
	}

	// If still empty, we fallback to AutoDetectBackend and resolve
	if backendName == "" {
		backendName = im.AutoDetectBackend(toolName)
	}

	// 4. Resolve to absolute path
	// Compute the expected installation path.
	fsName := env.GetFSToolName(toolName, backendName)
	installPath := filepath.Join(env.GetInstallsDir(), fsName, version)

	if _, err := os.Stat(installPath); err != nil {
		return fmt.Errorf("tool %s@%s is not installed", toolName, version)
	}

	fmt.Println(installPath)
	return nil
}
