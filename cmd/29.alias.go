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

var (
	aliasGlobal bool
)

func init() {
	aliasCmd.PersistentFlags().BoolVar(&aliasGlobal, "global", false, "manage global aliases (~/.config/unirtm/unirtm.toml)")
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasUnsetCmd)
	aliasCmd.AddCommand(aliasResolveCmd)
	rootCmd.AddCommand(aliasCmd)
}

var aliasResolveCmd = &cobra.Command{
	Use:   "resolve <tool> <alias>",
	Short: "Resolve an alias to its actual version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, alias := args[0], args[1]
		cfg, err := config.LoadFull()
		if err != nil {
			return err
		}

		if toolAliases, ok := cfg.Aliases[tool]; ok {
			if version, ok := toolAliases[alias]; ok {
				fmt.Println(version)
				return nil
			}
		}

		pterm.Error.Printf("Alias %s not found for tool %s\n", alias, tool)
		return fmt.Errorf("alias not found")
	},
}

var aliasCmd = &cobra.Command{
	Use:     "alias",
	Aliases: []string{"tool-alias"},
	Short:   "Manage version aliases",
	Long: `Manage version aliases for tools.

Aliases allow you to refer to a specific version by a name (e.g. "lts", "work").
They can be managed globally or at the project level using the --global flag.

If no subcommand is provided, it lists all aliases.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return aliasListCmd.RunE(cmd, args)
	},
}

var aliasListCmd = &cobra.Command{
	Use:     "ls [tool]",
	Aliases: []string{"list"},
	Short:   "List aliases",
	Long:    `List aliases from all configuration levels (Global, Parent, Local).`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFull()
		if err != nil {
			return err
		}

		if len(cfg.Aliases) == 0 {
			pterm.Info.Println("No aliases configured.")
			return nil
		}

		filterTool := ""
		if len(args) == 1 {
			filterTool = args[0]
		}

		pterm.DefaultSection.Println("Tool Aliases")

		// Sort tools for consistent output
		var tools []string
		for t := range cfg.Aliases {
			if filterTool != "" && t != filterTool {
				continue
			}
			tools = append(tools, t)
		}
		sort.Strings(tools)

		for _, tool := range tools {
			aliases := cfg.Aliases[tool]
			pterm.Bold.Printf("• %s\n", tool)
			
			var items []pterm.BulletListItem
			var names []string
			for n := range aliases {
				names = append(names, n)
			}
			sort.Strings(names)

			for _, name := range names {
				version := aliases[name]
				// We can detect source by looking at global config specifically
				sourceTag := pterm.LightBlue("[Local]")
				if globalCfg, err := config.LoadGlobal(); err == nil {
					if gAliases, ok := globalCfg.Aliases[tool]; ok {
						if gVersion, ok := gAliases[name]; ok && gVersion == version {
							sourceTag = pterm.LightMagenta("[Global]")
						}
					}
				}

				items = append(items, pterm.BulletListItem{
					Level: 1,
					Text: fmt.Sprintf("%s %s %s %s", 
						pterm.Cyan(name), 
						pterm.Gray("→"), 
						pterm.Green(version), 
						sourceTag),
				})
			}
			pterm.DefaultBulletList.WithItems(items).Render()
			fmt.Println() // Spacer
		}
		return nil
	},
}

var aliasSetCmd = &cobra.Command{
	Use:     "set <tool> <alias> <version>",
	Aliases: []string{"add"},
	Short:   "Set an alias",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, alias, version := args[0], args[1], args[2]

		// 1. Resolve tool name and backend
		toolName := tool
		backendName := ""
		if idx := strings.Index(tool, ":"); idx != -1 {
			backendName = tool[:idx]
			toolName = tool[idx+1:]
		}

		// 2. Load config
		cfgPath := resolveConfigFilePath(aliasGlobal)
		m, err := loadRawTOML(cfgPath)
		if err != nil {
			return err
		}

		rawAliases, ok := m["aliases"].(map[string]interface{})
		if !ok {
			rawAliases = make(map[string]interface{})
			m["aliases"] = rawAliases
		}

		toolAliases, ok := rawAliases[tool].(map[string]interface{})
		if !ok {
			toolAliases = make(map[string]interface{})
			rawAliases[tool] = toolAliases
		}

		toolAliases[alias] = version

		if err := saveRawTOML(cfgPath, m); err != nil {
			return fmt.Errorf("failed to save alias: %w", err)
		}

		pterm.Success.Printf("Set alias %s=%s for tool %s in %s\n", alias, version, tool, cfgPath)

		// [Surpass] Validation: Check if version exists (optional warning)
		fsToolName := env.GetFSToolName(toolName, backendName)
		installPath := filepath.Join(env.GetInstallsDir(), fsToolName, version)
		if _, err := os.Stat(installPath); os.IsNotExist(err) {
			pterm.Warning.Printf("Version %s for %s is not currently installed. You may need to run 'unirtm install' later.\n", version, tool)
		}

		return nil
	},
}

var aliasUnsetCmd = &cobra.Command{
	Use:     "unset <tool> <alias>",
	Aliases: []string{"rm", "remove", "delete"},
	Short:   "Delete an alias",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, alias := args[0], args[1]

		cfgPath := resolveConfigFilePath(aliasGlobal)
		m, err := loadRawTOML(cfgPath)
		if err != nil {
			return err
		}

		rawAliases, ok := m["aliases"].(map[string]interface{})
		if !ok {
			pterm.Warning.Println("No aliases found.")
			return nil
		}

		toolAliases, ok := rawAliases[tool].(map[string]interface{})
		if !ok {
			pterm.Warning.Printf("No aliases found for tool %s\n", tool)
			return nil
		}

		if _, ok := toolAliases[alias]; !ok {
			pterm.Warning.Printf("Alias %s not found for tool %s\n", alias, tool)
			return nil
		}

		delete(toolAliases, alias)
		if len(toolAliases) == 0 {
			delete(rawAliases, tool)
		}
		if len(rawAliases) == 0 {
			delete(m, "aliases")
		}

		if err := saveRawTOML(cfgPath, m); err != nil {
			return fmt.Errorf("failed to save alias: %w", err)
		}

		pterm.Success.Printf("Deleted alias %s for tool %s in %s\n", alias, tool, cfgPath)
		return nil
	},
}
