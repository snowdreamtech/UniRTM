// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
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
	unuseAll bool
)

func init() {
	unuseCmd.Flags().BoolVar(&unuseAll, "all", false, "remove all versions of the tool from the config")
	if rootCmd != nil {
		rootCmd.AddCommand(unuseCmd)
	}
}

// unuseCmd removes a tool entry from unirtm.toml (does not uninstall files).
var unuseCmd = &cobra.Command{
	Use:     "unuse <tool[@version]>...",
	Short:   "Remove tools from the config file without uninstalling",
	Aliases: []string{"rm", "remove"},
	Long: `Remove tools from the config file without uninstalling them.

Deletes the specified tool(s) from the [tools] section of unirtm.toml
and removes the corresponding database record. The installed files are
kept on disk — use 'unirtm uninstall' to also delete the binaries.

Examples:
  # Remove a specific version
  unirtm unuse cli/cli@2.70.0

  # Remove a tool (removes its DB record regardless of version)
  unirtm unuse node

  # Aliases
  unirtm rm cli/cli
  unirtm remove node`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUnuse,
}

func runUnuse(cmd *cobra.Command, args []string) error {
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
		formatter.Error(fmt.Sprintf("Failed to open database: %v", err))
		return err
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to create repository: %v", err))
		return err
	}

	// Also update [tools] section of config file.
	cfgPath := resolveConfigFilePath(false)
	cfgMap, _ := loadRawTOML(cfgPath)
	toolsSection, _ := cfgMap["tools"].(map[string]interface{})
	if toolsSection == nil {
		toolsSection = make(map[string]interface{})
		cfgMap["tools"] = toolsSection
	}

	cfgChanged := false

	for _, arg := range args {
		tool := arg
		version := ""

		// Parse tool@version syntax.
		if idx := strings.Index(arg, "@"); idx != -1 {
			tool = arg[:idx]
			version = arg[idx+1:]
		}

		if version != "" {
			// Delete specific version from DB.
			if err := installRepo.Delete(ctx, tool, version); err != nil {
				formatter.Warning(fmt.Sprintf("Could not remove %s@%s from DB: %v", tool, version, err))
			} else {
				formatter.Success(fmt.Sprintf("Removed %s@%s from database", tool, version), nil)
			}
		} else {
			// Delete all versions for this tool from DB.
			all, _ := installRepo.List(ctx)
			removed := 0
			for _, inst := range all {
				if inst == nil || inst.Tool != tool {
					continue
				}
				if err := installRepo.Delete(ctx, inst.Tool, inst.Version); err != nil {
					formatter.Warning(fmt.Sprintf("Could not remove %s@%s: %v", tool, inst.Version, err))
				} else {
					removed++
				}
			}
			if removed > 0 {
				formatter.Success(fmt.Sprintf("Removed %d version(s) of %s from database", removed, tool), nil)
			} else {
				formatter.Warning(fmt.Sprintf("No database records found for %s", tool))
			}
		}

		// Remove from [tools] section of config.
		if _, ok := toolsSection[tool]; ok {
			delete(toolsSection, tool)
			cfgChanged = true
			formatter.Success(fmt.Sprintf("Removed %s from [tools] in %s", tool, cfgPath), nil)
		}
	}

	// Write config if changed.
	if cfgChanged {
		if len(toolsSection) == 0 {
			delete(cfgMap, "tools")
		}
		if err := saveRawTOML(cfgPath, cfgMap); err != nil {
			formatter.Warning(fmt.Sprintf("Could not update %s: %v", cfgPath, err))
		}
	}

	return nil
}
