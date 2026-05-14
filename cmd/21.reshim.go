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
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// reshimTool limits reshim to a specific tool (optional)
	reshimTool string
)

// init registers the reshim command to the root command.
func init() {
	reshimCmd.Flags().StringVarP(&reshimTool, "tool", "t", "", "limit reshim to a specific tool")
	if rootCmd != nil {
		rootCmd.AddCommand(reshimCmd)
	}
}

// reshimCmd represents the reshim command which regenerates all shim scripts.
// Equivalent to `mise reshim`.
var reshimCmd = &cobra.Command{
	Use:   "reshim",
	Short: "Regenerate shim scripts for installed tools",
	Long: `Regenerate shim scripts for all installed tools.

The reshim command recreates the shim scripts in the shims directory.
This is useful when shims become outdated after a manual change, after
modifying the shims directory, or after updating UniRTM itself.

Examples:
  # Regenerate all shims
  unirtm reshim

  # Regenerate shims for a specific tool
  unirtm reshim --tool node`,
	Args: cobra.NoArgs,
	RunE: runReshim,
}

// runReshim executes the reshim command.
// It queries all installed tools and regenerates their shim scripts.
func runReshim(cmd *cobra.Command, args []string) error {
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
	if reshimTool != "" {
		filtered := installations[:0]
		for _, inst := range installations {
			if inst.Tool == reshimTool {
				filtered = append(filtered, inst)
			}
		}
		installations = filtered
	}

	if len(installations) == 0 {
		formatter.Info("No installed tools found — nothing to reshim", nil)
		return nil
	}

	shimsDir := env.GetShimsDir()
	dataDir := env.GetDataDir()
	generator := service.NewGenerator(shimsDir, dataDir+"/installs")
	providerRegistry := provider.NewRegistry()
	shimCount := 0

	for _, inst := range installations {
		if dryRun {
			formatter.Info(fmt.Sprintf("[dry-run] Would regenerate shim for %s", inst.Tool), nil)
			continue
		}

		p := providerRegistry.GetWithBackend(inst.Tool, inst.Backend)
		executables, err := p.ListExecutables(inst.InstallPath, inst.Version)
		if err != nil {
			executables = []string{inst.Tool}
		}

		if err := generator.GenerateShim(ctx, inst.Tool, executables...); err != nil {
			formatter.Warning(fmt.Sprintf("Failed to generate shim for %s@%s: %v", inst.Tool, inst.Version, err), nil)
			continue
		}
		shimCount++
		if verbose {
			formatter.Info(fmt.Sprintf("Regenerated shim for %s@%s", inst.Tool, inst.Version), map[string]interface{}{
				"tool":    inst.Tool,
				"version": inst.Version,
			})
		}
	}

	if !dryRun {
		formatter.Success(fmt.Sprintf("Reshimmed %d tool installation(s) in %s", shimCount, shimsDir), map[string]interface{}{
			"count":     shimCount,
			"shims_dir": shimsDir,
		})
	}

	return nil
}
