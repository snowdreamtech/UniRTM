// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTrustManager struct{}

func (m *mockTrustManager) IsTrusted(path string) bool { return true }
func (m *mockTrustManager) Trust(path string) error    { return nil }
func (m *mockTrustManager) Untrust(path string) error  { return nil }

func newTestConfigManager() *viperConfigManager {
	m := NewConfigManager().(*viperConfigManager)
	m.trustManager = &mockTrustManager{}
	return m
}

// TestConfigManager_Load tests the Load method with various scenarios
func TestConfigManager_Load(t *testing.T) {
	ctx := context.Background()
	manager := newTestConfigManager()

	t.Run("load valid TOML file", func(t *testing.T) {
		// Create a temporary TOML file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		tomlContent := `
[tools]
node = { version = "20.0.0" }
python = { version = "3.11.0", backend = "aqua" }

[env]
NODE_ENV = "development"
PYTHON_PATH = "/usr/local/bin/python"

[settings]
cache_dir = "/tmp/cache"
data_dir = "/tmp/data"
cache_ttl = 3600
concurrency = 4

[tasks.test]
description = "Run tests"
run = "npm test"
`
		err := os.WriteFile(configPath, []byte(tomlContent), 0644)
		require.NoError(t, err)

		// Load the configuration
		config, err := manager.Load(ctx, configPath)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify Tools
		assert.Len(t, config.Tools, 2)
		assert.Equal(t, "20.0.0", config.Tools["node"].Version)
		assert.Equal(t, "3.11.0", config.Tools["python"].Version)
		assert.Equal(t, "aqua", config.Tools["python"].Backend)

		// Verify Env
		assert.Len(t, config.Env, 2)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", config.Env["node_env"])
		assert.Equal(t, "/usr/local/bin/python", config.Env["python_path"])

		// Verify Settings
		assert.Equal(t, "/tmp/cache", config.Settings.CacheDir)
		assert.Equal(t, "/tmp/data", config.Settings.DataDir)
		assert.Equal(t, 3600, config.Settings.CacheTTL)
		assert.Equal(t, 4, config.Settings.Concurrency)

		// Verify Tasks
		assert.Len(t, config.Tasks, 1)
		assert.Equal(t, "Run tests", config.Tasks["test"].Description)
		assert.Equal(t, "npm test", config.Tasks["test"].Run)
	})

	t.Run("load valid YAML file", func(t *testing.T) {
		// Create a temporary YAML file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
tools:
  node:
    version: "20.0.0"
  python:
    version: "3.11.0"
    backend: "aqua"

env:
  NODE_ENV: "development"
  PYTHON_PATH: "/usr/local/bin/python"

settings:
  cache_dir: "/tmp/cache"
  data_dir: "/tmp/data"
  cache_ttl: 3600
  concurrency: 4

tasks:
  test:
    description: "Run tests"
    run: "npm test"
`
		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		// Load the configuration
		config, err := manager.Load(ctx, configPath)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify Tools
		assert.Len(t, config.Tools, 2)
		assert.Equal(t, "20.0.0", config.Tools["node"].Version)
		assert.Equal(t, "3.11.0", config.Tools["python"].Version)
		assert.Equal(t, "aqua", config.Tools["python"].Backend)

		// Verify Env
		assert.Len(t, config.Env, 2)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", config.Env["node_env"])

		// Verify Settings
		assert.Equal(t, "/tmp/cache", config.Settings.CacheDir)
		assert.Equal(t, 3600, config.Settings.CacheTTL)
	})

	t.Run("file not found", func(t *testing.T) {
		config, err := manager.Load(ctx, "/nonexistent/config.toml")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "configuration file not found")
	})

	t.Run("invalid TOML syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.toml")

		invalidContent := `
[tools
node = { version = "20.0.0"
`
		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		config, err := manager.Load(ctx, configPath)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		invalidContent := `
tools:
  node:
    version: "20.0.0"
  - invalid list item
`
		err := os.WriteFile(configPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		config, err := manager.Load(ctx, configPath)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("empty configuration file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.toml")

		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)

		config, err := manager.Load(ctx, configPath)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Empty config should have empty maps
		assert.Empty(t, config.Tools)
		assert.Empty(t, config.Env)
		assert.Empty(t, config.Tasks)
	})
}

// TestConfigManager_LoadHierarchy tests hierarchical configuration loading
func TestConfigManager_LoadHierarchy(t *testing.T) {
	ctx := context.Background()

	t.Run("load hierarchy with multiple levels", func(t *testing.T) {
		// Create a temporary directory structure
		tmpDir := t.TempDir()

		// Create project config
		projectConfig := filepath.Join(tmpDir, "unirtm.toml")
		projectContent := `
[tools]
node = { version = "18.0.0" }
python = { version = "3.10.0" }

[env]
NODE_ENV = "production"

[settings]
cache_ttl = 7200
`
		err := os.WriteFile(projectConfig, []byte(projectContent), 0644)
		require.NoError(t, err)

		// Create local config (overrides project)
		localConfig := filepath.Join(tmpDir, ".unirtm.local.toml")
		localContent := `
[tools]
node = { version = "20.0.0" }

[env]
NODE_ENV = "development"
DEBUG = "true"

[settings]
concurrency = 8
`
		err = os.WriteFile(localConfig, []byte(localContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Load hierarchy
		manager := newTestConfigManager()
		config, err := manager.LoadHierarchy(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify that local overrides project
		assert.Equal(t, "20.0.0", config.Tools["node"].Version, "local should override project for node")
		assert.Equal(t, "3.10.0", config.Tools["python"].Version, "python should come from project")

		// Verify env merging
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", config.Env["node_env"], "local should override project for node_env")
		assert.Equal(t, "true", config.Env["debug"], "debug should come from local")

		// Verify settings merging
		assert.Equal(t, 7200, config.Settings.CacheTTL, "cache_ttl should come from project")
		assert.Equal(t, 8, config.Settings.Concurrency, "concurrency should come from local")
	})

	t.Run("load hierarchy with no config files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to the temp directory with no config files
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		manager := newTestConfigManager()
		config, err := manager.LoadHierarchy(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Should return empty config
		assert.Empty(t, config.Tools)
		assert.Empty(t, config.Env)
		assert.Empty(t, config.Tasks)
	})

	t.Run("load hierarchy with invalid file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create invalid project config
		projectConfig := filepath.Join(tmpDir, "unirtm.toml")
		invalidContent := `[tools invalid syntax`
		err := os.WriteFile(projectConfig, []byte(invalidContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		manager := newTestConfigManager()
		config, err := manager.LoadHierarchy(ctx)
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})
}

// TestConfigManager_Validate tests configuration validation
func TestConfigManager_Validate(t *testing.T) {
	ctx := context.Background()
	manager := newTestConfigManager()

	t.Run("validate valid configuration", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Env: map[string]string{
				"NODE_ENV": "production",
			},
			Settings: Settings{
				CacheDir:    "/tmp/cache",
				DataDir:     "/tmp/data",
				CacheTTL:    3600,
				Concurrency: 4,
			},
			Tasks: map[string]Task{
				"test": {
					Description: "Run tests",
					Run:         "npm test",
				},
			},
		}

		err := manager.Validate(ctx, config)
		assert.NoError(t, err)
	})

	t.Run("validate nil configuration", func(t *testing.T) {
		err := manager.Validate(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration is nil")
	})

	t.Run("validate configuration with missing tool version", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: ""}, // Missing version
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version is required")
	})

	t.Run("validate configuration with negative cache TTL", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Settings: Settings{
				CacheTTL: -100, // Negative value
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache_ttl must be non-negative")
	})

	t.Run("validate configuration with negative concurrency", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Settings: Settings{
				Concurrency: -4, // Negative value
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "concurrency must be non-negative")
	})

	t.Run("validate configuration with missing task run command", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Tasks: map[string]Task{
				"test": {
					Description: "Run tests",
					Run:         "", // Missing run command
				},
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "run command is required")
	})

	t.Run("validate configuration with circular task dependencies", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Tasks: map[string]Task{
				"task1": {
					Description: "Task 1",
					Run:         "echo task1",
					Depends:     []string{"task2"},
				},
				"task2": {
					Description: "Task 2",
					Run:         "echo task2",
					Depends:     []string{"task1"}, // Circular dependency
				},
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("validate configuration with non-existent task dependency", func(t *testing.T) {
		config := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Tasks: map[string]Task{
				"task1": {
					Description: "Task 1",
					Run:         "echo task1",
					Depends:     []string{"nonexistent"}, // Non-existent task
				},
			},
		}

		err := manager.Validate(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "depends on non-existent task")
	})
}

// TestConfigManager_Merge tests configuration merging
func TestConfigManager_Merge(t *testing.T) {
	manager := newTestConfigManager()

	t.Run("merge two configurations", func(t *testing.T) {
		config1 := &Config{
			Tools: map[string]ToolConfig{
				"node":   {Version: "18.0.0"},
				"python": {Version: "3.10.0"},
			},
			Env: map[string]string{
				"node_env": "production",
				"debug":    "false",
			},
			Settings: Settings{
				CacheDir: "/tmp/cache1",
				CacheTTL: 3600,
			},
			Tasks: map[string]Task{
				"test": {
					Description: "Run tests",
					Run:         "npm test",
				},
			},
		}

		config2 := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"}, // Override node version
				"go":   {Version: "1.21.0"}, // Add new tool
			},
			Env: map[string]string{
				"node_env": "development", // Override node_env
				"path":     "/usr/local/bin",
			},
			Settings: Settings{
				DataDir:     "/tmp/data2",
				Concurrency: 8,
			},
			Tasks: map[string]Task{
				"build": {
					Description: "Build project",
					Run:         "npm run build",
				},
			},
		}

		merged, err := manager.Merge(config1, config2)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Verify Tools merging
		assert.Len(t, merged.Tools, 3)
		assert.Equal(t, "20.0.0", merged.Tools["node"].Version, "node should be overridden")
		assert.Equal(t, "3.10.0", merged.Tools["python"].Version, "python should be preserved")
		assert.Equal(t, "1.21.0", merged.Tools["go"].Version, "go should be added")

		// Verify Env merging
		assert.Len(t, merged.Env, 3)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", merged.Env["node_env"], "node_env should be overridden")
		assert.Equal(t, "false", merged.Env["debug"], "debug should be preserved")
		assert.Equal(t, "/usr/local/bin", merged.Env["path"], "path should be added")

		// Verify Settings merging
		assert.Equal(t, "/tmp/cache1", merged.Settings.CacheDir, "CacheDir from config1")
		assert.Equal(t, "/tmp/data2", merged.Settings.DataDir, "DataDir from config2")
		assert.Equal(t, 3600, merged.Settings.CacheTTL, "CacheTTL from config1")
		assert.Equal(t, 8, merged.Settings.Concurrency, "Concurrency from config2")

		// Verify Tasks merging
		assert.Len(t, merged.Tasks, 2)
		assert.Equal(t, "Run tests", merged.Tasks["test"].Description)
		assert.Equal(t, "Build project", merged.Tasks["build"].Description)
	})

	t.Run("merge with empty configuration", func(t *testing.T) {
		config1 := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
		}

		config2 := &Config{
			Tools: map[string]ToolConfig{},
			Env:   map[string]string{},
			Tasks: map[string]Task{},
		}

		merged, err := manager.Merge(config1, config2)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// config1 values should be preserved
		assert.Len(t, merged.Tools, 1)
		assert.Equal(t, "20.0.0", merged.Tools["node"].Version)
	})

	t.Run("merge with no configurations", func(t *testing.T) {
		merged, err := manager.Merge()
		assert.Error(t, err)
		assert.Nil(t, merged)
		assert.Contains(t, err.Error(), "no configurations to merge")
	})

	t.Run("merge with nil configuration", func(t *testing.T) {
		config1 := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
		}

		merged, err := manager.Merge(config1, nil)
		assert.Error(t, err)
		assert.Nil(t, merged)
		assert.Contains(t, err.Error(), "configuration at index 1 is nil")
	})

	t.Run("merge multiple configurations with precedence", func(t *testing.T) {
		system := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "16.0.0"},
			},
			Settings: Settings{
				CacheTTL: 86400,
			},
		}

		global := &Config{
			Tools: map[string]ToolConfig{
				"node":   {Version: "18.0.0"},
				"python": {Version: "3.10.0"},
			},
			Settings: Settings{
				Concurrency: 4,
			},
		}

		project := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Settings: Settings{
				CacheDir: "/project/cache",
			},
		}

		local := &Config{
			Env: map[string]string{
				"node_env": "development",
			},
		}

		merged, err := manager.Merge(system, global, project, local)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Verify precedence: local > project > global > system
		assert.Equal(t, "20.0.0", merged.Tools["node"].Version, "node from project")
		assert.Equal(t, "3.10.0", merged.Tools["python"].Version, "python from global")
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", merged.Env["node_env"], "node_env from local")
		assert.Equal(t, "/project/cache", merged.Settings.CacheDir, "CacheDir from project")
		assert.Equal(t, 4, merged.Settings.Concurrency, "Concurrency from global")
		assert.Equal(t, 86400, merged.Settings.CacheTTL, "CacheTTL from system")
	})
}

// TestConfigManager_Integration tests end-to-end scenarios
func TestConfigManager_Integration(t *testing.T) {
	ctx := context.Background()
	manager := newTestConfigManager()

	t.Run("load, validate, and merge workflow", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create base config
		baseConfigPath := filepath.Join(tmpDir, "base.toml")
		baseContent := `
[tools]
node = { version = "18.0.0" }

[settings]
cache_ttl = 3600
`
		err := os.WriteFile(baseConfigPath, []byte(baseContent), 0644)
		require.NoError(t, err)

		// Create override config
		overrideConfigPath := filepath.Join(tmpDir, "override.toml")
		overrideContent := `
[tools]
node = { version = "20.0.0" }
python = { version = "3.11.0" }

[settings]
concurrency = 8
`
		err = os.WriteFile(overrideConfigPath, []byte(overrideContent), 0644)
		require.NoError(t, err)

		// Load both configs
		baseConfig, err := manager.Load(ctx, baseConfigPath)
		require.NoError(t, err)

		overrideConfig, err := manager.Load(ctx, overrideConfigPath)
		require.NoError(t, err)

		// Validate both configs
		err = manager.Validate(ctx, baseConfig)
		require.NoError(t, err)

		err = manager.Validate(ctx, overrideConfig)
		require.NoError(t, err)

		// Merge configs
		merged, err := manager.Merge(baseConfig, overrideConfig)
		require.NoError(t, err)

		// Validate merged config
		err = manager.Validate(ctx, merged)
		require.NoError(t, err)

		// Verify merged result
		assert.Equal(t, "20.0.0", merged.Tools["node"].Version)
		assert.Equal(t, "3.11.0", merged.Tools["python"].Version)
		assert.Equal(t, 3600, merged.Settings.CacheTTL)
		assert.Equal(t, 8, merged.Settings.Concurrency)
	})
}

// TestConfigManager_ApplyEnvironment tests environment-specific configuration overrides
func TestConfigManager_ApplyEnvironment(t *testing.T) {
	manager := newTestConfigManager()

	t.Run("apply development environment", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node":   {Version: "20.0.0"},
				"python": {Version: "3.11.0"},
			},
			Env: map[string]string{
				"NODE_ENV": "production",
				"DEBUG":    "false",
			},
			Settings: Settings{
				CacheDir:    "/prod/cache",
				CacheTTL:    7200,
				Concurrency: 4,
			},
			Tasks: map[string]Task{
				"build": {
					Description: "Build for production",
					Run:         "npm run build",
				},
			},
			Environments: map[string]EnvironmentConfig{
				"development": {
					Tools: map[string]ToolConfig{
						"node": {Version: "18.0.0"}, // Override node version
					},
					Env: map[string]string{
						"NODE_ENV": "development", // Override NODE_ENV
						"DEBUG":    "true",        // Override DEBUG
					},
					Settings: Settings{
						CacheDir: "/dev/cache", // Override cache dir
					},
					Tasks: map[string]Task{
						"build": {
							Description: "Build for development",
							Run:         "npm run dev",
						},
					},
				},
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "development")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify tool overrides
		assert.Equal(t, "18.0.0", result.Tools["node"].Version, "node version should be overridden")
		assert.Equal(t, "3.11.0", result.Tools["python"].Version, "python version should be preserved")

		// Verify env overrides
		assert.Equal(t, "development", result.Env["NODE_ENV"], "NODE_ENV should be overridden")
		assert.Equal(t, "true", result.Env["DEBUG"], "DEBUG should be overridden")

		// Verify settings overrides
		assert.Equal(t, "/dev/cache", result.Settings.CacheDir, "CacheDir should be overridden")
		assert.Equal(t, 7200, result.Settings.CacheTTL, "CacheTTL should be preserved")
		assert.Equal(t, 4, result.Settings.Concurrency, "Concurrency should be preserved")

		// Verify task overrides
		assert.Equal(t, "Build for development", result.Tasks["build"].Description)
		assert.Equal(t, "npm run dev", result.Tasks["build"].Run)
	})

	t.Run("apply staging environment", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Settings: Settings{
				CacheTTL: 3600,
			},
			Environments: map[string]EnvironmentConfig{
				"staging": {
					Settings: Settings{
						CacheTTL:    7200,
						Concurrency: 8,
					},
				},
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "staging")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify settings overrides
		assert.Equal(t, 7200, result.Settings.CacheTTL, "CacheTTL should be overridden")
		assert.Equal(t, 8, result.Settings.Concurrency, "Concurrency should be set")

		// Verify tools are preserved
		assert.Equal(t, "20.0.0", result.Tools["node"].Version)
	})

	t.Run("apply production environment", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "18.0.0"},
			},
			Env: map[string]string{
				"NODE_ENV": "development",
			},
			Environments: map[string]EnvironmentConfig{
				"production": {
					Tools: map[string]ToolConfig{
						"node": {Version: "20.0.0"},
					},
					Env: map[string]string{
						"NODE_ENV":     "production",
						"ENABLE_CACHE": "true",
					},
					Settings: Settings{
						CacheTTL:    86400,
						Concurrency: 16,
					},
				},
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "production")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify all overrides
		assert.Equal(t, "20.0.0", result.Tools["node"].Version)
		assert.Equal(t, "production", result.Env["NODE_ENV"])
		assert.Equal(t, "true", result.Env["ENABLE_CACHE"])
		assert.Equal(t, 86400, result.Settings.CacheTTL)
		assert.Equal(t, 16, result.Settings.Concurrency)
	})

	t.Run("environment not found", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Environments: map[string]EnvironmentConfig{
				"development": {
					Tools: map[string]ToolConfig{
						"node": {Version: "18.0.0"},
					},
				},
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "environment \"nonexistent\" not found")
	})

	t.Run("nil configuration", func(t *testing.T) {
		result, err := manager.ApplyEnvironment(nil, "development")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "configuration is nil")
	})

	t.Run("empty environment name", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "environment name is empty")
	})

	t.Run("empty environment config", func(t *testing.T) {
		baseConfig := &Config{
			Tools: map[string]ToolConfig{
				"node": {Version: "20.0.0"},
			},
			Env: map[string]string{
				"NODE_ENV": "production",
			},
			Settings: Settings{
				CacheTTL: 3600,
			},
			Environments: map[string]EnvironmentConfig{
				"test": {}, // Empty environment config
			},
		}

		result, err := manager.ApplyEnvironment(baseConfig, "test")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Base config should be preserved
		assert.Equal(t, "20.0.0", result.Tools["node"].Version)
		assert.Equal(t, "production", result.Env["NODE_ENV"])
		assert.Equal(t, 3600, result.Settings.CacheTTL)
	})
}

// TestConfigManager_LoadWithEnvironment tests loading configuration with environment overrides
func TestConfigManager_LoadWithEnvironment(t *testing.T) {
	ctx := context.Background()

	t.Run("load with environment from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		configContent := `
[tools]
node = { version = "20.0.0" }
python = { version = "3.11.0" }

[env]
NODE_ENV = "production"

[settings]
cache_ttl = 7200

[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }

[environments.development.env]
NODE_ENV = "development"
DEBUG = "true"

[environments.development.settings]
cache_ttl = 3600
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create a config file in the current directory
		err = os.WriteFile("unirtm.toml", []byte(configContent), 0644)
		require.NoError(t, err)

		manager := newTestConfigManager()
		config, err := manager.LoadWithEnvironment(ctx, "development")
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify environment overrides were applied
		assert.Equal(t, "18.0.0", config.Tools["node"].Version, "node should be overridden")
		assert.Equal(t, "3.11.0", config.Tools["python"].Version, "python should be preserved")
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", config.Env["node_env"], "NODE_ENV should be overridden")
		assert.Equal(t, "true", config.Env["debug"], "DEBUG should be added")
		assert.Equal(t, 3600, config.Settings.CacheTTL, "CacheTTL should be overridden")
	})

	t.Run("load without environment", func(t *testing.T) {
		tmpDir := t.TempDir()

		configContent := `
[tools]
node = { version = "20.0.0" }

[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }
`
		err := os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(configContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		manager := newTestConfigManager()
		config, err := manager.LoadWithEnvironment(ctx, "")
		require.NoError(t, err)
		require.NotNil(t, config)

		// Base config should be returned without environment overrides
		assert.Equal(t, "20.0.0", config.Tools["node"].Version)
	})

	t.Run("load with nonexistent environment", func(t *testing.T) {
		tmpDir := t.TempDir()

		configContent := `
[tools]
node = { version = "20.0.0" }

[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }
`
		err := os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(configContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		manager := newTestConfigManager()
		config, err := manager.LoadWithEnvironment(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "environment \"nonexistent\" not found")
	})
}

// TestConfigManager_EnvironmentIntegration tests complete environment workflow
func TestConfigManager_EnvironmentIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("complete environment workflow", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create base config
		baseContent := `
[tools]
node = { version = "20.0.0" }
python = { version = "3.11.0" }
go = { version = "1.21.0" }

[env]
NODE_ENV = "production"
LOG_LEVEL = "info"

[settings]
cache_dir = "/prod/cache"
cache_ttl = 86400
concurrency = 16

[tasks.build]
description = "Build for production"
run = "npm run build"

[tasks.test]
description = "Run tests"
run = "npm test"

[environments.development]
[environments.development.tools]
node = { version = "18.0.0" }

[environments.development.env]
NODE_ENV = "development"
LOG_LEVEL = "debug"
DEBUG = "true"

[environments.development.settings]
cache_dir = "/dev/cache"
cache_ttl = 3600
concurrency = 4

[environments.development.tasks.build]
description = "Build for development"
run = "npm run dev"

[environments.staging]
[environments.staging.env]
NODE_ENV = "staging"

[environments.staging.settings]
cache_ttl = 43200
concurrency = 8

[environments.production]
[environments.production.tools]
node = { version = "20.0.0" }
go = { version = "1.22.0" }

[environments.production.env]
NODE_ENV = "production"
LOG_LEVEL = "warn"
ENABLE_MONITORING = "true"

[environments.production.settings]
cache_ttl = 172800
concurrency = 32
`
		err := os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(baseContent), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		manager := newTestConfigManager()

		// Test development environment
		devConfig, err := manager.LoadWithEnvironment(ctx, "development")
		require.NoError(t, err)
		require.NotNil(t, devConfig)

		assert.Equal(t, "18.0.0", devConfig.Tools["node"].Version)
		assert.Equal(t, "3.11.0", devConfig.Tools["python"].Version)
		assert.Equal(t, "1.21.0", devConfig.Tools["go"].Version)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "development", devConfig.Env["node_env"])
		assert.Equal(t, "debug", devConfig.Env["log_level"])
		assert.Equal(t, "true", devConfig.Env["debug"])
		assert.Equal(t, "/dev/cache", devConfig.Settings.CacheDir)
		assert.Equal(t, 3600, devConfig.Settings.CacheTTL)
		assert.Equal(t, 4, devConfig.Settings.Concurrency)
		assert.Equal(t, "Build for development", devConfig.Tasks["build"].Description)
		assert.Equal(t, "npm run dev", devConfig.Tasks["build"].Run)
		assert.Equal(t, "Run tests", devConfig.Tasks["test"].Description)

		// Test staging environment
		stagingConfig, err := manager.LoadWithEnvironment(ctx, "staging")
		require.NoError(t, err)
		require.NotNil(t, stagingConfig)

		assert.Equal(t, "20.0.0", stagingConfig.Tools["node"].Version)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "staging", stagingConfig.Env["node_env"])
		assert.Equal(t, "info", stagingConfig.Env["log_level"])
		assert.Equal(t, "/prod/cache", stagingConfig.Settings.CacheDir)
		assert.Equal(t, 43200, stagingConfig.Settings.CacheTTL)
		assert.Equal(t, 8, stagingConfig.Settings.Concurrency)

		// Test production environment
		prodConfig, err := manager.LoadWithEnvironment(ctx, "production")
		require.NoError(t, err)
		require.NotNil(t, prodConfig)

		assert.Equal(t, "20.0.0", prodConfig.Tools["node"].Version)
		assert.Equal(t, "3.11.0", prodConfig.Tools["python"].Version)
		assert.Equal(t, "1.22.0", prodConfig.Tools["go"].Version)
		// Note: Viper lowercases all keys by default
		assert.Equal(t, "production", prodConfig.Env["node_env"])
		assert.Equal(t, "warn", prodConfig.Env["log_level"])
		assert.Equal(t, "true", prodConfig.Env["enable_monitoring"])
		assert.Equal(t, "/prod/cache", prodConfig.Settings.CacheDir)
		assert.Equal(t, 172800, prodConfig.Settings.CacheTTL)
		assert.Equal(t, 32, prodConfig.Settings.Concurrency)

		// Validate all configurations
		err = manager.Validate(ctx, devConfig)
		require.NoError(t, err)

		err = manager.Validate(ctx, stagingConfig)
		require.NoError(t, err)

		err = manager.Validate(ctx, prodConfig)
		require.NoError(t, err)
	})
}
