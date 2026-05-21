// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: Config{
				Tools: map[string]ToolConfig{
					"node":   {Version: "20.0.0"},
					"python": {Version: "3.11.0", Backend: "github"},
				},
				Env: map[string]interface{}{
					"NODE_ENV": "development",
				},
				Settings: Settings{
					CacheDir: "/tmp/cache",
					DataDir:  "/tmp/data",
					CacheTTL: 3600,
				},
				Tasks: map[string]Task{
					"build": {
						Description: "Build the project",
						Run:         "go build",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with minimal fields",
			config: Config{
				Tools: map[string]ToolConfig{
					"node": {Version: "20.0.0"},
				},
				Settings: Settings{},
			},
			wantErr: false,
		},
		{
			name: "empty config is valid",
			config: Config{
				Settings: Settings{},
			},
			wantErr: false,
		},
		{
			name: "invalid tool config - missing version",
			config: Config{
				Tools: map[string]ToolConfig{
					"node": {Version: ""},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid settings - negative cache TTL",
			config: Config{
				Settings: Settings{
					CacheTTL: -100,
				},
			},
			wantErr: true,
			errMsg:  "cache_ttl must be non-negative",
		},

		{
			name: "invalid task - missing run command",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Description: "Build the project",
						Run:         "",
					},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "run command is required",
		},
		{
			name: "invalid task dependency - non-existent task",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     "go build",
						Depends: []string{"test"},
					},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "depends on non-existent task",
		},
		{
			name: "invalid task dependency - circular dependency",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     "go build",
						Depends: []string{"test"},
					},
					"test": {
						Run:     "go test",
						Depends: []string{"build"},
					},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "multiple validation errors",
			config: Config{
				Tools: map[string]ToolConfig{
					"node":   {Version: ""},
					"python": {Version: "3.11.0"},
				},
				Settings: Settings{
					CacheTTL: -100,
				},
				Tasks: map[string]Task{
					"build": {Run: ""},
				},
			},
			wantErr: true,
			errMsg:  "configuration validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestToolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ToolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with exact version",
			config: ToolConfig{
				Version: "1.20.0",
			},
			wantErr: false,
		},
		{
			name: "valid with version range",
			config: ToolConfig{
				Version: ">=1.20.0",
			},
			wantErr: false,
		},
		{
			name: "valid with alias",
			config: ToolConfig{
				Version: "latest",
			},
			wantErr: false,
		},
		{
			name: "valid with backend and provider",
			config: ToolConfig{
				Version:  "1.20.0",
				Backend:  "github",
				Provider: "node",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty version",
			config: ToolConfig{
				Version: "",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid - whitespace only version",
			config: ToolConfig{
				Version: "   ",
			},
			wantErr: false, // Whitespace is technically a non-empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSettings_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Settings
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with all fields",
			config: Settings{
				CacheDir: "/tmp/cache",
				DataDir:  "/tmp/data",
				CacheTTL: 3600,
			},
			wantErr: false,
		},
		{
			name: "valid with zero values",
			config: Settings{
				CacheTTL: 0,
			},
			wantErr: false,
		},
		{
			name:    "valid empty settings",
			config:  Settings{},
			wantErr: false,
		},
		{
			name: "invalid - negative cache TTL",
			config: Settings{
				CacheTTL: -100,
			},
			wantErr: true,
			errMsg:  "cache_ttl must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with all fields",
			task: Task{
				Description: "Build the project",
				Run:         "go build",
				Env: map[string]interface{}{
					"CGO_ENABLED": "0",
				},
				Depends: []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "valid with minimal fields",
			task: Task{
				Run: "go build",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty run command",
			task: Task{
				Description: "Build the project",
				Run:         "",
			},
			wantErr: true,
			errMsg:  "run command is required",
		},
		{
			name: "invalid - whitespace only run command",
			task: Task{
				Run: "   ",
			},
			wantErr: false, // Whitespace is technically a non-empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateTaskDependencies(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid - no dependencies",
			config: Config{
				Tasks: map[string]Task{
					"build": {Run: "go build"},
					"test":  {Run: "go test"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - linear dependencies",
			config: Config{
				Tasks: map[string]Task{
					"build": {Run: "go build"},
					"test": {
						Run:     "go test",
						Depends: []string{"build"},
					},
					"deploy": {
						Run:     "deploy.sh",
						Depends: []string{"test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - multiple dependencies",
			config: Config{
				Tasks: map[string]Task{
					"lint":  {Run: "golangci-lint run"},
					"test":  {Run: "go test"},
					"build": {Run: "go build"},
					"deploy": {
						Run:     "deploy.sh",
						Depends: []string{"lint", "test", "build"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - non-existent dependency",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     "go build",
						Depends: []string{"test"},
					},
				},
			},
			wantErr: true,
			errMsg:  "depends on non-existent task",
		},
		{
			name: "invalid - direct circular dependency",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     "go build",
						Depends: []string{"test"},
					},
					"test": {
						Run:     "go test",
						Depends: []string{"build"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "invalid - indirect circular dependency",
			config: Config{
				Tasks: map[string]Task{
					"a": {
						Run:     "task a",
						Depends: []string{"b"},
					},
					"b": {
						Run:     "task b",
						Depends: []string{"c"},
					},
					"c": {
						Run:     "task c",
						Depends: []string{"a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
		{
			name: "invalid - self-dependency",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     "go build",
						Depends: []string{"build"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateTaskDependencies()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestConfig_ValidateInitializesNilMaps tests that Validate initializes nil maps
func TestConfig_ValidateInitializesNilMaps(t *testing.T) {
	config := Config{
		Settings: Settings{},
	}

	err := config.Validate()
	require.NoError(t, err)

	assert.NotNil(t, config.Tools)
	assert.NotNil(t, config.Tasks)
	assert.Len(t, config.Tools, 0)
	assert.Len(t, config.Tasks, 0)
}

func TestEnvironmentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  EnvironmentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid environment config with all fields",
			config: EnvironmentConfig{
				Tools: map[string]ToolConfig{
					"node": {Version: "18.0.0"},
				},
				Env: map[string]interface{}{
					"NODE_ENV": "production",
				},
				Settings: Settings{
					CacheTTL: 7200,
				},
				Tasks: map[string]Task{
					"build": {Run: "npm run build"},
				},
			},
			wantErr: false,
		},
		{
			name:    "valid empty environment config",
			config:  EnvironmentConfig{},
			wantErr: false,
		},
		{
			name: "invalid tool in environment config",
			config: EnvironmentConfig{
				Tools: map[string]ToolConfig{
					"node": {Version: ""},
				},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid settings in environment config",
			config: EnvironmentConfig{
				Settings: Settings{
					CacheTTL: -100,
				},
			},
			wantErr: true,
			errMsg:  "cache_ttl must be non-negative",
		},
		{
			name: "invalid task in environment config",
			config: EnvironmentConfig{
				Tasks: map[string]Task{
					"build": {Run: ""},
				},
			},
			wantErr: true,
			errMsg:  "run command is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateWithEnvironments(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with environments",
			config: Config{
				Tools: map[string]ToolConfig{
					"node": {Version: "20.0.0"},
				},
				Environments: map[string]EnvironmentConfig{
					"development": {
						Tools: map[string]ToolConfig{
							"node": {Version: "18.0.0"},
						},
					},
					"production": {
						Settings: Settings{
							CacheTTL: 7200,
						},
					},
				},
				Settings: Settings{},
			},
			wantErr: false,
		},
		{
			name: "invalid environment config",
			config: Config{
				Tools: map[string]ToolConfig{
					"node": {Version: "20.0.0"},
				},
				Environments: map[string]EnvironmentConfig{
					"development": {
						Tools: map[string]ToolConfig{
							"node": {Version: ""},
						},
					},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "environment \"development\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
