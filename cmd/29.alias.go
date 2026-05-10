// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	aliasGlobal bool
)

func init() {
	aliasCmd.PersistentFlags().BoolVar(&aliasGlobal, "global", false, "manage global aliases (~/.config/unirtm/unirtm.toml)")
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasDeleteCmd)
	rootCmd.AddCommand(aliasCmd)
}

var aliasCmd = &cobra.Command{
	Use:     "alias",
	Aliases: []string{"tool-alias"},
	Short:   "Manage version aliases",
	Long: `Manage version aliases for tools.

Aliases allow you to refer to a specific version by a name (e.g. "lts", "work").
They can be managed globally or at the project level using the --global flag.`,
}

var aliasListCmd = &cobra.Command{
	Use:     "list [tool]",
	Aliases: []string{"ls"},
	Short:   "List aliases",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigFilePath(aliasGlobal)
		m, err := loadRawTOML(cfgPath)
		if err != nil {
			return err
		}

		rawAliases, ok := m["aliases"].(map[string]interface{})
		if !ok || len(rawAliases) == 0 {
			pterm.Info.Println("No aliases configured.")
			return nil
		}

		var data [][]string
		data = append(data, []string{"Tool", "Alias", "Version"})

		if len(args) == 1 {
			tool := args[0]
			toolAliases, ok := rawAliases[tool].(map[string]interface{})
			if !ok {
				pterm.Warning.Printf("No aliases found for tool %s\n", tool)
				return nil
			}
			for alias, ver := range toolAliases {
				data = append(data, []string{tool, alias, fmt.Sprintf("%v", ver)})
			}
		} else {
			for toolName, toolAliasesIfc := range rawAliases {
				toolAliases, ok := toolAliasesIfc.(map[string]interface{})
				if !ok {
					continue
				}
				for alias, ver := range toolAliases {
					data = append(data, []string{toolName, alias, fmt.Sprintf("%v", ver)})
				}
			}
		}
		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
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
		return nil
	},
}

var aliasDeleteCmd = &cobra.Command{
	Use:     "delete <tool> <alias>",
	Aliases: []string{"rm", "remove"},
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
			pterm.Warning.Printf("No aliases found.")
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
