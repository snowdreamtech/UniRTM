// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	shellAliasCmd.AddCommand(shellAliasListCmd)
	shellAliasCmd.AddCommand(shellAliasAddCmd)
	shellAliasCmd.AddCommand(shellAliasRemoveCmd)
	if rootCmd != nil {
		rootCmd.AddCommand(shellAliasCmd)
	}
}

// shellAliasCmd manages tool version aliases in the config file.
var shellAliasCmd = &cobra.Command{
	Use:     "shell-alias",
	Short:   "Manage tool version aliases",
	Aliases: []string{"alias"},
	Long: `Manage tool version aliases in the unirtm.toml file.

Aliases allow you to refer to a specific version by a name (e.g. "lts", "work").
They are defined in the [aliases] section of the config.

Sub-commands:
  list             List all configured aliases
  add <tool> <alias> <version>  Add a new alias
  remove <tool> <alias>         Remove an alias

Examples:
  # List all aliases
  unirtm alias list

  # Add an alias for node
  unirtm alias add node lts 22.14.0

  # Remove an alias
  unirtm alias remove node lts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return shellAliasListCmd.RunE(shellAliasListCmd, args)
	},
}

var shellAliasListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all configured tool version aliases",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    runShellAliasList,
}

var shellAliasAddCmd = &cobra.Command{
	Use:   "add <tool> <alias> <version>",
	Short: "Add a new tool version alias",
	Args:  cobra.ExactArgs(3),
	RunE:  runShellAliasAdd,
}

var shellAliasRemoveCmd = &cobra.Command{
	Use:     "remove <tool> <alias>",
	Short:   "Remove a tool version alias",
	Aliases: []string{"rm", "delete"},
	Args:    cobra.ExactArgs(2),
	RunE:    runShellAliasRemove,
}

func runShellAliasList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	cfgPath := resolveConfigFilePath(false)
	cfgMap, err := loadRawTOML(cfgPath)
	if err != nil {
		formatter.Info("No aliases found (config file missing).", nil)
		return nil
	}

	aliases, ok := cfgMap["aliases"].(map[string]interface{})
	if !ok || len(aliases) == 0 {
		formatter.Info("No aliases configured.", nil)
		return nil
	}

	if jsonOutput {
		formatter.Success("Aliases", map[string]interface{}{"aliases": aliases})
		return nil
	}

	tableData := pterm.TableData{
		{"TOOL", "ALIAS", "VERSION"},
	}

	// Sort tool names for consistent output
	tools := make([]string, 0, len(aliases))
	for t := range aliases {
		tools = append(tools, t)
	}
	sort.Strings(tools)

	for _, t := range tools {
		toolAliases, ok := aliases[t].(map[string]interface{})
		if !ok {
			continue
		}
		
		// Sort aliases within tool
		aliasNames := make([]string, 0, len(toolAliases))
		for a := range toolAliases {
			aliasNames = append(aliasNames, a)
		}
		sort.Strings(aliasNames)

		for _, a := range aliasNames {
			tableData = append(tableData, []string{
				pterm.FgCyan.Sprint(t),
				pterm.FgYellow.Sprint(a),
				fmt.Sprint(toolAliases[a]),
			})
		}
	}

	fmt.Println()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	return nil
}

func runShellAliasAdd(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	tool := args[0]
	alias := args[1]
	version := args[2]

	cfgPath := resolveConfigFilePath(false)
	cfgMap, _ := loadRawTOML(cfgPath)

	aliases, ok := cfgMap["aliases"].(map[string]interface{})
	if !ok {
		aliases = make(map[string]interface{})
		cfgMap["aliases"] = aliases
	}

	toolAliases, ok := aliases[tool].(map[string]interface{})
	if !ok {
		toolAliases = make(map[string]interface{})
		aliases[tool] = toolAliases
	}

	toolAliases[alias] = version

	if err := saveRawTOML(cfgPath, cfgMap); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}

	formatter.Success(fmt.Sprintf("Added alias %s=%s for tool %s", alias, version, tool), nil)
	return nil
}

func runShellAliasRemove(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	tool := args[0]
	alias := args[1]

	cfgPath := resolveConfigFilePath(false)
	cfgMap, err := loadRawTOML(cfgPath)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load config: %v", err))
		return err
	}

	aliases, ok := cfgMap["aliases"].(map[string]interface{})
	if !ok {
		formatter.Warning(fmt.Sprintf("No aliases found for tool %s", tool))
		return nil
	}

	toolAliases, ok := aliases[tool].(map[string]interface{})
	if !ok {
		formatter.Warning(fmt.Sprintf("No aliases found for tool %s", tool))
		return nil
	}

	if _, ok := toolAliases[alias]; !ok {
		formatter.Warning(fmt.Sprintf("Alias %s not found for tool %s", alias, tool))
		return nil
	}

	delete(toolAliases, alias)
	if len(toolAliases) == 0 {
		delete(aliases, tool)
	}
	if len(aliases) == 0 {
		delete(cfgMap, "aliases")
	}

	if err := saveRawTOML(cfgPath, cfgMap); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}

	formatter.Success(fmt.Sprintf("Removed alias %s for tool %s", alias, tool), nil)
	return nil
}
