// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

	// 1. Load merged configuration (hierarchy)
	cfg, err := config.LoadFull()
	if err != nil {
		// If no config found, it's not necessarily an error, mise returns empty
		cfg = &config.Config{}
	}

	// 2. Open database to find installations
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

	// 3. Collect paths
	shimsDir := env.GetShimsDir()
	paths := []string{shimsDir}
	seen := make(map[string]bool)
	seen[shimsDir] = true

	// Get sorted tool names to ensure deterministic output
	var toolNames []string
	for name := range cfg.Tools {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	// Iterate over tools defined in current config
	im, _ := getInstallationManager(ctx, cfg)
	for _, toolNameKey := range toolNames {
		toolCfg := cfg.Tools[toolNameKey]
		_, toolName, version, _ := im.ParseToolSpec(toolNameKey)
		if toolCfg.Backend != "" {
			// Backend in config is ignored if we already have it in DB,
			// but we keep the logic consistent.
		}
		if toolCfg.Version != "" {
			version = toolCfg.Version
		}

		// Check environment variable override: <PREFIX>_<TOOL>_VERSION
		toolKey := strings.ToUpper(strings.ReplaceAll(toolName, "-", "_")) + "_VERSION"
		if v := env.Get(toolKey); v != "" {
			version = v
		}

		// Find installation for this tool and version
		inst, err := installRepo.FindByToolAndVersion(ctx, toolName, version)
		if err != nil || inst == nil {
			// Try without v prefix as fallback
			v2 := version
			if strings.HasPrefix(v2, "v") {
				v2 = v2[1:]
			} else {
				v2 = "v" + v2
			}
			inst, _ = installRepo.FindByToolAndVersion(ctx, toolName, v2)
			if inst == nil {
				continue // Not installed, skip
			}
		}

		// Use provider to get correct bin paths
		p := provider.DefaultRegistry.GetWithBackend(inst.Tool, inst.Backend)
		if p == nil {
			continue
		}

		binPaths, err := p.GetBinPaths(inst.Tool, inst.InstallPath, inst.Version)
		if err != nil {
			continue
		}

		for _, bp := range binPaths {
			if !seen[bp] {
				if _, statErr := os.Stat(bp); statErr == nil {
					paths = append(paths, bp)
					seen[bp] = true
				}
			}
		}
	}

	isTerminal := term.IsTerminal(int(os.Stdout.Fd())) && !jsonOutput
	if isTerminal {
		pterm.DefaultHeader.
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightMagenta)).
			WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
			WithMargin(10).
			Println("UniRTM Active Bin Paths")
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{"paths": paths})
	}

	for _, p := range paths {
		if isTerminal {
			pterm.Println(pterm.FgGray.Sprint(p))
		} else {
			fmt.Println(p)
		}
	}
	return nil
}
