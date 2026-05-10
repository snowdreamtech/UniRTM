// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var aliasCmd = &cobra.Command{
	Use:     "alias",
	Aliases: []string{"tool-alias"},
	Short:   "Manage version aliases",
	Long:    `Manage global version aliases for tools.`,
}

var aliasListCmd = &cobra.Command{
	Use:   "list [tool]",
	Short: "List aliases",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigPath := filepath.Join(env.GetConfigDir(), "config.toml")
		v := viper.New()
		v.SetConfigFile(globalConfigPath)
		v.SetConfigType("toml")
		_ = v.ReadInConfig()

		if len(args) == 1 {
			tool := args[0]
			toolAliases := v.GetStringMapString(fmt.Sprintf("aliases.%s", tool))
			if len(toolAliases) == 0 {
				pterm.Warning.Printf("No aliases found for tool %s\n", tool)
				return nil
			}
			pterm.DefaultTable.WithHasHeader().WithData(toTableData(tool, toolAliases)).Render()
			return nil
		}

		// List all
		allAliasesMap := v.GetStringMap("aliases")
		if len(allAliasesMap) == 0 {
			pterm.Info.Println("No aliases configured.")
			return nil
		}

		var data [][]string
		data = append(data, []string{"Tool", "Alias", "Version"})
		for toolName, toolAliasesIfc := range allAliasesMap {
			toolAliases, ok := toolAliasesIfc.(map[string]interface{})
			if !ok {
				continue
			}
			for alias, verIfc := range toolAliases {
				data = append(data, []string{toolName, alias, fmt.Sprintf("%v", verIfc)})
			}
		}
		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
		return nil
	},
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <tool> <alias> <version>",
	Short: "Set an alias",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, alias, version := args[0], args[1], args[2]

		globalConfigPath := filepath.Join(env.GetConfigDir(), "config.toml")
		err := os.MkdirAll(filepath.Dir(globalConfigPath), 0755)
		if err != nil {
			return err
		}

		v := viper.New()
		v.SetConfigFile(globalConfigPath)
		v.SetConfigType("toml")
		_ = v.ReadInConfig() // ignore error if file doesn't exist

		key := fmt.Sprintf("aliases.%s.%s", tool, alias)
		v.Set(key, version)

		if err := v.WriteConfigAs(globalConfigPath); err != nil {
			return fmt.Errorf("failed to save alias: %w", err)
		}

		pterm.Success.Printf("Set alias %s=%s for tool %s\n", alias, version, tool)
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

		globalConfigPath := filepath.Join(env.GetConfigDir(), "config.toml")
		v := viper.New()
		v.SetConfigFile(globalConfigPath)
		v.SetConfigType("toml")
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("no global config found: %w", err)
		}

		key := fmt.Sprintf("aliases.%s.%s", tool, alias)
		if !v.IsSet(key) {
			pterm.Warning.Printf("Alias %s not found for tool %s\n", alias, tool)
			return nil
		}

		// Viper doesn't have an easy way to delete a specific nested key in the config without a workaround or re-marshaling manually, 
		// but since viper.Get handles it, we can get the tool aliases map, delete the key, and write it back.
		toolAliases := v.GetStringMapString(fmt.Sprintf("aliases.%s", tool))
		delete(toolAliases, alias)
		
		if len(toolAliases) == 0 {
			// If empty, we can just omit it or set it to empty map, but Viper set won't delete the node cleanly.
			// Reassigning works.
		}
		
		v.Set(fmt.Sprintf("aliases.%s", tool), toolAliases)

		if err := v.WriteConfigAs(globalConfigPath); err != nil {
			return fmt.Errorf("failed to save alias: %w", err)
		}

		pterm.Success.Printf("Deleted alias %s for tool %s\n", alias, tool)
		return nil
	},
}

func toTableData(tool string, aliases map[string]string) [][]string {
	var data [][]string
	data = append(data, []string{"Tool", "Alias", "Version"})
	for k, v := range aliases {
		data = append(data, []string{tool, k, v})
	}
	return data
}

func init() {
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasDeleteCmd)
	rootCmd.AddCommand(aliasCmd)
}
