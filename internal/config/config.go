// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	Tools ToolMap `toml:"-" yaml:"-" mapstructure:"-"`
	ToolsRaw map[string]interface{} `toml:"tools" yaml:"tools" mapstructure:"tools"`
	Env map[string]interface{} `toml:"env" yaml:"env" mapstructure:"env"`
	Settings Settings `toml:"settings" yaml:"settings" mapstructure:"settings"`
	Tasks map[string]Task `toml:"tasks" yaml:"tasks" mapstructure:"tasks"`
	Environments map[string]EnvironmentConfig `toml:"environments,omitempty" yaml:"environments,omitempty" mapstructure:"environments,omitempty"`
	Aliases map[string]map[string]string `toml:"aliases,omitempty" yaml:"aliases,omitempty" mapstructure:"aliases,omitempty"`
}

type EnvironmentConfig struct {
	Tools ToolMap `toml:"-" yaml:"-" mapstructure:"-"`
	ToolsRaw map[string]interface{} `toml:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools,omitempty"`
	Env map[string]interface{} `toml:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`
	Settings Settings `toml:"settings,omitempty" yaml:"settings,omitempty" mapstructure:"settings,omitempty"`
	Tasks map[string]Task `toml:"tasks,omitempty" yaml:"tasks,omitempty" mapstructure:"tasks,omitempty"`
}

type ToolConfig struct {
	Version string `toml:"version" yaml:"version" mapstructure:"version"`
	Backend string `toml:"backend,omitempty" yaml:"backend,omitempty" mapstructure:"backend,omitempty"`
	Provider string `toml:"provider,omitempty" yaml:"provider,omitempty" mapstructure:"provider,omitempty"`
	PreInstall string `toml:"pre_install,omitempty" yaml:"pre_install,omitempty" mapstructure:"pre_install,omitempty"`
	PostInstall string `toml:"post_install,omitempty" yaml:"post_install,omitempty" mapstructure:"post_install,omitempty"`
	GPGKeys []string `toml:"gpg_keys,omitempty" yaml:"gpg_keys,omitempty" mapstructure:"gpg_keys,omitempty"`
}

type ToolMap map[string]ToolConfig

func (c *Config) PostLoad() {
	if c.ToolsRaw != nil {
		c.Tools = make(ToolMap)
		for k, v := range c.ToolsRaw {
			c.Tools[k] = parseToolConfig(v)
		}
	}
	for name, env := range c.Environments {
		env.PostLoad()
		c.Environments[name] = env
	}
}

func (ec *EnvironmentConfig) PostLoad() {
	if ec.ToolsRaw != nil {
		ec.Tools = make(ToolMap)
		for k, v := range ec.ToolsRaw {
			ec.Tools[k] = parseToolConfig(v)
		}
	}
}

func parseToolConfig(v interface{}) ToolConfig {
	var tc ToolConfig
	switch val := v.(type) {
	case string:
		tc.Version = val
	case map[string]interface{}:
		if version, ok := val["version"].(string); ok {
			tc.Version = version
		}
		if backend, ok := val["backend"].(string); ok {
			tc.Backend = backend
		}
		if provider, ok := val["provider"].(string); ok {
			tc.Provider = provider
		}
		if preInstall, ok := val["pre_install"].(string); ok {
			tc.PreInstall = preInstall
		}
		if postInstall, ok := val["post_install"].(string); ok {
			tc.PostInstall = postInstall
		}
		if gpgKeys, ok := val["gpg_keys"].([]interface{}); ok {
			for _, gk := range gpgKeys {
				if s, ok := gk.(string); ok {
					tc.GPGKeys = append(tc.GPGKeys, s)
				}
			}
		}
	}
	return tc
}

func (tm ToolMap) MarshalTOML() (interface{}, error) {
	raw := make(map[string]interface{})
	for k, tc := range tm {
		if tc.Backend == "" && tc.Provider == "" && tc.PreInstall == "" && tc.PostInstall == "" && len(tc.GPGKeys) == 0 {
			raw[k] = tc.Version
		} else {
			raw[k] = tc
		}
	}
	return raw, nil
}

type Settings struct {
	CacheDir string `toml:"cache_dir" yaml:"cache_dir" mapstructure:"cache_dir"`
	DataDir string `toml:"data_dir" yaml:"data_dir" mapstructure:"data_dir"`
	CacheTTL int `toml:"cache_ttl" yaml:"cache_ttl" mapstructure:"cache_ttl"`
	Concurrency int `toml:"concurrency" yaml:"concurrency" mapstructure:"concurrency"`
	Lockfile bool `toml:"lockfile,omitempty" yaml:"lockfile,omitempty" mapstructure:"lockfile,omitempty"`
	Locked bool `toml:"locked,omitempty" yaml:"locked,omitempty" mapstructure:"locked,omitempty"`
	GitHubProxy string `toml:"github_proxy,omitempty" yaml:"github_proxy,omitempty" mapstructure:"github_proxy,omitempty"`
	HttpProxy string `toml:"http_proxy,omitempty" yaml:"http_proxy,omitempty" mapstructure:"http_proxy,omitempty"`
	HttpsProxy string `toml:"https_proxy,omitempty" yaml:"https_proxy,omitempty" mapstructure:"https_proxy,omitempty"`
	GitHubToken string `toml:"github_token,omitempty" yaml:"github_token,omitempty" mapstructure:"github_token,omitempty"`
	HTTPTimeout int `toml:"http_timeout,omitempty" yaml:"http_timeout,omitempty" mapstructure:"http_timeout,omitempty"`
	TaskTimeout int `toml:"task_timeout,omitempty" yaml:"task_timeout,omitempty" mapstructure:"task_timeout,omitempty"`
	TaskOutput string `toml:"task_output,omitempty" yaml:"task_output,omitempty" mapstructure:"task_output,omitempty"`
	Experimental bool `toml:"experimental,omitempty" yaml:"experimental,omitempty" mapstructure:"experimental,omitempty"`
	AutoInstall *bool `toml:"auto_install,omitempty" yaml:"auto_install,omitempty" mapstructure:"auto_install,omitempty"`
	Color string `toml:"color,omitempty" yaml:"color,omitempty" mapstructure:"color,omitempty"`
	AlwaysKeepDownload bool `toml:"always_keep_download,omitempty" yaml:"always_keep_download,omitempty" mapstructure:"always_keep_download,omitempty"`
	CeilingPaths []string `toml:"ceiling_paths,omitempty" yaml:"ceiling_paths,omitempty" mapstructure:"ceiling_paths,omitempty"`
	TrustedConfigPaths []string `toml:"trusted_config_paths,omitempty" yaml:"trusted_config_paths,omitempty" mapstructure:"trusted_config_paths,omitempty"`
	GPGVerify string `toml:"gpg_verify" yaml:"gpg_verify" mapstructure:"gpg_verify"`
	GPGKeys []string `toml:"gpg_keys" yaml:"gpg_keys" mapstructure:"gpg_keys"`
	VerifyMetadata *bool `toml:"verify_metadata,omitempty" yaml:"verify_metadata,omitempty" mapstructure:"verify_metadata,omitempty"`
	NoProxy []string `toml:"no_proxy,omitempty" yaml:"no_proxy,omitempty" mapstructure:"no_proxy,omitempty"`
	Tools map[string]map[string]interface{} `toml:"tools,omitempty" yaml:"tools,omitempty" mapstructure:"tools,omitempty"`
}

type Task struct {
	Description string `toml:"description" yaml:"description" mapstructure:"description"`
	Run string `toml:"run" yaml:"run" mapstructure:"run"`
	Env map[string]interface{} `toml:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`
	Depends []string `toml:"depends,omitempty" yaml:"depends,omitempty" mapstructure:"depends,omitempty"`
	Timeout int `toml:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout,omitempty"`
	Output string `toml:"output,omitempty" yaml:"output,omitempty" mapstructure:"output,omitempty"`
}

func (c *Config) Validate() error {
	c.PostLoad()
	var errs []string
	if c.Tools == nil { c.Tools = make(ToolMap) }
	for toolName, toolConfig := range c.Tools {
		if err := toolConfig.Validate(); err != nil { errs = append(errs, fmt.Sprintf("tool %q: %v", toolName, err)) }
	}
	if err := c.Settings.Validate(); err != nil { errs = append(errs, fmt.Sprintf("settings: %v", err)) }
	if c.Tasks == nil { c.Tasks = make(map[string]Task) }
	for taskName, task := range c.Tasks {
		if err := task.Validate(); err != nil { errs = append(errs, fmt.Sprintf("task %q: %v", taskName, err)) }
	}
	if err := c.validateTaskDependencies(); err != nil { errs = append(errs, err.Error()) }
	if c.Environments == nil { c.Environments = make(map[string]EnvironmentConfig) }
	for envName, envConfig := range c.Environments {
		if err := envConfig.Validate(); err != nil { errs = append(errs, fmt.Sprintf("environment %q: %v", envName, err)) }
	}
	if len(errs) > 0 { return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - ")) }
	return nil
}

// Merge merges another configuration into this one.
// The current configuration takes precedence over the other one (deep merge).
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	// Merge tools
	if c.Tools == nil {
		c.Tools = make(ToolMap)
	}
	for k, v := range other.Tools {
		if _, exists := c.Tools[k]; !exists {
			c.Tools[k] = v
		}
	}

	// Merge environment variables
	if c.Env == nil {
		c.Env = make(map[string]interface{})
	}
	for k, v := range other.Env {
		if _, exists := c.Env[k]; !exists {
			c.Env[k] = v
		}
	}

	// Merge tasks
	if c.Tasks == nil {
		c.Tasks = make(map[string]Task)
	}
	for k, v := range other.Tasks {
		if _, exists := c.Tasks[k]; !exists {
			c.Tasks[k] = v
		}
	}

	// Merge environments
	if c.Environments == nil {
		c.Environments = make(map[string]EnvironmentConfig)
	}
	for k, v := range other.Environments {
		if _, exists := c.Environments[k]; !exists {
			c.Environments[k] = v
		}
	}

	// Merge aliases
	if c.Aliases == nil {
		c.Aliases = make(map[string]map[string]string)
	}
	for k, v := range other.Aliases {
		if _, exists := c.Aliases[k]; !exists {
			c.Aliases[k] = v
		} else {
			for ak, av := range v {
				if _, aexists := c.Aliases[k][ak]; !aexists {
					c.Aliases[k][ak] = av
				}
			}
		}
	}
}

func (c *Config) validateTaskDependencies() error {
	var errs []string
	for taskName, task := range c.Tasks {
		for _, dep := range task.Depends {
			if _, exists := c.Tasks[dep]; !exists { errs = append(errs, fmt.Sprintf("task %q depends on non-existent task %q", taskName, dep)) }
		}
	}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var hasCycle func(string) bool
	hasCycle = func(taskName string) bool {
		visited[taskName] = true
		recStack[taskName] = true
		task, exists := c.Tasks[taskName]
		if !exists { return false }
		for _, dep := range task.Depends {
			if !visited[dep] {
				if hasCycle(dep) { return true }
			} else if recStack[dep] {
				errs = append(errs, fmt.Sprintf("circular dependency detected involving task %q", taskName))
				return true
			}
		}
		recStack[taskName] = false
		return false
	}
	for taskName := range c.Tasks {
		if !visited[taskName] { hasCycle(taskName) }
	}
	if len(errs) > 0 { return errors.New(strings.Join(errs, "; ")) }
	return nil
}

func (tc *ToolConfig) Validate() error {
	if tc.Version == "" { return errors.New("version is required") }
	return nil
}

func (s *Settings) Validate() error {
	var errs []string
	if s.CacheTTL < 0 { errs = append(errs, "cache_ttl must be non-negative") }
	if s.Concurrency < 0 { errs = append(errs, "concurrency must be non-negative") }
	if s.HTTPTimeout < 0 { errs = append(errs, "http_timeout must be non-negative") }
	if len(errs) > 0 { return errors.New(strings.Join(errs, "; ")) }
	return nil
}

func (t *Task) Validate() error {
	if t.Run == "" { return errors.New("run command is required") }
	return nil
}

func (ec *EnvironmentConfig) Validate() error {
	var errs []string
	for toolName, toolConfig := range ec.Tools {
		if err := toolConfig.Validate(); err != nil { errs = append(errs, fmt.Sprintf("tool %q: %v", toolName, err)) }
	}
	if err := ec.Settings.Validate(); err != nil { errs = append(errs, fmt.Sprintf("settings: %v", err)) }
	for taskName, task := range ec.Tasks {
		if err := task.Validate(); err != nil { errs = append(errs, fmt.Sprintf("task %q: %v", taskName, err)) }
	}
	if len(errs) > 0 { return errors.New(strings.Join(errs, "; ")) }
	return nil
}
