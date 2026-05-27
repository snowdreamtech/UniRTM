// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config provides configuration management for UniRTM.
package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"gopkg.in/yaml.v3"
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

// defaultConfigManager implements ConfigManager using native parsers for TOML/YAML.
type defaultConfigManager struct {
	// homeDir is the user's home directory for resolving global config paths
	homeDir      string
	trustManager TrustManager
}

// NewConfigManager creates a new ConfigManager instance.
//
// It parses TOML and YAML configuration files and supports
// hierarchical configuration loading with proper precedence rules.
func NewConfigManager() ConfigManager {
	homeDir, _ := os.UserHomeDir()
	return &defaultConfigManager{
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
func (m *defaultConfigManager) Load(ctx context.Context, path string) (*Config, error) {
	// Check if file exists
	if _, err := OsStat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file contents
	contentBytes, err := OsReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// 1. Smart Type Casting for Environment Variables
	// Convert common boolean strings to actual booleans to support logical checks in templates
	templateCtx := pongo2.Context{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}

	typedEnv := make(map[string]interface{})
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			k, v := pair[0], pair[1]
			lowV := strings.ToLower(v)

			// Convert to boolean if possible
			if lowV == "true" || lowV == "1" || lowV == "yes" || lowV == "on" {
				typedEnv[k] = true
			} else if lowV == "false" || lowV == "0" || lowV == "no" || lowV == "off" {
				typedEnv[k] = false
			} else {
				typedEnv[k] = v
			}
		}
	}
	templateCtx["env"] = typedEnv
	templateCtx["Env"] = typedEnv

	// 2. Context Enrichment (Mise compatibility)
	templateCtx["config"] = map[string]interface{}{
		"dir": filepath.Dir(path),
	}

	// 2.1 Tool functions for templates
	templateCtx["exec"] = func(cmd string) string {
		out, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	templateCtx["which"] = func(bin string) string {
		path, err := exec.LookPath(bin)
		if err != nil {
			return ""
		}
		return path
	}

	templateCtx["get_env"] = func(key string, defaultVal ...string) interface{} {
		val, exists := typedEnv[key]
		if !exists || val == "" {
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return ""
		}
		return val
	}

	templateCtx["file"] = func(path string) string {
		data, err := OsReadFile(path)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}

	templateCtx["exists"] = func(path string) bool {
		_, err := OsStat(path)
		return err == nil
	}

	// 3. Syntax Bridging (Jinja2 -> Pongo2)
	// We use regex to replace common Jinja2 patterns that Pongo2 doesn't support natively.
	content := bridgeJinja2(string(contentBytes))

	// Parse and execute template (Jinja2/Pongo2)
	tpl, err := pongo2.FromString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template in %s: %w", path, err)
	}

	rendered, err := tpl.Execute(templateCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to render template in %s: %w", path, err)
	}

	renderedBytes := []byte(rendered)

	// Determine config type from extension
	ext := filepath.Ext(path)
	configType := strings.TrimPrefix(ext, ".")
	if configType == "" {
		configType = "toml" // default
	}

	// Unmarshal into Config struct preserving key case
	var config Config
	switch configType {
	case "toml":
		if err := toml.Unmarshal(renderedBytes, &config); err != nil {
			return nil, fmt.Errorf("invalid syntax in configuration file %s: %w", path, err)
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(renderedBytes, &config); err != nil {
			return nil, fmt.Errorf("invalid syntax in configuration file %s: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported configuration file format %q in %s", configType, path)
	}

	// Process shorthand tool versions
	config.PostLoad()

	// Initialize maps if they are nil
	if config.Tools == nil {
		config.Tools = make(ToolMap)
	}
	if config.Env == nil {
		config.Env = make(map[string]interface{})
	}
	// Inject environment variables into current process
	for k, v := range config.Env {
		if s, ok := v.(string); ok {
			rendered := renderTemplate(s, templateCtx)
			config.Env[k] = rendered
			os.Setenv(k, rendered)
		}
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

// ResolveAlias returns the aliased version for a tool if it exists.
// Otherwise, it returns the original version request.
func (c *Config) ResolveAlias(tool, version string) string {
	if c.Aliases == nil {
		return version
	}
	if toolAliases, ok := c.Aliases[tool]; ok {
		if resolved, ok := toolAliases[version]; ok {
			return resolved
		}
	}
	return version
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
func (m *defaultConfigManager) LoadHierarchy(ctx context.Context) (*Config, error) {
	var configs []*Config

	// 1. Load System configuration
	systemPaths := []string{
		"/etc/unirtm/config.toml",
		"/etc/unirtm/config.yaml",
		"/etc/unirtm/config.yml",
	}
	for _, path := range systemPaths {
		if cfg, err := m.tryLoad(ctx, path, false, nil); cfg != nil && err == nil {
			configs = append(configs, cfg)
			break
		}
	}

	// 2. Load Global configuration
	globalPaths := []string{
		filepath.Join(env.GetConfigDir(), "config.toml"),
		filepath.Join(env.GetConfigDir(), "config.yaml"),
		filepath.Join(env.GetConfigDir(), "config.yml"),
	}
	for _, path := range globalPaths {
		if cfg, err := m.tryLoad(ctx, path, false, nil); cfg != nil && err == nil {
			configs = append(configs, cfg)
			break
		}
	}

	// Merge initial configs to get settings like CeilingPaths and TrustedConfigPaths
	initialMerged := &Config{
		Tools:   make(map[string]ToolConfig),
		Env:     make(map[string]interface{}),
		Tasks:   make(map[string]Task),
		Aliases: make(map[string]map[string]string),
	}
	for _, c := range configs {
		initialMerged, _ = m.Merge(initialMerged, c)
	}

	// 3. Discover Project and Local configs recursively UP
	cwd, _ := os.Getwd()
	curr := cwd
	var projectConfigs []*Config

	// Helper to check if a path is a ceiling
	isCeiling := func(path string) bool {
		absPath, _ := FilepathAbs(path)
		for _, cp := range initialMerged.Settings.CeilingPaths {
			absCP, _ := FilepathAbs(cp)
			if absPath == absCP {
				return true
			}
		}
		// Always stop at root
		parent := filepath.Dir(absPath)
		return parent == absPath
	}

	for {
		// Files to check in current directory (highest precedence at the end of the slice)
		files := []string{
			// 1. Mise compatibility files (Lowest precedence)
			filepath.Join(curr, ".mise.yml"),
			filepath.Join(curr, ".mise.yaml"),
			filepath.Join(curr, ".mise.toml"),

			// 2. Standard UniRTM project files
			filepath.Join(curr, "unirtm.yml"),
			filepath.Join(curr, "unirtm.yaml"),
			filepath.Join(curr, "unirtm.toml"),
			filepath.Join(curr, ".unirtm.yml"),
			filepath.Join(curr, ".unirtm.yaml"),
			filepath.Join(curr, ".unirtm.toml"),

			// 3. Local overrides (Highest precedence)
			filepath.Join(curr, ".mise.local.yml"),
			filepath.Join(curr, ".mise.local.yaml"),
			filepath.Join(curr, ".mise.local.toml"),
			filepath.Join(curr, ".unirtm.local.yml"),
			filepath.Join(curr, ".unirtm.local.yaml"),
			filepath.Join(curr, ".unirtm.local.toml"),
		}

		dirConfigs := []*Config{}
		for _, path := range files {
			cfg, err := m.tryLoad(ctx, path, true, &initialMerged.Settings)
			if err != nil {
				return nil, fmt.Errorf("failed to load configuration from %s: %w", path, err)
			}
			if cfg != nil {
				dirConfigs = append(dirConfigs, cfg)
			}
		}

		// Add discovered configs from this directory to the front of projectConfigs
		// (since deeper configs have higher precedence, and m.Merge(base, override) means override wins)
		projectConfigs = append(dirConfigs, projectConfigs...)

		if isCeiling(curr) {
			break
		}
		curr = filepath.Dir(curr)
	}

	configs = append(configs, projectConfigs...)

	// 4. Final merge
	if len(configs) == 0 {
		return &Config{
			Tools:        make(map[string]ToolConfig),
			Env:          make(map[string]interface{}),
			Tasks:        make(map[string]Task),
			Environments: make(map[string]EnvironmentConfig),
			Aliases:      make(map[string]map[string]string),
		}, nil
	}

	finalConfig, err := m.Merge(configs...)
	if err != nil {
		return nil, fmt.Errorf("failed to merge hierarchical configurations: %w", err)
	}

	return finalConfig, nil
}

// tryLoad attempts to load a config file if it exists and satisfies trust requirements.
func (m *defaultConfigManager) tryLoad(ctx context.Context, path string, enforceTrust bool, initialSettings *Settings) (*Config, error) {
	if _, err := OsStat(path); os.IsNotExist(err) {
		return nil, nil
	}

	if enforceTrust {
		// Trust check
		trustedPaths := initialSettings.TrustedConfigPaths
		isGloballyTrusted := false
		for _, tp := range trustedPaths {
			absTP, _ := FilepathAbs(tp)
			if path == absTP {
				isGloballyTrusted = true
				break
			}
		}

		if !isGloballyTrusted {
			status := m.trustManager.TrustStatus(path)
			if status == TrustStatusUntrusted {
				return nil, nil
			}
			if status == TrustStatusModified {
				cfg, err := m.Load(ctx, path)
				if err != nil {
					return nil, err
				}
				// Strip sensitive fields
				cfg.Env = make(map[string]interface{})
				cfg.Tasks = make(map[string]Task)
				// Also strip from environments
				for name, env := range cfg.Environments {
					env.Env = nil
					env.Tasks = nil
					cfg.Environments[name] = env
				}
				return cfg, nil
			}
			if status == TrustStatusTrusted {
				return m.Load(ctx, path)
			}
		} else {
			return m.Load(ctx, path)
		}
	} else {
		return m.Load(ctx, path)
	}

	return nil, nil
}

// LoadWithEnvironment loads configuration from all hierarchy levels and applies
// environment-specific overrides for the specified environment.
//
// This method first loads the hierarchical configuration, then applies the
// environment-specific overrides if the environment exists in the configuration.
//
// Returns the merged configuration with environment overrides applied, or an error
// if loading or merging fails.
func (m *defaultConfigManager) LoadWithEnvironment(ctx context.Context, environment string) (*Config, error) {
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
func (m *defaultConfigManager) Validate(ctx context.Context, config *Config) error {
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
func (m *defaultConfigManager) Merge(configs ...*Config) (*Config, error) {
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
		if config.Settings.Jobs != 0 {
			merged.Settings.Jobs = config.Settings.Jobs
		}

		if config.Settings.GitHubProxy != "" {
			merged.Settings.GitHubProxy = config.Settings.GitHubProxy
		}
		if config.Settings.HttpProxy != "" {
			merged.Settings.HttpProxy = config.Settings.HttpProxy
		}
		if config.Settings.HttpsProxy != "" {
			merged.Settings.HttpsProxy = config.Settings.HttpsProxy
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
		if config.Settings.VerifyMetadata != nil {
			merged.Settings.VerifyMetadata = config.Settings.VerifyMetadata
		}
		if len(config.Settings.CeilingPaths) > 0 {
			merged.Settings.CeilingPaths = append(merged.Settings.CeilingPaths, config.Settings.CeilingPaths...)
		}
		if config.Settings.TrustedConfigPaths != nil {
			merged.Settings.TrustedConfigPaths = append(merged.Settings.TrustedConfigPaths, config.Settings.TrustedConfigPaths...)
		}
		if config.Settings.Tools != nil {
			if merged.Settings.Tools == nil {
				merged.Settings.Tools = make(map[string]map[string]interface{})
			}
			for toolName, settings := range config.Settings.Tools {
				if merged.Settings.Tools[toolName] == nil {
					merged.Settings.Tools[toolName] = make(map[string]interface{})
				}
				for k, v := range settings {
					merged.Settings.Tools[toolName][k] = v
				}
			}
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
func (m *defaultConfigManager) ApplyEnvironment(config *Config, environment string) (*Config, error) {
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
	if envConfig.Settings.Jobs != 0 {
		result.Settings.Jobs = envConfig.Settings.Jobs
	}

	if envConfig.Settings.GitHubProxy != "" {
		result.Settings.GitHubProxy = envConfig.Settings.GitHubProxy
	}
	if envConfig.Settings.HttpProxy != "" {
		result.Settings.HttpProxy = envConfig.Settings.HttpProxy
	}
	if envConfig.Settings.HttpsProxy != "" {
		result.Settings.HttpsProxy = envConfig.Settings.HttpsProxy
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
	if envConfig.Settings.VerifyMetadata != nil {
		result.Settings.VerifyMetadata = envConfig.Settings.VerifyMetadata
	}

	return result, nil
}

func renderTemplate(content string, ctx pongo2.Context) string {
	if !strings.Contains(content, "{%") && !strings.Contains(content, "{{") {
		return content
	}

	// Syntax Bridging (Jinja2 -> Pongo2)
	content = bridgeJinja2(content)

	tpl, err := pongo2.FromString(content)
	if err != nil {
		return content
	}
	rendered, err := tpl.Execute(ctx)
	if err != nil {
		return content
	}
	return rendered
}

// bridgeJinja2 replaces common Jinja2 patterns that Pongo2 doesn't support natively.
func bridgeJinja2(content string) string {
	// Replace 'is defined' -> '' (Pongo2 treats existence as truthy)
	// Supports dotted names like env.CI
	reDefined := regexp.MustCompile(`([\w.]+)\s+is\s+defined`)
	content = reDefined.ReplaceAllString(content, "$1")

	// Replace 'is undefined' -> 'not $1'
	reUndefined := regexp.MustCompile(`([\w.]+)\s+is\s+undefined`)
	content = reUndefined.ReplaceAllString(content, "not $1")

	// Replace '~' (Jinja2 string concat) -> '+' (Pongo2 string concat)
	content = strings.ReplaceAll(content, " ~ ", " + ")

	return content
}
