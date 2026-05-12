// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config provides configuration management for UniRTM.
package config

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/viper"
)

// ConfigManager defines the interface for configuration management operations.
//
// It provides methods for loading, validating, and merging configuration files
// from multiple sources with hierarchical precedence rules.
type ConfigManager interface {
	// Load loads configuration from the specified path.
	// Returns an error if the file cannot be read or parsed.
	Load(ctx context.Context, path string) (*Config, error)

	// LoadHierarchy loads configuration from all hierarchy levels.
	// Hierarchy: system → global → project → local
	// Returns the merged configuration with local overriding project overriding global overriding system.
	LoadHierarchy(ctx context.Context) (*Config, error)

	// LoadWithEnvironment loads configuration from all hierarchy levels and applies
	// environment-specific overrides for the specified environment.
	// Returns the merged configuration with environment overrides applied.
	LoadWithEnvironment(ctx context.Context, environment string) (*Config, error)

	// Validate validates the configuration.
	// Returns an error if validation fails, with all validation errors reported.
	Validate(ctx context.Context, config *Config) error

	// Merge merges multiple configurations with precedence rules.
	// Later configurations in the slice override earlier ones.
	// Returns the merged configuration or an error if merging fails.
	Merge(configs ...*Config) (*Config, error)

	// ApplyEnvironment applies environment-specific overrides to a configuration.
	// Returns a new configuration with the environment overrides applied.
	ApplyEnvironment(config *Config, environment string) (*Config, error)
}

// viperConfigManager implements ConfigManager using Viper for TOML/YAML parsing.
type viperConfigManager struct {
	// homeDir is the user's home directory for resolving global config paths
	homeDir      string
	trustManager TrustManager
}

// NewConfigManager creates a new ConfigManager instance.
//
// It uses Viper for parsing TOML and YAML configuration files and supports
// hierarchical configuration loading with proper precedence rules.
func NewConfigManager() ConfigManager {
	homeDir, _ := os.UserHomeDir()
	return &viperConfigManager{
		homeDir:      homeDir,
		trustManager: NewTrustManager(),
	}
}

// Load loads configuration from the specified path.
//
// The file format (TOML or YAML) is automatically detected from the file extension.
// Supported extensions: .toml, .yaml, .yml
//
// Returns an error if:
//   - The file does not exist
//   - The file cannot be read
//   - The file contains invalid syntax
//   - The file format is not supported
func (m *viperConfigManager) Load(ctx context.Context, path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file contents
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// Prepare template context
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}

	tmplData := struct {
		Env  map[string]string
		OS   string
		Arch string
	}{
		Env:  envMap,
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(path)).Parse(string(contentBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template in %s: %w", path, err)
	}

	var renderedBuf bytes.Buffer
	if err := tmpl.Execute(&renderedBuf, tmplData); err != nil {
		return nil, fmt.Errorf("failed to render template in %s: %w", path, err)
	}

	// Create a new Viper instance for this file
	v := viper.New()

	// Determine config type from extension
	ext := filepath.Ext(path)
	configType := strings.TrimPrefix(ext, ".")
	if configType == "" {
		configType = "toml" // default
	}
	v.SetConfigType(configType)

	// Read the configuration file from buffer
	if err := v.ReadConfig(&renderedBuf); err != nil {
		// Provide descriptive error messages for common issues
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("configuration file not found: %s", path)
		}
		// Check for syntax errors
		if strings.Contains(err.Error(), "toml") || strings.Contains(err.Error(), "yaml") {
			return nil, fmt.Errorf("invalid syntax in configuration file %s: %w", path, err)
		}
		return nil, fmt.Errorf("failed to parse rendered configuration file %s: %w", path, err)
	}

	// Unmarshal into Config struct
	var config Config

	// Configure Viper to use the correct struct tags
	// Viper uses mapstructure by default, but we need to support both toml and yaml tags
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file %s: %w", path, err)
	}

	// Initialize maps if they are nil
	if config.Tools == nil {
		config.Tools = make(map[string]ToolConfig)
	}
	if config.Env == nil {
		config.Env = make(map[string]interface{})
	}
	if config.Tasks == nil {
		config.Tasks = make(map[string]Task)
	}
	if config.Environments == nil {
		config.Environments = make(map[string]EnvironmentConfig)
	}
	if config.Aliases == nil {
		config.Aliases = make(map[string]map[string]string)
	}

	return &config, nil
}

// LoadHierarchy loads configuration from all hierarchy levels.
//
// Configuration hierarchy (lowest to highest precedence):
//  1. System: /etc/unirtm/config.toml
//  2. Global: ~/.config/unirtm/config.toml
//  3. Project: ./unirtm.toml or ./.unirtm.toml
//  4. Local: ./.unirtm.local.toml
//
// Each level overrides values from lower levels. Missing files at any level
// are silently skipped (not an error).
//
// Returns the merged configuration or an error if any existing file fails to parse.
func (m *viperConfigManager) LoadHierarchy(ctx context.Context) (*Config, error) {
	var configs []*Config

	hierarchyLevels := [][]string{
		{ // System
			"/etc/unirtm/config.toml",
			"/etc/unirtm/config.yaml",
			"/etc/unirtm/config.yml",
		},
		{ // Global
			filepath.Join(env.GetConfigDir(), "config.toml"),
			filepath.Join(env.GetConfigDir(), "config.yaml"),
			filepath.Join(env.GetConfigDir(), "config.yml"),
		},
		{ // Project (primary)
			"./unirtm.toml",
			"./unirtm.yaml",
			"./unirtm.yml",
		},
		{ // Project (alternate)
			"./.unirtm.toml",
			"./.unirtm.yaml",
			"./.unirtm.yml",
		},
		{ // Local
			"./.unirtm.local.toml",
			"./.unirtm.local.yaml",
			"./.unirtm.local.yml",
		},
	}

	// Load each configuration level
	for i, level := range hierarchyLevels {
		for _, path := range level {
			// Skip if file doesn't exist
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}

			status := TrustStatusTrusted
			if m.trustManager != nil {
				status = m.trustManager.TrustStatus(path)
			}

			// Enforce trust for Project and Local configs (indices >= 2)
			if i >= 2 && status != TrustStatusTrusted {
				if status == TrustStatusModified {
					pterm.Warning.Printfln("Configuration file has been modified since it was last trusted: %s\nRun `unirtm trust %s` to review and trust the new contents.", path, path)
				} else {
					pterm.Warning.Printfln("Skipping untrusted configuration file: %s\nRun `unirtm trust %s` to trust it.", path, path)
				}
				break // Do not try fallback formats if the file exists but is untrusted
			}

			// Load the configuration
			config, err := m.Load(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("failed to load configuration from %s: %w", path, err)
			}

			configs = append(configs, config)
			break // Only load one file per level (TOML has priority)
		}
	}

	// If no configuration files found, return empty config
	if len(configs) == 0 {
		return &Config{
			Tools:        make(map[string]ToolConfig),
			Env:          make(map[string]interface{}),
			Tasks:        make(map[string]Task),
			Environments: make(map[string]EnvironmentConfig),
			Aliases:      make(map[string]map[string]string),
		}, nil
	}

	// Merge all configurations
	merged, err := m.Merge(configs...)
	if err != nil {
		return nil, fmt.Errorf("failed to merge hierarchical configurations: %w", err)
	}

	return merged, nil
}

// LoadWithEnvironment loads configuration from all hierarchy levels and applies
// environment-specific overrides for the specified environment.
//
// This method first loads the hierarchical configuration, then applies the
// environment-specific overrides if the environment exists in the configuration.
//
// Returns the merged configuration with environment overrides applied, or an error
// if loading or merging fails.
func (m *viperConfigManager) LoadWithEnvironment(ctx context.Context, environment string) (*Config, error) {
	// Load base configuration from hierarchy
	config, err := m.LoadHierarchy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load hierarchical configuration: %w", err)
	}

	// Apply environment-specific overrides if environment is specified
	if environment != "" {
		config, err = m.ApplyEnvironment(config, environment)
		if err != nil {
			return nil, fmt.Errorf("failed to apply environment %q: %w", environment, err)
		}
	}

	return config, nil
}

// Validate validates the configuration.
//
// This delegates to the Config.Validate() method which performs comprehensive
// validation of all configuration fields and returns all validation errors.
//
// Returns an error if validation fails, with all validation errors reported.
func (m *viperConfigManager) Validate(ctx context.Context, config *Config) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	return config.Validate()
}

// Merge merges multiple configurations with precedence rules.
//
// Later configurations in the slice override earlier ones. For maps (Tools, Env, Tasks),
// keys from later configs override keys from earlier configs. For scalar fields in Settings,
// non-zero values from later configs override earlier values.
//
// Precedence rules:
//   - Tools: Later tool definitions completely replace earlier ones (per tool name)
//   - Env: Later environment variables override earlier ones (per variable name)
//   - Tasks: Later task definitions completely replace earlier ones (per task name)
//   - Settings: Non-zero values from later configs override earlier values
//
// Returns the merged configuration or an error if merging fails.
func (m *viperConfigManager) Merge(configs ...*Config) (*Config, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configurations to merge")
	}

	// Start with an empty configuration
	merged := &Config{
		Tools:        make(map[string]ToolConfig),
		Env:          make(map[string]interface{}),
		Tasks:        make(map[string]Task),
		Environments: make(map[string]EnvironmentConfig),
		Aliases:      make(map[string]map[string]string),
	}

	// Merge each configuration in order
	for i, config := range configs {
		if config == nil {
			return nil, fmt.Errorf("configuration at index %d is nil", i)
		}

		// Merge Tools (later overrides earlier)
		for toolName, toolConfig := range config.Tools {
			merged.Tools[toolName] = toolConfig
		}

		// Merge Env (later overrides earlier)
		for envKey, envValue := range config.Env {
			merged.Env[envKey] = envValue
		}

		// Merge Tasks (later overrides earlier)
		for taskName, task := range config.Tasks {
			merged.Tasks[taskName] = task
		}

		// Merge Environments (later overrides earlier)
		for envName, envConfig := range config.Environments {
			merged.Environments[envName] = envConfig
		}

		// Merge Aliases (later overrides earlier)
		for toolName, aliases := range config.Aliases {
			if merged.Aliases[toolName] == nil {
				merged.Aliases[toolName] = make(map[string]string)
			}
			for aliasName, aliasVersion := range aliases {
				merged.Aliases[toolName][aliasName] = aliasVersion
			}
		}

		// Merge Settings (non-zero values override)
		if config.Settings.CacheDir != "" {
			merged.Settings.CacheDir = config.Settings.CacheDir
		}
		if config.Settings.DataDir != "" {
			merged.Settings.DataDir = config.Settings.DataDir
		}
		if config.Settings.CacheTTL != 0 {
			merged.Settings.CacheTTL = config.Settings.CacheTTL
		}
		if config.Settings.Concurrency != 0 {
			merged.Settings.Concurrency = config.Settings.Concurrency
		}
		if config.Settings.GitHubProxy != "" {
			merged.Settings.GitHubProxy = config.Settings.GitHubProxy
		}
		if config.Settings.GitHubToken != "" {
			merged.Settings.GitHubToken = config.Settings.GitHubToken
		}
		if config.Settings.HTTPTimeout != 0 {
			merged.Settings.HTTPTimeout = config.Settings.HTTPTimeout
		}
		if config.Settings.TaskTimeout != 0 {
			merged.Settings.TaskTimeout = config.Settings.TaskTimeout
		}
		if config.Settings.TaskOutput != "" {
			merged.Settings.TaskOutput = config.Settings.TaskOutput
		}
		if config.Settings.AutoInstall != nil {
			merged.Settings.AutoInstall = config.Settings.AutoInstall
		}
		if config.Settings.Color != "" {
			merged.Settings.Color = config.Settings.Color
		}
		if config.Settings.AlwaysKeepDownload {
			merged.Settings.AlwaysKeepDownload = config.Settings.AlwaysKeepDownload
		}
	}

	return merged, nil
}

// ApplyEnvironment applies environment-specific overrides to a configuration.
//
// This method takes a base configuration and applies the overrides from the specified
// environment. If the environment does not exist in the configuration, it returns an error.
//
// The merging follows these rules:
//   - Tools: Environment tools override base tools (per tool name)
//   - Env: Environment variables are merged with base variables (environment overrides base)
//   - Tasks: Environment tasks override base tasks (per task name)
//   - Settings: Non-zero environment settings override base settings
//
// Returns a new configuration with the environment overrides applied.
func (m *viperConfigManager) ApplyEnvironment(config *Config, environment string) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is nil")
	}

	if environment == "" {
		return nil, fmt.Errorf("environment name is empty")
	}

	// Check if the environment exists
	envConfig, exists := config.Environments[environment]
	if !exists {
		return nil, fmt.Errorf("environment %q not found in configuration", environment)
	}

	// Create a copy of the base configuration
	result := &Config{
		Tools:        make(map[string]ToolConfig),
		Env:          make(map[string]interface{}),
		Tasks:        make(map[string]Task),
		Environments: make(map[string]EnvironmentConfig),
		Aliases:      make(map[string]map[string]string),
		Settings:     config.Settings,
	}

	// Copy base tools
	for toolName, toolConfig := range config.Tools {
		result.Tools[toolName] = toolConfig
	}

	// Copy base environment variables
	for envKey, envValue := range config.Env {
		result.Env[envKey] = envValue
	}

	// Copy base tasks
	for taskName, task := range config.Tasks {
		result.Tasks[taskName] = task
	}

	// Copy environments (for reference, though typically not used after applying)
	for envName, env := range config.Environments {
		result.Environments[envName] = env
	}

	// Copy aliases
	for toolName, aliases := range config.Aliases {
		result.Aliases[toolName] = make(map[string]string)
		for aliasName, aliasVersion := range aliases {
			result.Aliases[toolName][aliasName] = aliasVersion
		}
	}

	// Apply environment-specific tool overrides
	for toolName, toolConfig := range envConfig.Tools {
		result.Tools[toolName] = toolConfig
	}

	// Apply environment-specific environment variable overrides
	for envKey, envValue := range envConfig.Env {
		result.Env[envKey] = envValue
	}

	// Apply environment-specific task overrides
	for taskName, task := range envConfig.Tasks {
		result.Tasks[taskName] = task
	}

	// Apply environment-specific settings overrides (non-zero values)
	if envConfig.Settings.CacheDir != "" {
		result.Settings.CacheDir = envConfig.Settings.CacheDir
	}
	if envConfig.Settings.DataDir != "" {
		result.Settings.DataDir = envConfig.Settings.DataDir
	}
	if envConfig.Settings.CacheTTL != 0 {
		result.Settings.CacheTTL = envConfig.Settings.CacheTTL
	}
	if envConfig.Settings.Concurrency != 0 {
		result.Settings.Concurrency = envConfig.Settings.Concurrency
	}
	if envConfig.Settings.GitHubProxy != "" {
		result.Settings.GitHubProxy = envConfig.Settings.GitHubProxy
	}
	if envConfig.Settings.GitHubToken != "" {
		result.Settings.GitHubToken = envConfig.Settings.GitHubToken
	}
	if envConfig.Settings.HTTPTimeout != 0 {
		result.Settings.HTTPTimeout = envConfig.Settings.HTTPTimeout
	}
	if envConfig.Settings.TaskTimeout != 0 {
		result.Settings.TaskTimeout = envConfig.Settings.TaskTimeout
	}
	if envConfig.Settings.TaskOutput != "" {
		result.Settings.TaskOutput = envConfig.Settings.TaskOutput
	}
	if envConfig.Settings.AutoInstall != nil {
		result.Settings.AutoInstall = envConfig.Settings.AutoInstall
	}
	if envConfig.Settings.Color != "" {
		result.Settings.Color = envConfig.Settings.Color
	}
	if envConfig.Settings.AlwaysKeepDownload {
		result.Settings.AlwaysKeepDownload = envConfig.Settings.AlwaysKeepDownload
	}

	return result, nil
}
