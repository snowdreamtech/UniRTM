// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// init registers the config command and its subcommands.
func init() {
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)

	if rootCmd != nil {
		rootCmd.AddCommand(configCmd)
	}
}

// configCmd is the parent config command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage UniRTM configuration",
	Long: `Manage UniRTM configuration files.

Subcommands:
  validate  Validate configuration files
  show      Display merged configuration
  set       Set a configuration value
  get       Get a configuration value

Examples:
  unirtm config validate
  unirtm config show
  unirtm config get tools.node.version
  unirtm config set tools.node.version 20.0.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// configValidateCmd validates configuration files.
var configValidateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate configuration files",
	Long: `Validate UniRTM configuration files for syntax and semantic errors.

If no file is specified, validates the default configuration hierarchy.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigValidate,
}

// configShowCmd displays the merged configuration.
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display merged configuration",
	Long:  `Display the merged configuration from all sources in the hierarchy.`,
	Args:  cobra.NoArgs,
	RunE:  runConfigShow,
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
			formatter.Info("No configuration files found, showing defaults", map[string]interface{}{
				"error": err.Error(),
			})
		}
		cfg = &config.Config{}
	}

	if jsonOutput {
		formatter.Success("Configuration", map[string]interface{}{
			"config": cfg,
		})
		return nil
	}

	// Human-readable display
	fmt.Println("Merged Configuration:")
	fmt.Println(strings.Repeat("-", 40))

	if len(cfg.Tools) == 0 {
		fmt.Println("  tools: (none)")
	} else {
		fmt.Println("  tools:")
		for name, tool := range cfg.Tools {
			fmt.Printf("    %s:\n", name)
			if tool.Version != "" {
				fmt.Printf("      version: %s\n", tool.Version)
			}
			if tool.Backend != "" {
				fmt.Printf("      backend: %s\n", tool.Backend)
			}
		}
	}

	fmt.Println("  settings:")
	fmt.Printf("    cache_ttl:   %d\n", cfg.Settings.CacheTTL)
	fmt.Printf("    data_dir:    %s\n", cfg.Settings.DataDir)
	fmt.Printf("    cache_dir:   %s\n", cfg.Settings.CacheDir)

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
	}

	v := viper.New()
	v.SetConfigFile(localConfigFile)
	v.SetConfigType("toml")
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

// loadViperConfigHierarchy loads configuration hierarchy into a viper instance.
func loadViperConfigHierarchy(v *viper.Viper) {
	candidates := []string{".unirtm.toml", "unirtm.toml"}
	if configPath != "" {
		candidates = append([]string{configPath}, candidates...)
	}
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			v.SetConfigFile(f)
			v.SetConfigType("toml")
			_ = v.MergeInConfig()
		}
	}
}
