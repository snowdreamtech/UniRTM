// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// init registers the config command and its subcommands.
func init() {
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configGenerateCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(configCmd)
	}
}

// configCmd is the parent config command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage UniRTM configuration",
	Long: `Manage UniRTM configuration settings.

This command manages settings that control UniRTM's behavior (e.g. cache TTL, data directory).
It is NOT for managing tools or environment variables (use 'set' or 'alias' for those).

If no subcommand is provided, it displays the current merged configuration.`,
	Aliases: []string{"cfg"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow(cmd, args)
	},
}

// configValidateCmd validates configuration files.
var configValidateCmd = &cobra.Command{
	Use:     "validate [file]",
	Aliases: []string{"check"},
	Short:   "Validate configuration files",
	Long: `Validate UniRTM configuration files for syntax and semantic errors.

If no file is specified, validates the default configuration hierarchy.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigValidate,
}

// configShowCmd displays the merged configuration.
var configShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Display merged configuration",
	Long:    `Display the merged configuration from all sources in the hierarchy.`,
	Aliases: []string{"ls", "cat", "list"},
	Args:    cobra.NoArgs,
	RunE:    runConfigShow,
}

// configGenerateCmd generates a default configuration file.
var configGenerateCmd = &cobra.Command{
	Use:   "generate [file]",
	Short: "Generate a default configuration file",
	Long: `Generate a default unirtm.toml configuration file with common settings
and placeholders for tools.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigGenerate,
}

// configSetCmd sets a configuration value.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in the local configuration file.

Examples:
  unirtm config set tools.node.version 20.0.0
  unirtm config set settings.cache_ttl 48h`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configGetCmd gets a configuration value.
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value from the merged configuration.

Examples:
  unirtm config get tools.node.version
  unirtm config get settings.cache_ttl`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// runConfigValidate validates configuration files.
//
// Validates: Requirements 13.6, 23.2
func runConfigValidate(cmd *cobra.Command, args []string) error {
	isTerminal := term.IsTerminal(int(os.Stdout.Fd())) && !jsonOutput
	if isTerminal {
		pterm.DefaultHeader.WithFullWidth().
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightMagenta)).
			WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
			Println("UniRTM Config Validation")
	}

	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()
	cm := config.NewConfigManager()

	var cfg *config.Config
	var loadErr error

	if len(args) == 1 {
		cfg, loadErr = cm.Load(ctx, args[0])
	} else {
		cfg, loadErr = cm.LoadHierarchy(ctx)
	}

	if loadErr != nil {
		formatter.Error("Failed to load configuration", map[string]interface{}{
			"error": loadErr.Error(),
		})
		return fmt.Errorf("load configuration: %w", loadErr)
	}

	if err := cm.Validate(ctx, cfg); err != nil {
		if jsonOutput {
			formatter.Error("Configuration validation failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			fmt.Fprintf(os.Stderr, "Configuration validation failed:\n  %s\n", err.Error())
		}
		return fmt.Errorf("configuration invalid: %w", err)
	}

	formatter.Success("Configuration is valid", nil)
	return nil
}

// runConfigShow displays the merged configuration.
//
// Validates: Requirements 13.6, 23.2
func runConfigShow(cmd *cobra.Command, args []string) error {
	isTerminal := term.IsTerminal(int(os.Stdout.Fd())) && !jsonOutput
	if isTerminal {
		pterm.DefaultHeader.WithFullWidth().
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightMagenta)).
			WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
			Println("UniRTM Configuration")
	}

	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	ctx := context.Background()
	cm := config.NewConfigManager()

	cfg, err := cm.LoadHierarchy(ctx)
	if err != nil {
		if verbose {
			pterm.Info.Printf("No configuration files found, using defaults: %v\n", err)
		}
		cfg = &config.Config{}
	}

	if jsonOutput {
		formatter.Success("Configuration", map[string]interface{}{
			"config": cfg,
		})
		return nil
	}

	// 1. Tool Section
	if !isTerminal {
		pterm.DefaultSection.Println("Tools Configuration")
	} else {
		pterm.DefaultSection.Println("Tools")
	}
	if len(cfg.Tools) == 0 {
		pterm.Info.Println("  (no tools defined)")
	} else {
		var toolItems []pterm.BulletListItem
		for name, tool := range cfg.Tools {
			desc := tool.Version
			if tool.Backend != "" {
				desc = fmt.Sprintf("%s (backend: %s)", desc, tool.Backend)
			}
			toolItems = append(toolItems, pterm.BulletListItem{Level: 0, Text: fmt.Sprintf("%s: %s", name, desc)})
		}
		pterm.DefaultBulletList.WithItems(toolItems).Render()
	}

	// 2. Settings Section
	if !isTerminal {
		pterm.DefaultSection.Println("UniRTM Settings")
	}
	pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Cache TTL:  %v", cfg.Settings.CacheTTL)},
		{Level: 0, Text: fmt.Sprintf("Data Dir:   %s", env.GetDataDir())},
		{Level: 0, Text: fmt.Sprintf("Config Dir: %s", env.GetConfigDir())},
		{Level: 0, Text: fmt.Sprintf("Cache Dir:  %s", env.GetCacheDir())},
	}).Render()

	// 3. Environment & Mirrors Section
	if len(cfg.Env) > 0 {
		var mirrorItems []pterm.BulletListItem
		var envItems []pterm.BulletListItem

		// Sort items into groups
		for k, v := range cfg.Env {
			valStr := fmt.Sprintf("%v", v)
			if valStr == "" {
				valStr = pterm.Gray("(unset)")
			}

			item := pterm.BulletListItem{Level: 0, Text: fmt.Sprintf("%s = %s", k, valStr)}

			// Detect mirrors/proxies
			kUpper := strings.ToUpper(k)
			if strings.Contains(kUpper, "MIRROR") || strings.Contains(kUpper, "PROXY") ||
				strings.Contains(kUpper, "REGISTRY") || strings.Contains(kUpper, "SERVER") ||
				strings.Contains(kUpper, "GOPROXY") || strings.Contains(kUpper, "INDEX_URL") {
				mirrorItems = append(mirrorItems, item)
			} else {
				envItems = append(envItems, item)
			}
		}

		if len(mirrorItems) > 0 {
			pterm.DefaultSection.Println("Mirrors & Proxies")
			pterm.DefaultBulletList.WithItems(mirrorItems).Render()
		}

		if len(envItems) > 0 {
			pterm.DefaultSection.Println("Environment Variables")
			pterm.DefaultBulletList.WithItems(envItems).Render()
		}
	}

	return nil
}

// runConfigGet retrieves a configuration value.
//
// Validates: Requirements 13.6, 23.2
func runConfigGet(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	key := args[0]

	v := viper.New()
	loadViperConfigHierarchy(v)

	value := v.Get(key)
	if value == nil {
		formatter.Error(fmt.Sprintf("Key %q not found in configuration", key), nil)
		return fmt.Errorf("key %q not found", key)
	}

	if jsonOutput {
		formatter.Success("Configuration value", map[string]interface{}{
			"key":   key,
			"value": value,
		})
		return nil
	}

	fmt.Printf("%v\n", value)
	return nil
}

// runConfigSet sets a configuration value in the local config file.
//
// Validates: Requirements 13.6, 23.2
func runConfigSet(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	key := args[0]
	value := args[1]

	localConfigFile := ".unirtm.toml"
	if configPath != "" {
		localConfigFile = configPath
	} else {
		// If no specific path is provided, find the existing config file, prioritizing TOML
		for _, f := range []string{".unirtm.toml", ".unirtm.yaml", ".unirtm.yml", "unirtm.toml", "unirtm.yaml", "unirtm.yml"} {
			if _, err := os.Stat(f); err == nil {
				localConfigFile = f
				break
			}
		}
	}

	v := viper.New()
	v.SetConfigFile(localConfigFile)
	ext := filepath.Ext(localConfigFile)
	configType := strings.TrimPrefix(ext, ".")
	if configType == "" {
		configType = "toml"
	}
	v.SetConfigType(configType)
	_ = v.ReadInConfig()

	v.Set(key, value)

	if err := v.WriteConfigAs(localConfigFile); err != nil {
		formatter.Error("Failed to write configuration", map[string]interface{}{
			"file":  localConfigFile,
			"error": err.Error(),
		})
		return fmt.Errorf("write config: %w", err)
	}

	formatter.Success(fmt.Sprintf("Set %s = %s in %s", key, value, localConfigFile), nil)
	return nil
}

// runConfigGenerate generates a default configuration file.
func runConfigGenerate(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	targetFile := ".unirtm.toml"
	if len(args) == 1 {
		targetFile = args[0]
	}

	if _, err := os.Stat(targetFile); err == nil {
		if !updateForce {
			formatter.Error(fmt.Sprintf("File %s already exists. Use --force to overwrite.", targetFile), nil)
			return fmt.Errorf("file exists: %s", targetFile)
		}
	}

	defaultConfig := `# UniRTM configuration file
# For more information, see https://github.com/snowdreamtech/unirtm

[settings]
# Time to live for tool version index cache
cache_ttl = "168h"

[tools]
# Define tools and their versions here
# node = "20.0.0"
# python = "3.12.0"
# cli/cli = "latest"

[env]
# Define environment variables here
# NODE_ENV = "development"

[tasks]
# Define project tasks here
# build = "go build -o bin/app ."
# test = "go test ./..."
`

	if err := os.WriteFile(targetFile, []byte(defaultConfig), 0644); err != nil {
		formatter.Error(fmt.Sprintf("Failed to write configuration to %s", targetFile), map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	formatter.Success(fmt.Sprintf("Generated default configuration in %s", targetFile), nil)
	return nil
}

// loadViperConfigHierarchy loads configuration hierarchy into a viper instance.
func loadViperConfigHierarchy(v *viper.Viper) {
	// Order matters for MergeInConfig: later files override earlier files.
	// By putting TOML last, it takes precedence over YAML if both exist in the same directory.
	candidates := []string{
		"unirtm.yml", "unirtm.yaml", "unirtm.toml",
		".unirtm.yml", ".unirtm.yaml", ".unirtm.toml",
	}
	if configPath != "" {
		// Command line flag has highest priority
		candidates = append(candidates, configPath)
	}
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			v.SetConfigFile(f)
			ext := filepath.Ext(f)
			configType := strings.TrimPrefix(ext, ".")
			if configType == "" {
				configType = "toml"
			}
			v.SetConfigType(configType)
			_ = v.MergeInConfig()
		}
	}
}
