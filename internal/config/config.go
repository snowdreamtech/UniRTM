// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config provides configuration data structures and validation for UniRTM.
//
// This package defines the core configuration types used throughout the UniRTM system,
// including tool configurations, settings, and tasks. All structures support both
// TOML and YAML serialization formats.
package config

import (
	"errors"
	"fmt"
	"strings"
)

// Config represents the root configuration structure for UniRTM.
//
// It contains tool version specifications, environment variables, global settings,
// and task definitions. Configuration files can be written in either TOML or YAML
// format and are loaded hierarchically (system → global → project → local).
type Config struct {
	// Tools maps tool names to version specifications.
	// Example: {"node": {Version: "20.0.0"}, "python": {Version: "3.11.0"}}
	Tools map[string]ToolConfig `toml:"tools" yaml:"tools" mapstructure:"tools"`

	// Env contains environment variable definitions.
	// These variables are set when activating the toolchain.
	Env map[string]interface{} `toml:"env" yaml:"env" mapstructure:"env"`

	// Settings contains global settings for UniRTM behavior.
	Settings Settings `toml:"settings" yaml:"settings" mapstructure:"settings"`

	// Tasks contains task definitions that can be executed via the CLI.
	Tasks map[string]Task `toml:"tasks" yaml:"tasks" mapstructure:"tasks"`

	// Environments contains environment-specific configuration overrides.
	// Keys are environment names (e.g., "development", "staging", "production").
	// Values are partial configurations that override the base configuration.
	Environments map[string]EnvironmentConfig `toml:"environments,omitempty" yaml:"environments,omitempty" mapstructure:"environments,omitempty"`

	// Aliases maps tool names to their version aliases.
	// Example: {"node": {"lts": "20.11.0", "latest": "22.0.0"}}
	Aliases map[string]map[string]string `toml:"aliases,omitempty" yaml:"aliases,omitempty" mapstructure:"aliases,omitempty"`
}

// EnvironmentConfig represents environment-specific configuration overrides.
//
// It contains the same structure as Config but all fields are optional.
// When an environment is selected, these values override the base configuration.
type EnvironmentConfig struct {
	// Tools maps tool names to version specifications for this environment.
	// These override the base tool configurations.
	Tools map[string]ToolConfig `toml:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools,omitempty"`

	// Env contains environment variable definitions for this environment.
	// These are merged with the base environment variables.
	Env map[string]interface{} `toml:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`

	// Settings contains settings overrides for this environment.
	Settings Settings `toml:"settings,omitempty" yaml:"settings,omitempty" mapstructure:"settings,omitempty"`

	// Tasks contains task definitions for this environment.
	// These override the base task definitions.
	Tasks map[string]Task `toml:"tasks,omitempty" yaml:"tasks,omitempty" mapstructure:"tasks,omitempty"`
}

// ToolConfig specifies the configuration for a single tool.
//
// It defines the version to install and optionally specifies which backend
// and provider to use for installation.
type ToolConfig struct {
	// Version is the version specification (exact, range, or alias).
	// Examples: "1.20.0", ">=1.20.0", "^3.11", "latest", "lts"
	Version string `toml:"version" yaml:"version" mapstructure:"version"`

	// Backend specifies the backend to use (optional).
	// If not specified, the system will select an appropriate backend.
	// Examples: "github", "aqua", "http"
	Backend string `toml:"backend,omitempty" yaml:"backend,omitempty" mapstructure:"backend,omitempty"`

	// Provider specifies the provider to use (optional).
	// If not specified, the system will select an appropriate provider.
	// Examples: "node", "python", "go", "generic"
	Provider string `toml:"provider,omitempty" yaml:"provider,omitempty" mapstructure:"provider,omitempty"`

	// PreInstall is a command to run before the tool is installed.
	PreInstall string `toml:"pre_install,omitempty" yaml:"pre_install,omitempty" mapstructure:"pre_install,omitempty"`

	// PostInstall is a command to run after the tool is installed successfully.
	PostInstall string `toml:"post_install,omitempty" yaml:"post_install,omitempty" mapstructure:"post_install,omitempty"`

	// GPGKeys is a tool-specific list of trusted GPG fingerprints.
	GPGKeys []string `toml:"gpg_keys,omitempty" yaml:"gpg_keys,omitempty" mapstructure:"gpg_keys,omitempty"`
}

// UnmarshalTOML implements custom unmarshalling for ToolConfig to support both
// a simple version string (shorthand) and a full configuration table.
func (tc *ToolConfig) UnmarshalTOML(data []byte) error {
	// Try unmarshalling as a string first (shorthand: "1.2.3")
	var s string
	if err := toml.Unmarshal(data, &s); err == nil {
		tc.Version = s
		return nil
	}

	// Try unmarshalling as a table (full: { version = "1.2.3", ... })
	type toolConfigAlias ToolConfig
	var alias toolConfigAlias
	if err := toml.Unmarshal(data, &alias); err != nil {
		return err
	}
	*tc = ToolConfig(alias)
	return nil
}

// MarshalTOML implements custom marshalling for ToolConfig to use the shorthand
// format (simple string) when only the version is specified.
func (tc ToolConfig) MarshalTOML() (interface{}, error) {
	if tc.Backend == "" && tc.Provider == "" && tc.PreInstall == "" && tc.PostInstall == "" && len(tc.GPGKeys) == 0 {
		return tc.Version, nil
	}

	// Use an alias to avoid infinite recursion
	type toolConfigAlias ToolConfig
	return toolConfigAlias(tc), nil
}


// Settings contains global settings for UniRTM behavior.
//
// These settings control caching, data storage, and operational parameters.
type Settings struct {
	// CacheDir is the directory for cached downloads.
	// If empty, defaults to the system cache directory.
	CacheDir string `toml:"cache_dir" yaml:"cache_dir" mapstructure:"cache_dir"`

	// DataDir is the directory for SQLite database and state.
	// If empty, defaults to the system data directory.
	DataDir string `toml:"data_dir" yaml:"data_dir" mapstructure:"data_dir"`

	// CacheTTL is the default cache TTL in seconds.
	// If zero, defaults to 86400 (24 hours).
	CacheTTL int `toml:"cache_ttl" yaml:"cache_ttl" mapstructure:"cache_ttl"`

	// Concurrency is the maximum number of concurrent operations.
	// If zero, defaults to the number of CPU cores.
	Concurrency int `toml:"concurrency" yaml:"concurrency" mapstructure:"concurrency"`

	// Lockfile controls whether unirtm.lock is generated and kept up to date.
	// When true, every successful `unirtm install` writes the resolved URL and
	// checksum back into unirtm.lock so that future installs can skip the API call.
	// Default: false (opt-in, same as mise).
	Lockfile bool `toml:"lockfile,omitempty" yaml:"lockfile,omitempty" mapstructure:"lockfile,omitempty"`

	// Locked enforces strict lockfile mode.
	// When true, `unirtm install` fails if a tool's URL is not already present
	// in unirtm.lock.  Useful in CI to guarantee fully reproducible, API-free builds.
	// Can also be set via the UNIRTM_LOCKED=1 environment variable.
	// Default: false.
	Locked bool `toml:"locked,omitempty" yaml:"locked,omitempty" mapstructure:"locked,omitempty"`

	// GitHubProxy is a proxy prefix for GitHub URLs to improve accessibility.
	// Example: "https://ghproxy.com/"
	GitHubProxy string `toml:"github_proxy,omitempty" yaml:"github_proxy,omitempty" mapstructure:"github_proxy,omitempty"`

	// GitHubToken is a personal access token for GitHub API to avoid rate limiting.
	GitHubToken string `toml:"github_token,omitempty" yaml:"github_token,omitempty" mapstructure:"github_token,omitempty"`

	// HTTPTimeout is the default timeout for HTTP requests in seconds.
	// If zero, defaults to 900 (15 minutes).
	HTTPTimeout int `toml:"http_timeout,omitempty" yaml:"http_timeout,omitempty" mapstructure:"http_timeout,omitempty"`

	// TaskTimeout is the default timeout for task execution in seconds.
	// If zero, tasks run indefinitely until completed or manually terminated.
	TaskTimeout int `toml:"task_timeout,omitempty" yaml:"task_timeout,omitempty" mapstructure:"task_timeout,omitempty"`

	// TaskOutput controls how task output is displayed.
	// Supported: "plain", "prefix", "interleave". Default: "plain".
	TaskOutput string `toml:"task_output,omitempty" yaml:"task_output,omitempty" mapstructure:"task_output,omitempty"`

	// Experimental enables experimental features.
	Experimental bool `toml:"experimental,omitempty" yaml:"experimental,omitempty" mapstructure:"experimental,omitempty"`

	// AutoInstall controls whether missing tools are automatically installed.
	// Default: true.
	AutoInstall *bool `toml:"auto_install,omitempty" yaml:"auto_install,omitempty" mapstructure:"auto_install,omitempty"`

	// Color controls whether color is enabled in the output.
	// Supported: "auto", "always", "never". Default: "auto".
	Color string `toml:"color,omitempty" yaml:"color,omitempty" mapstructure:"color,omitempty"`

	// AlwaysKeepDownload controls whether artifacts are kept after installation.
	// Default: false.
	AlwaysKeepDownload bool `toml:"always_keep_download,omitempty" yaml:"always_keep_download,omitempty" mapstructure:"always_keep_download,omitempty"`

	// CeilingPaths specifies directory levels where configuration discovery should stop.
	CeilingPaths []string `toml:"ceiling_paths,omitempty" yaml:"ceiling_paths,omitempty" mapstructure:"ceiling_paths,omitempty"`

	// TrustedConfigPaths specifies directory paths where configurations are automatically trusted.
	TrustedConfigPaths []string `toml:"trusted_config_paths,omitempty" yaml:"trusted_config_paths,omitempty" mapstructure:"trusted_config_paths,omitempty"`

	// GPGVerify controls GPG signature verification behavior.
	// Supported: "strict" (fail if invalid), "warn" (log if invalid), "off" (skip).
	// GPGVerify sets the GPG verification level: strict, warn, off
	GPGVerify string `toml:"gpg_verify" yaml:"gpg_verify" mapstructure:"gpg_verify"`

	// GPGKeys is a list of trusted GPG fingerprints for all tools.
	GPGKeys []string `toml:"gpg_keys" yaml:"gpg_keys" mapstructure:"gpg_keys"`

	// Tools contains tool-specific settings.
	// Example: [settings.node]\ninstall_pnpm = true
	Tools map[string]map[string]interface{} `toml:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools,omitempty"`
}

// Task represents a task definition that can be executed via the CLI.
//
// Tasks are user-defined commands that can be run in the context of the
// configured toolchain.
type Task struct {
	// Description is a human-readable description of the task.
	Description string `toml:"description" yaml:"description" mapstructure:"description"`

	// Run is the command to execute.
	// This can be a single command or a multi-line script.
	Run string `toml:"run" yaml:"run" mapstructure:"run"`

	// Env contains task-specific environment variables.
	// These are merged with the base environment variables.
	Env map[string]interface{} `toml:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`

	// Depends lists task names that must run before this task.
	Depends []string `toml:"depends,omitempty" yaml:"depends,omitempty" mapstructure:"depends,omitempty"`

	// Timeout is the timeout for this specific task in seconds.
	// Overrides global Settings.TaskTimeout.
	Timeout int `toml:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout,omitempty"`

	// Output controls how output is displayed for this specific task.
	// Overrides global Settings.TaskOutput.
	Output string `toml:"output,omitempty" yaml:"output,omitempty" mapstructure:"output,omitempty"`
}

// Validate validates the Config structure and returns an error if validation fails.
//
// This method checks that all required fields are present and that values are
// within acceptable ranges. It returns a descriptive error identifying all
// validation failures, not just the first one.
func (c *Config) Validate() error {
	var errs []string

	// Validate Tools
	if c.Tools == nil {
		c.Tools = make(map[string]ToolConfig)
	}

	for toolName, toolConfig := range c.Tools {
		if err := toolConfig.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("tool %q: %v", toolName, err))
		}
	}

	// Validate Settings
	if err := c.Settings.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("settings: %v", err))
	}

	// Validate Tasks
	if c.Tasks == nil {
		c.Tasks = make(map[string]Task)
	}

	for taskName, task := range c.Tasks {
		if err := task.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("task %q: %v", taskName, err))
		}
	}

	// Validate task dependencies
	if err := c.validateTaskDependencies(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Environments
	if c.Environments == nil {
		c.Environments = make(map[string]EnvironmentConfig)
	}

	for envName, envConfig := range c.Environments {
		if err := envConfig.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("environment %q: %v", envName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// validateTaskDependencies checks that all task dependencies reference existing tasks
// and that there are no circular dependencies.
func (c *Config) validateTaskDependencies() error {
	var errs []string

	// Check that all dependencies reference existing tasks
	for taskName, task := range c.Tasks {
		for _, dep := range task.Depends {
			if _, exists := c.Tasks[dep]; !exists {
				errs = append(errs, fmt.Sprintf("task %q depends on non-existent task %q", taskName, dep))
			}
		}
	}

	// Check for circular dependencies using depth-first search
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(taskName string) bool {
		visited[taskName] = true
		recStack[taskName] = true

		task, exists := c.Tasks[taskName]
		if !exists {
			return false
		}

		for _, dep := range task.Depends {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				errs = append(errs, fmt.Sprintf("circular dependency detected involving task %q", taskName))
				return true
			}
		}

		recStack[taskName] = false
		return false
	}

	for taskName := range c.Tasks {
		if !visited[taskName] {
			hasCycle(taskName)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates the ToolConfig structure and returns an error if validation fails.
//
// This method checks that the version field is present and non-empty.
func (tc *ToolConfig) Validate() error {
	if tc.Version == "" {
		return errors.New("version is required")
	}

	// Version can be an exact version, a range, or an alias
	// We don't validate the format here as that's handled by the version parser
	// Just ensure it's not empty

	return nil
}

// Validate validates the Settings structure and returns an error if validation fails.
//
// This method checks that numeric values are non-negative and within reasonable ranges.
func (s *Settings) Validate() error {
	var errs []string

	// CacheDir and DataDir can be empty (will use defaults)
	// But if specified, they should be valid paths (we don't validate path format here)

	// CacheTTL must be non-negative
	if s.CacheTTL < 0 {
		errs = append(errs, "cache_ttl must be non-negative")
	}

	// Concurrency must be non-negative
	if s.Concurrency < 0 {
		errs = append(errs, "concurrency must be non-negative")
	}

	// HTTPTimeout must be non-negative
	if s.HTTPTimeout < 0 {
		errs = append(errs, "http_timeout must be non-negative")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates the Task structure and returns an error if validation fails.
//
// This method checks that the run command is present and non-empty.
func (t *Task) Validate() error {
	if t.Run == "" {
		return errors.New("run command is required")
	}

	return nil
}

// Validate validates the EnvironmentConfig structure and returns an error if validation fails.
//
// This method validates all tools, tasks, and settings within the environment configuration.
func (ec *EnvironmentConfig) Validate() error {
	var errs []string

	// Validate Tools
	for toolName, toolConfig := range ec.Tools {
		if err := toolConfig.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("tool %q: %v", toolName, err))
		}
	}

	// Validate Settings
	if err := ec.Settings.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("settings: %v", err))
	}

	// Validate Tasks
	for taskName, task := range ec.Tasks {
		if err := task.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("task %q: %v", taskName, err))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
