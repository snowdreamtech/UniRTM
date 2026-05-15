// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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

This command aligns with 'mise current'. It looks at your configuration files
(unirtm.toml, .tool-versions) to determine which tools and versions are
currently requested in this directory.

If no tool is specified, it shows the active versions for all tools defined
in the current hierarchy.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
	// 1. Load configuration from hierarchy (align with mise)
	cfg, err := config.LoadFull()
	if err != nil {
		return err
	}

	filterTool := ""
	if len(args) == 1 {
		filterTool = args[0]
	}

	// 2. Resolve requested tools and their installation status
	type toolStatus struct {
		name      string
		versions  []string
		installed []bool
	}
	var results []toolStatus

	// Get all tool names from merged config
	var toolNames []string
	for name := range cfg.Tools {
		if filterTool != "" && name != filterTool {
			continue
		}
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	for _, name := range toolNames {
		toolCfg := cfg.Tools[name]
		versions := strings.Fields(toolCfg.Version)
		var installed []bool

		for _, v := range versions {
			// Robust detection engine
			isInst := false
			basePath := filepath.Join(env.GetInstallsDir(), env.GetFSToolName(name, toolCfg.Backend))
			
			// 1. Generate core variants (v, V, none)
			pureVersion := v
			if strings.HasPrefix(strings.ToLower(v), "v") {
				pureVersion = v[1:]
			}
			
			checkVariants := []string{v, pureVersion, "v" + pureVersion, "V" + pureVersion}
			
			// 2. Exact match check
			seen := make(map[string]bool)
			for _, variant := range checkVariants {
				if variant == "" || seen[variant] { continue }
				seen[variant] = true
				if _, err := os.Stat(filepath.Join(basePath, variant)); err == nil {
					isInst = true
					break
				}
			}

			// 3. [Surpass] Fuzzy/Normalized match for dates (2024.05.15 == 2024-05-15)
			if !isInst {
				normalizedReq := strings.NewReplacer("-", "", ".", "", "_", "").Replace(strings.ToLower(pureVersion))
				entries, _ := os.ReadDir(basePath)
				for _, entry := range entries {
					if !entry.IsDir() { continue }
					
					eName := entry.Name()
					pureEName := eName
					if strings.HasPrefix(strings.ToLower(eName), "v") {
						pureEName = eName[1:]
					}
					
					normalizedEntry := strings.NewReplacer("-", "", ".", "", "_", "").Replace(strings.ToLower(pureEName))
					if normalizedReq == normalizedEntry {
						isInst = true
						break
					}
				}
			}
			installed = append(installed, isInst)
		}

		results = append(results, toolStatus{
			name:      name,
			versions:  versions,
			installed: installed,
		})
	}

	if len(results) == 0 {
		if filterTool != "" {
			return fmt.Errorf("tool %s is not defined in current configuration", filterTool)
		}
		pterm.Info.Println("No tools defined in current configuration.")
		return nil
	}

	// 3. Output Logic
	// Surpass: Visual detection. If TTY and not specific tool, show rich UI.
	if !pterm.PrintColor || jsonOutput || filterTool != "" {
		// Plain mode (Mise style) or specific tool
		for _, res := range results {
			if filterTool != "" {
				fmt.Println(strings.Join(res.versions, " "))
			} else {
				fmt.Printf("%s %s\n", res.name, strings.Join(res.versions, " "))
			}
		}
		return nil
	}

	// [Surpass] Interactive UI
	pterm.DefaultSection.Println("Active Runtime Versions")
	
	var tableData [][]string
	tableData = append(tableData, []string{"Tool", "Version", "Status"})

	hasMissing := false
	for _, res := range results {
		for i, v := range res.versions {
			status := pterm.LightGreen("✓ installed")
			if !res.installed[i] {
				status = pterm.LightRed("✗ missing")
				hasMissing = true
			}
			
			toolDisplay := res.name
			if i > 0 {
				toolDisplay = "" 
			}
			
			tableData = append(tableData, []string{
				pterm.Bold.Sprint(toolDisplay),
				pterm.Cyan(v),
				status,
			})
		}
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	if hasMissing {
		fmt.Println()
		pterm.Warning.Printfln("Some versions are specified but not installed. Run '%s' to fix.", pterm.LightMagenta("unirtm install"))
	}

	return nil
}
