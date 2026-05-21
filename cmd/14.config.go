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
	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
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
		pterm.DefaultSection.Println("Config Validation")
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
		pterm.DefaultSection.Println("Configuration")
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

	// 0. Active Configuration Files Section
	if !isTerminal {
		pterm.DefaultSection.Println("Active Configuration Files")
	} else {
		pterm.DefaultSection.Println("Config Files")
	}

	var activeFiles []string
	systemPaths := []string{
		"/etc/unirtm/config.toml",
		"/etc/unirtm/config.yaml",
		"/etc/unirtm/config.yml",
	}
	for _, p := range systemPaths {
		if _, err := os.Stat(p); err == nil {
			activeFiles = append(activeFiles, p)
		}
	}

	globalPaths := []string{
		filepath.Join(env.GetConfigDir(), "config.toml"),
		filepath.Join(env.GetConfigDir(), "config.yaml"),
		filepath.Join(env.GetConfigDir(), "config.yml"),
	}
	for _, p := range globalPaths {
		if _, err := os.Stat(p); err == nil {
			activeFiles = append(activeFiles, p)
		}
	}

	cwd, _ := os.Getwd()
	curr := cwd
	var projectFiles []string

	isCeiling := func(path string) bool {
		absPath, _ := filepath.Abs(path)
		for _, cp := range cfg.Settings.CeilingPaths {
			absCP, _ := filepath.Abs(cp)
			if absPath == absCP {
				return true
			}
		}
		parent := filepath.Dir(absPath)
		return parent == absPath
	}

	for {
		files := []string{
			filepath.Join(curr, ".mise.yml"),
			filepath.Join(curr, ".mise.yaml"),
			filepath.Join(curr, ".mise.toml"),
			filepath.Join(curr, "unirtm.yml"),
			filepath.Join(curr, "unirtm.yaml"),
			filepath.Join(curr, "unirtm.toml"),
			filepath.Join(curr, ".unirtm.yml"),
			filepath.Join(curr, ".unirtm.yaml"),
			filepath.Join(curr, ".unirtm.toml"),
			filepath.Join(curr, ".mise.local.yml"),
			filepath.Join(curr, ".mise.local.yaml"),
			filepath.Join(curr, ".mise.local.toml"),
			filepath.Join(curr, ".unirtm.local.yml"),
			filepath.Join(curr, ".unirtm.local.yaml"),
			filepath.Join(curr, ".unirtm.local.toml"),
		}

		var dirFiles []string
		for _, p := range files {
			if _, err := os.Stat(p); err == nil {
				dirFiles = append(dirFiles, p)
			}
		}
		projectFiles = append(dirFiles, projectFiles...)

		if isCeiling(curr) {
			break
		}
		curr = filepath.Dir(curr)
	}
	activeFiles = append(activeFiles, projectFiles...)

	if len(activeFiles) == 0 {
		pterm.Info.Println("  (no configuration files found)")
	} else {
		var fileItems []pterm.BulletListItem
		for _, f := range activeFiles {
			fileItems = append(fileItems, pterm.BulletListItem{Level: 0, Text: pterm.FgCyan.Sprint(f)})
		}
		pterm.DefaultBulletList.WithItems(fileItems).Render()
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

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cm := config.NewConfigManager()
	cfg, err := cm.LoadHierarchy(ctx)
	if err != nil || cfg == nil {
		cfg = &config.Config{}
	}

	// Resolve dotted key into the config struct using a simple top-level section lookup
	parts := strings.SplitN(key, ".", 2)
	var value interface{}
	switch parts[0] {
	case "tools":
		if len(parts) > 1 {
			sub := strings.SplitN(parts[1], ".", 2)
			if t, ok := cfg.Tools[sub[0]]; ok {
				value = t
			}
		} else {
			value = cfg.Tools
		}
	case "settings":
		value = cfg.Settings
	case "env":
		if len(parts) > 1 {
			value = cfg.Env[parts[1]]
		} else {
			value = cfg.Env
		}
	case "tasks":
		value = cfg.Tasks
	default:
		value = nil
	}

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
		for _, f := range []string{".unirtm.toml", ".unirtm.yaml", ".unirtm.yml", "unirtm.toml", "unirtm.yaml", "unirtm.yml"} {
			if _, err := os.Stat(f); err == nil {
				localConfigFile = f
				break
			}
		}
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(localConfigFile), "."))
	if ext == "" {
		ext = "toml"
	}

	// Read existing content as a generic map to preserve unknown keys
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(localConfigFile); err == nil {
		switch ext {
		case "toml":
			_ = toml.Unmarshal(data, &existing)
		case "yaml", "yml":
			_ = yaml.Unmarshal(data, &existing)
		}
	}

	// Set the value at the dotted key path
	setNestedKey(existing, strings.Split(key, "."), value)

	// Serialize back to file
	var out []byte
	var writeErr error
	switch ext {
	case "toml":
		out, writeErr = toml.Marshal(existing)
	case "yaml", "yml":
		out, writeErr = yaml.Marshal(existing)
	default:
		writeErr = fmt.Errorf("unsupported config format: %s", ext)
	}
	if writeErr != nil {
		formatter.Error("Failed to serialize configuration", map[string]interface{}{
			"error": writeErr.Error(),
		})
		return fmt.Errorf("serialize config: %w", writeErr)
	}

	if err := os.WriteFile(localConfigFile, out, 0644); err != nil {
		formatter.Error("Failed to write configuration", map[string]interface{}{
			"file":  localConfigFile,
			"error": err.Error(),
		})
		return fmt.Errorf("write config: %w", err)
	}

	formatter.Success(fmt.Sprintf("Set %s = %s in %s", key, value, localConfigFile), nil)
	return nil
}

// setNestedKey sets a value in a nested map at the given key path.
func setNestedKey(m map[string]interface{}, keys []string, value interface{}) {
	if len(keys) == 1 {
		m[keys[0]] = value
		return
	}
	sub, ok := m[keys[0]].(map[string]interface{})
	if !ok {
		sub = make(map[string]interface{})
	}
	setNestedKey(sub, keys[1:], value)
	m[keys[0]] = sub
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
