// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

// init registers the which command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(whichCmd)
	}
}

// whichCmd represents the which command which prints the full path to the binary
// of a specific tool. Equivalent to `mise which <tool>`.
var whichCmd = &cobra.Command{
	Use:   "which <tool> [version]",
	Short: "Display the path to the tool binary",
	Long: `Display the full path to the binary of a tool.

The which command searches the install_path of the tool in the database,
then looks for the binary in common bin/ sub-directories.

Examples:
  # Show binary path of node (latest installed)
  unirtm which node

  # Show binary path of a specific version
  unirtm which node 20.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runWhich,
}

// runWhich executes the which command.
func runWhich(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()
	target := args[0]
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

	// 1. Try to find as a tool first
	var installations []*repository.Installation
	if version != "" {
		inst, err := installRepo.FindByToolAndVersion(ctx, target, version)
		if err == nil {
			installations = append(installations, inst)
		}
	} else {
		all, err := installRepo.List(ctx)
		if err == nil {
			for _, i := range all {
				if i.Tool == target {
					installations = append(installations, i)
					break
				}
			}
		}
	}

	// 2. If not found or more tools to check, list all installations for binary matching
	if len(installations) == 0 {
		all, err := installRepo.List(ctx)
		if err == nil {
			// In 'which <binary>', we check all installed tools
			installations = all
		}
	}

	for _, inst := range installations {
		p := provider.DefaultRegistry.GetWithBackend(inst.Tool, inst.Backend)
		if p == nil {
			continue
		}

		execs, err := p.ListExecutables(inst.InstallPath, inst.Version)
		if err != nil {
			continue
		}

		for _, exec := range execs {
			// Normalize to absolute path
			absPath := exec
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(inst.InstallPath, exec)
			}

			// Check if this executable matches our target
			if filepath.Base(absPath) == target {
				fmt.Println(absPath)
				return nil
			}
		}

		// Fallback: If we searched by tool name but didn't find a binary with exact name,
		// and the tool only has one primary binary, maybe return that?
		// For example: tool 'maven' provides 'bin/mvn'.
		if inst.Tool == target && len(execs) > 0 {
			absPath := execs[0]
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(inst.InstallPath, execs[0])
			}
			fmt.Println(absPath)
			return nil
		}
	}

	formatter.Error(fmt.Sprintf("Binary for %s not found", target), map[string]interface{}{
		"target": target,
	})
	return fmt.Errorf("binary for %s not found", target)
}
