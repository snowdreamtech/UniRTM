// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(binPathsCmd)
	}
}

// binPathsCmd lists all active runtime bin directories.
var binPathsCmd = &cobra.Command{
	Use:     "bin-paths",
	Aliases: []string{"bin"},
	Short:   "List all active runtime bin directories",
	Long: `List all active runtime bin directories.

Outputs one directory per line — shims dir first, then each installed
tool's bin directory. Useful for shell hook scripts that need to prepend
the correct directories to PATH.

Examples:
  # Print all bin paths (one per line)
  unirtm bin-paths

  # JSON output
  unirtm bin-paths --json

  # Use in a shell script
  export PATH="$(unirtm bin-paths | tr '\n' ':')$PATH"`,
	Args: cobra.NoArgs,
	RunE: runBinPaths,
}

func runBinPaths(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Always include shims dir first.
	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()

	paths := []string{shimsDir}

	// Add each installed tool's bin dir (if it exists on disk).
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err == nil {
		defer db.Close()
		if repo, err := sqlite.NewInstallationRepository(db.Conn()); err == nil {
			if installations, err := repo.List(ctx); err == nil {
				seen := make(map[string]bool)
				for _, inst := range installations {
					if inst == nil {
						continue
					}
					binDir := filepath.Join(installsDir, inst.Tool, inst.Version, "bin")
					if !seen[binDir] {
						if _, statErr := os.Stat(binDir); statErr == nil {
							paths = append(paths, binDir)
							seen[binDir] = true
						}
					}
				}
			}
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{"paths": paths})
	}

	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}
