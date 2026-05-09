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

	var installPath string
	var inst *repository.Installation
	if version != "" {
		var err error
		inst, err = installRepo.FindByToolAndVersion(ctx, tool, version)
		if err != nil {
			formatter.Error(fmt.Sprintf("Tool %s@%s is not installed", tool, version), map[string]interface{}{
				"tool": tool, "version": version, "error": err.Error(),
			})
			return fmt.Errorf("tool %s@%s not found: %w", tool, version, err)
		}
		installPath = inst.InstallPath
	} else {
		// Find latest installed version
		installations, err := installRepo.List(ctx)
		if err != nil {
			return fmt.Errorf("list installations: %w", err)
		}
		for _, i := range installations {
			if i.Tool == tool {
				inst = i
				installPath = i.InstallPath
			}
		}
		if installPath == "" {
			formatter.Error(fmt.Sprintf("Tool %s is not installed", tool), map[string]interface{}{"tool": tool})
			return fmt.Errorf("tool %s is not installed", tool)
		}
	}

	// Search for binary in common locations within installPath
	binaryName := tool
	candidates := []string{
		filepath.Join(installPath, "bin", binaryName),
		filepath.Join(installPath, binaryName),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			fmt.Println(candidate)
			return nil
		}
	}

	// Also check the shims dir as fallback
	shimPath := filepath.Join(env.GetShimsDir(), binaryName)
	if info, err := os.Stat(shimPath); err == nil && !info.IsDir() {
		fmt.Println(shimPath)
		return nil
	}

	// Fallback to Provider's ListExecutables
	p := provider.DefaultRegistry.GetWithBackend(inst.Tool, inst.Backend)
	execs, err := p.ListExecutables(installPath, inst.Version)
	if err == nil && len(execs) > 0 {
		fmt.Println(execs[0])
		return nil
	}

	formatter.Error(fmt.Sprintf("Binary for %s not found in %s", tool, installPath), map[string]interface{}{
		"tool":         tool,
		"install_path": installPath,
	})
	return fmt.Errorf("binary for %s not found in %s", tool, installPath)
}
