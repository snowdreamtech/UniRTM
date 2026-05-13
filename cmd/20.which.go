// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

type match struct {
	Path     string
	Version  string
	Tool     string
	Provider string
	Active   bool
	Source   string
}

var (
	whichAll      bool
	whichTool     string
	whichVersion  bool
	whichProvider bool
)

// init registers the which command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(whichCmd)
	}
	whichCmd.Flags().BoolVarP(&whichAll, "all", "a", false, "Show all matches")
	whichCmd.Flags().StringVarP(&whichTool, "tool", "t", "", "Filter by specific tool")
	whichCmd.Flags().BoolVarP(&whichVersion, "version", "V", false, "Show version instead of path")
	whichCmd.Flags().BoolVarP(&whichProvider, "provider", "p", false, "Show provider instead of path")
	whichCmd.Flags().BoolVar(&whichProvider, "plugin", false, "Alias for --provider")
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx := context.Background()
		var candidates []string

		// 1. From Config
		configMgr := config.NewConfigManager()
		if cfg, err := configMgr.LoadHierarchy(ctx); err == nil {
			for name := range cfg.Tasks {
				candidates = append(candidates, name)
			}
			for name := range cfg.Tools {
				candidates = append(candidates, name)
			}
		}

		// 2. From Database
		dbPath := env.GetDatabasePath()
		if db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true}); err == nil {
			defer db.Close()
			if installRepo, err := sqlite.NewInstallationRepository(db.Conn()); err == nil {
				if all, err := installRepo.List(ctx); err == nil {
					for _, inst := range all {
						candidates = append(candidates, inst.Tool)
					}
				}
			}
		}

		// De-duplicate
		unique := make(map[string]struct{})
		var final []string
		for _, c := range candidates {
			if _, ok := unique[c]; !ok {
				unique[c] = struct{}{}
				final = append(final, c)
			}
		}

		return final, cobra.ShellCompDirectiveNoFileComp
	},
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
	var versionArg string
	if len(args) == 2 {
		versionArg = args[1]
	}

	// 1. Load Configuration
	configMgr := config.NewConfigManager()
	cfg, err := configMgr.LoadHierarchy(ctx)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// 2. Try to find as a Task first (Surpassing mise!)
	if !whichVersion && !whichProvider && whichTool == "" {
		if task, ok := cfg.Tasks[target]; ok {
			if verbose {
				fmt.Printf("Task:        %s\n", target)
				fmt.Printf("Description: %s\n", task.Description)
				fmt.Printf("Run:         %s\n", task.Run)
				if len(task.Depends) > 0 {
					fmt.Printf("Depends:     %s\n", strings.Join(task.Depends, ", "))
				}
				return nil
			}
			// Default output for task could be the run command or just a message
			fmt.Printf("Task: %s (run: %s)\n", target, task.Run)
			return nil
		}
	}

	// 3. Open database to find installations
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

	// 4. Resolve Active Version
	// Logic: Argument version > Environment variable > Config file > Latest installed
	activeVersions := make(map[string]string)
	if versionArg != "" {
		activeVersions[target] = versionArg
	} else {
		// Check environment variable UNIRTM_<TOOL>_VERSION
		envVar := fmt.Sprintf("UNIRTM_%s_VERSION", strings.ToUpper(strings.ReplaceAll(target, "-", "_")))
		if v := os.Getenv(envVar); v != "" {
			activeVersions[target] = v
		} else if tc, ok := cfg.Tools[target]; ok {
			activeVersions[target] = tc.Version
		}
	}

	allInstallations, err := installRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list installations: %w", err)
	}

	var matches []match

	for _, inst := range allInstallations {
		// Filter by tool if --tool is specified
		if whichTool != "" && inst.Tool != whichTool {
			continue
		}

		p := provider.DefaultRegistry.GetWithBackend(inst.Tool, inst.Backend)
		if p == nil {
			continue
		}

		execs, err := p.ListExecutables(inst.InstallPath, inst.Version)
		if err != nil {
			continue
		}

		isActive := false
		source := "Installed"
		if v, ok := activeVersions[inst.Tool]; ok && v == inst.Version {
			isActive = true
			source = "Active"
		}

		for _, exec := range execs {
			absPath := exec
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(inst.InstallPath, exec)
			}

			if filepath.Base(absPath) == target || (inst.Tool == target && exec == execs[0]) {
				matches = append(matches, match{
					Path:     absPath,
					Version:  inst.Version,
					Tool:     inst.Tool,
					Provider: p.Name(),
					Active:   isActive,
					Source:   source,
				})
				// If not --all and we found an active one, we're done
				if !whichAll && isActive {
					printMatch(matches[len(matches)-1])
					return nil
				}
			}
		}
	}

	if len(matches) > 0 {
		// If not --all, print the first one (prefer active)
		if !whichAll {
			for _, m := range matches {
				if m.Active {
					printMatch(m)
					return nil
				}
			}
			printMatch(matches[0])
			return nil
		}

		// Print all matches
		for _, m := range matches {
			printMatch(m)
		}
		return nil
	}

	// 5. Helpful error message with fuzzy suggestions (Surpassing mise!)
	msg := fmt.Sprintf("Binary or task for %s not found", target)
	formatter.Error(msg, map[string]interface{}{
		"target": target,
	})

	// Gather all possible targets for fuzzy matching
	var candidates []string
	for name := range cfg.Tasks {
		candidates = append(candidates, name)
	}
	for name := range cfg.Tools {
		candidates = append(candidates, name)
	}
	// Also add installed tools that might not be in the current config
	for _, inst := range allInstallations {
		candidates = append(candidates, inst.Tool)
	}

	// Remove duplicates
	uniqueCandidates := make(map[string]struct{})
	var finalCandidates []string
	for _, c := range candidates {
		if _, ok := uniqueCandidates[c]; !ok {
			uniqueCandidates[c] = struct{}{}
			finalCandidates = append(finalCandidates, c)
		}
	}

	// Find close matches using common utility
	output.Suggest(os.Stderr, target, finalCandidates)

	// Check if it exists as a tool but no installation
	if _, ok := cfg.Tools[target]; ok {
		fmt.Fprintf(os.Stderr, "\nTip: %s is defined in your config but not installed. Run 'unirtm install' to install it.\n", target)
	}

	return fmt.Errorf("not found: %s", target)
}

func printMatch(m match) {
	if whichVersion {
		fmt.Println(m.Version)
		return
	}
	if whichProvider {
		fmt.Println(m.Provider)
		return
	}
	if verbose {
		fmt.Printf("%-15s %s\n", "Path:", m.Path)
		fmt.Printf("%-15s %s\n", "Version:", m.Version)
		fmt.Printf("%-15s %s\n", "Tool:", m.Tool)
		fmt.Printf("%-15s %s\n", "Provider:", m.Provider)
		fmt.Printf("%-15s %s\n", "Source:", m.Source)
		return
	}
	fmt.Println(m.Path)
}
