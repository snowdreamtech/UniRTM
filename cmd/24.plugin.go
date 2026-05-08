// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

// init registers the plugin command and its subcommands to the root command.
func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(pluginCmd)
	}
}

// pluginCmd is the parent command for plugin management.
// Equivalent to `mise plugin`.
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage UniRTM plugins",
	Long: `Manage UniRTM backend and provider plugins.

UniRTM supports Go-native plugins that extend its functionality with
custom backends (tool sources) and providers (tool-specific install logic).
Plugins are standalone executable binaries prefixed with 'unirtm-plugin-'
placed in the plugins directory (~/.local/share/unirtm/plugins/).

Use 'unirtm plugin list' to see loaded plugins.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// pluginListCmd lists all currently loaded plugins.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List loaded plugins",
	Long: `List all plugins that are currently loaded by UniRTM.

Examples:
  unirtm plugin list
  unirtm plugin list --json`,
	Args: cobra.NoArgs,
	RunE: runPluginList,
}

// runPluginList lists all loaded plugins from the plugin manager.
func runPluginList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()
	pluginsDir := getDefaultDataDir() + "/plugins"

	// Create plugin manager (pass nil registries — we only want to list, not load into active registries)
	pm := service.NewPluginManager(pluginsDir, nil, nil)
	defer pm.Cleanup()
	if err := pm.LoadAll(ctx); err != nil {
		formatter.Warning(fmt.Sprintf("Some plugins failed to load: %v", err), nil)
	}

	plugins := pm.ListLoaded()
	if len(plugins) == 0 {
		formatter.Info(fmt.Sprintf("No plugins found in %s", pluginsDir), nil)
		formatter.Info("See docs/development/plugin-development.md to create plugins", nil)
		return nil
	}

	if jsonOutput {
		formatter.Success("Plugins", map[string]interface{}{
			"count":   len(plugins),
			"plugins": plugins,
		})
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tAPI VERSION\tPATH")
	fmt.Fprintln(w, "----\t----\t-----------\t----")
	for _, p := range plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Type, p.APIVersion, p.Path)
	}
	w.Flush()
	return nil
}

// pluginInstallCmd registers a plugin from a local executable file path.
var pluginInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a plugin from a local executable file",
	Long: `Install a plugin from a compiled Go plugin file.

The plugin file must be a standalone executable (e.g. built with 'go build')
and ideally prefixed with 'unirtm-plugin-'. See docs/development/plugin-development.md
for the plugin API.

Examples:
  unirtm plugin install ./unirtm-plugin-myplugin`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInstall,
}

// runPluginInstall copies a plugin .so file to the plugins directory.
func runPluginInstall(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	srcPath := args[0]
	pluginsDir := getDefaultDataDir() + "/plugins"

	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would install plugin from %s to %s", srcPath, pluginsDir), nil)
		return nil
	}

	// Ensure plugins directory exists
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("create plugins directory: %w", err)
	}

	// Read source
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read plugin file %s: %w", srcPath, err)
	}

	// Determine destination filename
	destName := srcPath
	for i := len(srcPath) - 1; i >= 0; i-- {
		if srcPath[i] == '/' || srcPath[i] == '\\' {
			destName = srcPath[i+1:]
			break
		}
	}
	destPath := pluginsDir + "/" + destName

	// Write destination
	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("write plugin to %s: %w", destPath, err)
	}

	formatter.Success(fmt.Sprintf("Plugin installed to %s", destPath), map[string]interface{}{
		"src":  srcPath,
		"dest": destPath,
	})
	return nil
}

// pluginRemoveCmd removes an installed plugin.
var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed plugin",
	Long: `Remove an installed plugin from the plugins directory.

Examples:
  unirtm plugin remove myplugin`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginRemove,
}

// runPluginRemove removes a plugin file from the plugins directory.
func runPluginRemove(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	name := args[0]
	pluginsDir := getDefaultDataDir() + "/plugins"

	// Try common extensions and prefixes
	candidates := []string{
		pluginsDir + "/unirtm-plugin-" + name,
		pluginsDir + "/unirtm-plugin-" + name + ".exe",
		pluginsDir + "/" + name,
		pluginsDir + "/" + name + ".exe",
	}

	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would remove plugin '%s' from %s", name, pluginsDir), nil)
		return nil
	}

	for _, p := range candidates {
		if err := os.Remove(p); err == nil {
			formatter.Success(fmt.Sprintf("Plugin '%s' removed (%s)", name, p), nil)
			return nil
		}
	}

	return fmt.Errorf("plugin '%s' not found in %s", name, pluginsDir)
}
