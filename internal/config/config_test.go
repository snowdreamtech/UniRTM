// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

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
						Run:         StringArray{"go build"},
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
						Run:         StringArray{""},
					},
				},
				Settings: Settings{},
			},
			wantErr: true,
			errMsg:  "run command or depends is required",
		},
		{
			name: "invalid task dependency - non-existent task",
			config: Config{
				Tasks: map[string]Task{
					"build": {
						Run:     StringArray{"go build"},
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
						Run:     StringArray{"go build"},
						Depends: []string{"test"},
					},
					"test": {
						Run:     StringArray{"go test"},
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
					"build": {Run: StringArray{""}},
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
		{
			name: "valid with tool alias",
			config: ToolConfig{
				Version: "3.14.1",
				Alias:   "shfmt",
			},
			wantErr: false,
		},
		{
			name: "invalid - tool alias with path separator",
			config: ToolConfig{
				Version: "3.14.1",
				Alias:   "foo/bar",
			},
			wantErr: true,
			errMsg:  "alias cannot contain path separators or colons",
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

func TestToolConfig_ParseAlias(t *testing.T) {
	tomlContent := `
[tools]
"github:mvdan/sh" = { version = "3.14.1", alias = "shfmt" }
`
	cfg := &Config{}
	err := toml.Unmarshal([]byte(tomlContent), cfg)
	require.NoError(t, err)
	cfg.PostLoad()
	require.NotNil(t, cfg.Tools)
	tc, ok := cfg.Tools["github:mvdan/sh"]
	require.True(t, ok)
	assert.Equal(t, "3.14.1", tc.Version)
	assert.Equal(t, "shfmt", tc.Alias)
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
				Run:         StringArray{"go build"},
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
				Run: StringArray{"go build"},
			},
			wantErr: false,
		},
		{
			name: "valid with only depends",
			task: Task{
				Description: "Run full local verification",
				Depends:     []string{"lint-all", "test", "audit"},
			},
			wantErr: false,
		},
		{
			name: "invalid - empty run command",
			task: Task{
				Description: "Build the project",
				Run:         StringArray{""},
			},
			wantErr: true,
			errMsg:  "run command or depends is required",
		},
		{
			name: "invalid - whitespace only run command",
			task: Task{
				Run: StringArray{"   "},
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
					"build": {Run: StringArray{"go build"}},
					"test":  {Run: StringArray{"go test"}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid - linear dependencies",
			config: Config{
				Tasks: map[string]Task{
					"build": {Run: StringArray{"go build"}},
					"test": {
						Run:     StringArray{"go test"},
						Depends: []string{"build"},
					},
					"deploy": {
						Run:     StringArray{"deploy.sh"},
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
					"lint":  {Run: StringArray{"golangci-lint run"}},
					"test":  {Run: StringArray{"go test"}},
					"build": {Run: StringArray{"go build"}},
					"deploy": {
						Run:     StringArray{"deploy.sh"},
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
						Run:     StringArray{"go build"},
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
						Run:     StringArray{"go build"},
						Depends: []string{"test"},
					},
					"test": {
						Run:     StringArray{"go test"},
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
						Run:     StringArray{"task a"},
						Depends: []string{"b"},
					},
					"b": {
						Run:     StringArray{"task b"},
						Depends: []string{"c"},
					},
					"c": {
						Run:     StringArray{"task c"},
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
						Run:     StringArray{"go build"},
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

func TestConfig_MergeOverlay(t *testing.T) {
	base := &Config{
		Tools: map[string]ToolConfig{
			"node": {Version: "18.0.0"},
			"go":   {Version: "1.21.0"},
		},
		Env: map[string]interface{}{
			"APP_ENV": "development",
			"PORT":    8080,
		},
		Tasks: map[string]Task{
			"build": {Run: StringArray{"go build"}},
		},
		Settings: Settings{
			CacheTTL: 3600,
			Color:    "auto",
		},
	}

	overlay := &Config{
		Tools: map[string]ToolConfig{
			"node": {Version: "20.0.0"},
			"rust": {Version: "1.75.0"},
		},
		Env: map[string]interface{}{
			"APP_ENV": "production",
			"DEBUG":   "true",
		},
		Tasks: map[string]Task{
			"build": {Run: StringArray{"cargo build --release"}},
			"test":  {Run: StringArray{"cargo test"}},
		},
		Settings: Settings{
			CacheTTL: 7200,
		},
	}

	base.MergeOverlay(overlay)

	// Tools
	assert.Equal(t, "20.0.0", base.Tools["node"].Version)
	assert.Equal(t, "1.21.0", base.Tools["go"].Version)
	assert.Equal(t, "1.75.0", base.Tools["rust"].Version)

	// Env
	assert.Equal(t, "production", base.Env["APP_ENV"])
	assert.Equal(t, 8080, base.Env["PORT"])
	assert.Equal(t, "true", base.Env["DEBUG"])

	// Tasks
	assert.Equal(t, StringArray{"cargo build --release"}, base.Tasks["build"].Run)
	assert.Equal(t, StringArray{"cargo test"}, base.Tasks["test"].Run)

	// Settings
	assert.Equal(t, DurationOrInt(7200), base.Settings.CacheTTL)
	assert.Equal(t, "auto", base.Settings.Color)
}

func TestDurationOrInt(t *testing.T) {
	var d DurationOrInt

	// Test UnmarshalText
	err := d.UnmarshalText([]byte("42"))
	if err != nil || d != 42 {
		t.Errorf("UnmarshalText 42 failed: %v, %v", err, d)
	}

	err = d.UnmarshalText([]byte("24h"))
	if err != nil || d != 86400 {
		t.Errorf("UnmarshalText 24h failed: %v, %v", err, d)
	}

	err = d.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Errorf("UnmarshalText invalid should fail")
	}

	// Test UnmarshalJSON
	err = d.UnmarshalJSON([]byte(`42`))
	if err != nil || d != 42 {
		t.Errorf("UnmarshalJSON 42 failed: %v, %v", err, d)
	}

	err = d.UnmarshalJSON([]byte(`"24h"`))
	if err != nil || d != 86400 {
		t.Errorf("UnmarshalJSON 24h failed: %v, %v", err, d)
	}

	err = d.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Errorf("UnmarshalJSON invalid should fail")
	}

	// Test UnmarshalYAML
	type Config struct {
		Val DurationOrInt `yaml:"val"`
	}
	var c Config
	err = yaml.Unmarshal([]byte("val: 42"), &c)
	if err != nil || c.Val != 42 {
		t.Errorf("UnmarshalYAML 42 failed: %v, %v", err, c.Val)
	}

	err = yaml.Unmarshal([]byte("val: 24h"), &c)
	if err != nil || c.Val != 86400 {
		t.Errorf("UnmarshalYAML 24h failed: %v, %v", err, c.Val)
	}

	err = yaml.Unmarshal([]byte("val: invalid"), &c)
	if err == nil {
		t.Errorf("UnmarshalYAML invalid should fail")
	}
}
