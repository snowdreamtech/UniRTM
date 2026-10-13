// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

var configGlobal bool

// init registers the config command and its subcommands.
func init() {
	configCmd.PersistentFlags().BoolVarP(&configGlobal, "global", "g", false, "operate on global user configuration")

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
	Use:   "config [key] [value]",
	Short: "Manage UniRTM configuration",
	Long: `Manage UniRTM configuration settings.

This command manages settings that control UniRTM's behavior (e.g. cache TTL, data directory).
It is NOT for managing tools or environment variables (use 'set' or 'alias' for those).

Examples:
  # View merged configuration
  unirtm config

  # View or edit global machine configuration
  unirtm config --global
  unirtm config --global settings.browser chrome
  unirtm config get --global settings.browser
  unirtm config set --global settings.browser chrome`,
	Aliases: []string{"cfg"},
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 {
			return runConfigSet(cmd, args)
		}
		if len(args) == 1 {
			return runConfigGet(cmd, args)
		}
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
  unirtm config set settings.cache_ttl 48h
  unirtm config set plugins."snowdreamtech/docker".config.port 8080
  unirtm config set plugins."community/deploy".config.cluster staging`,
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
  unirtm config get settings.cache_ttl
  unirtm config get plugins."snowdreamtech/docker".config.port`,
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

	var cfg *config.Config
	var err error
	if configGlobal {
		cfg, err = config.LoadGlobal()
	} else if configPath != "" {
		cfg, err = cm.Load(ctx, configPath)
	} else {
		cfg, err = cm.LoadHierarchy(ctx)
	}
	if err != nil {
		if verbose {
			output.Infof("No configuration files found, using defaults: %v", err)
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
		if configGlobal {
			pterm.DefaultSection.Println("Global Configuration")
		} else {
			pterm.DefaultSection.Println("Active Configuration Files")
		}
	} else {
		if configGlobal {
			pterm.DefaultSection.Println("Global Config")
		} else {
			pterm.DefaultSection.Println("Config Files")
		}
	}

	var activeFiles []string
	if configGlobal {
		p := config.GetGlobalConfigPath()
		if _, err := os.Stat(p); err == nil {
			activeFiles = append(activeFiles, p)
		}
	} else if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			activeFiles = append(activeFiles, configPath)
		}
	} else {
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
}

	if len(activeFiles) == 0 {
		output.Info("  (no configuration files found)")
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
		output.Info("  (no tools defined)")
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
	settingsList := []pterm.BulletListItem{
		{Level: 0, Text: fmt.Sprintf("Cache TTL:  %v", cfg.Settings.CacheTTL)},
		{Level: 0, Text: fmt.Sprintf("Data Dir:   %s", env.GetDataDir())},
		{Level: 0, Text: fmt.Sprintf("Config Dir: %s", env.GetConfigDir())},
		{Level: 0, Text: fmt.Sprintf("Cache Dir:  %s", env.GetCacheDir())},
	}
	if cfg.Settings.Browser != "" {
		settingsList = append(settingsList, pterm.BulletListItem{Level: 0, Text: fmt.Sprintf("Browser:    %s", cfg.Settings.Browser)})
	}
	pterm.DefaultBulletList.WithItems(settingsList).Render()

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

	// 4. Plugins & Extensions Section
	if len(cfg.Plugins) > 0 {
		if !isTerminal {
			pterm.DefaultSection.Println("Plugins & Extensions")
		} else {
			pterm.DefaultSection.Println("Plugins")
		}
		var pluginItems []pterm.BulletListItem
		for name, p := range cfg.Plugins {
			desc := p.Version
			if desc == "" {
				desc = "enabled"
			}
			if len(p.Config) > 0 {
				desc = fmt.Sprintf("%s (%d config items)", desc, len(p.Config))
			}
			pluginItems = append(pluginItems, pterm.BulletListItem{Level: 0, Text: fmt.Sprintf("%s: %s", name, desc)})
		}
		pterm.DefaultBulletList.WithItems(pluginItems).Render()
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

	var cfg *config.Config
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if configGlobal {
		var err error
		cfg, err = config.LoadGlobal()
		if err != nil || cfg == nil {
			cfg = &config.Config{}
		}
	} else if configPath != "" {
		cm := config.NewConfigManager()
		var err error
		cfg, err = cm.Load(ctx, configPath)
		if err != nil || cfg == nil {
			cfg = &config.Config{}
		}
	} else {
		cm := config.NewConfigManager()
		var err error
		cfg, err = cm.LoadHierarchy(ctx)
		if err != nil || cfg == nil {
			cfg = &config.Config{}
		}
	}

	// Resolve dotted key into the config struct using a simple top-level section lookup
	cleanKey := key
	if !strings.HasPrefix(cleanKey, "settings.") && (cleanKey == "browser" || cleanKey == "cache_ttl" || cleanKey == "data_dir") {
		cleanKey = "settings." + cleanKey
	}

	segments := config.SplitConfigKey(cleanKey)
	var value interface{}
	switch segments[0] {
	case "tools":
		if len(segments) > 1 {
			if t, ok := cfg.Tools[segments[1]]; ok {
				value = t
			}
		} else {
			value = cfg.Tools
		}
	case "settings":
		if len(segments) > 1 {
			switch segments[1] {
			case "browser":
				value = cfg.Settings.Browser
			case "cache_ttl":
				value = cfg.Settings.CacheTTL
			default:
				value = cfg.Settings
			}
		} else {
			value = cfg.Settings
		}
	case "env":
		if len(segments) > 1 {
			value = cfg.Env[segments[1]]
		} else {
			value = cfg.Env
		}
	case "tasks":
		value = cfg.Tasks
	case "plugins":
		if len(segments) > 1 {
			pluginID := segments[1]
			if p, ok := cfg.Plugins[pluginID]; ok {
				if len(segments) == 2 {
					value = p
				} else if segments[2] == "config" {
					if len(segments) == 3 {
						value = p.Config
					} else {
						value = p.Config[segments[3]]
					}
				} else if segments[2] == "version" {
					value = p.Version
				} else {
					value = p.Config[segments[2]]
				}
			}
		} else {
			value = cfg.Plugins
		}
	default:
		value = nil
	}

	if value == nil || value == "" {
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

// runConfigSet sets a configuration value in the local or global config file.
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

	// Validate namespace adhering to reserved core tables and plugin namespaces
	if err := config.ValidateKeyNamespace(key); err != nil {
		formatter.Error("Invalid configuration namespace", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("invalid configuration key: %w", err)
	}

	// Validate against plugin manifest schema if present
	segments := config.SplitConfigKey(key)
	if len(segments) >= 4 && segments[0] == "plugins" && segments[2] == "config" {
		pluginID := segments[1]
		property := segments[3]
		cwdDir, _ := os.Getwd()
		manifest, _ := config.FindPluginManifest(pluginID, cwdDir)
		if manifest != nil && manifest.Contributes.Configuration != nil {
			if schema, exists := manifest.Contributes.Configuration[property]; exists {
				if err := config.ValidateConfigValue(property, value, schema); err != nil {
					formatter.Error("Plugin configuration schema validation failed", map[string]interface{}{
						"plugin":   pluginID,
						"property": property,
						"error":    err.Error(),
					})
					return fmt.Errorf("plugin schema validation failed: %w", err)
				}
			}
		}
	}

	var targetConfigFile string
	if configGlobal {
		targetConfigFile = config.GetGlobalConfigPath()
		if err := os.MkdirAll(filepath.Dir(targetConfigFile), 0755); err != nil {
			formatter.Error("Failed to create global config directory", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("create global config dir: %w", err)
		}
	} else {
		// Enforce: Browser setting is strictly user-global only
		cleanKey := strings.TrimPrefix(key, "settings.")
		if cleanKey == "browser" {
			formatter.Error("Browser configuration is only permitted in global config to ensure project portability across team machines.", map[string]interface{}{
				"hint": fmt.Sprintf("Use 'unirtm config --global %s %s'", key, value),
			})
			return fmt.Errorf("browser setting is only allowed globally; please use 'unirtm config --global %s %s'", key, value)
		}

		targetConfigFile = ".unirtm.toml"
		if configPath != "" {
			targetConfigFile = configPath
		} else {
			for _, f := range []string{".unirtm.toml", ".unirtm.yaml", ".unirtm.yml", "unirtm.toml", "unirtm.yaml", "unirtm.yml"} {
				if _, err := os.Stat(f); err == nil {
					targetConfigFile = f
					break
				}
			}
		}
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(targetConfigFile), "."))
	if ext == "" {
		ext = "toml"
	}

	// Normalize settings shorthand (e.g. "browser" -> "settings.browser")
	storeKey := key
	if !strings.HasPrefix(storeKey, "settings.") && (storeKey == "browser" || storeKey == "cache_ttl") {
		storeKey = "settings." + storeKey
	}

	// Acquire exclusive file lock to ensure atomic, safe writes
	fileLock, lockErr := config.AcquireFileLock(targetConfigFile, 5*time.Second)
	if lockErr != nil {
		formatter.Error("Failed to acquire configuration file lock", map[string]interface{}{
			"error": lockErr.Error(),
		})
		return fmt.Errorf("acquire config lock: %w", lockErr)
	}
	defer fileLock.Release()

	// Serialize back to file preserving comments for TOML
	var out []byte
	var writeErr error
	switch ext {
	case "toml":
		existingRaw, _ := os.ReadFile(targetConfigFile)
		updatedTOML, err := config.UpsertTOMLSetting(string(existingRaw), config.SplitConfigKey(storeKey), value)
		if err != nil {
			writeErr = err
		} else {
			out = []byte(updatedTOML)
		}
	case "yaml", "yml":
		existing := make(map[string]interface{})
		if data, err := os.ReadFile(targetConfigFile); err == nil {
			_ = yaml.Unmarshal(data, &existing)
		}
		setNestedKey(existing, config.SplitConfigKey(storeKey), value)
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

	if err := os.WriteFile(targetConfigFile, out, 0644); err != nil {
		formatter.Error("Failed to write configuration", map[string]interface{}{
			"file":  targetConfigFile,
			"error": err.Error(),
		})
		return fmt.Errorf("write config: %w", err)
	}

	// Automatically trust the config file updated by unirtm config set
	trustMgr := config.NewTrustManager()
	_ = trustMgr.Trust(targetConfigFile)

	formatter.Success(fmt.Sprintf("Set %s = %s in %s", key, value, targetConfigFile), nil)
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
