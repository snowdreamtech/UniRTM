// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActivationManager(t *testing.T) {
	shimsDir := "/usr/local/unirtm/shims"
	dataDir := "/var/lib/unirtm"

	manager := NewActivationManager(shimsDir, dataDir, provider.NewRegistry())

	require.NotNil(t, manager)
	assert.Equal(t, shimsDir, manager.shimsDir)
	assert.Equal(t, dataDir, manager.dataDir)
}

func TestActivationManager_GenerateActivationScript_Bash(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:    ShellBash,
		Scope:    ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"node":   "20.0.0",
			"python": "3.11.0",
		},
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellBash, script.Shell)
	assert.NotEmpty(t, script.Content)
	assert.NotEmpty(t, script.Instructions)

	// Verify script content
	assert.Contains(t, script.Content, "export PATH=\"/usr/local/unirtm/shims:$PATH\"")
	assert.Contains(t, script.Content, "export UNIRTM_NODE_VERSION=\"20.0.0\"")
	assert.Contains(t, script.Content, "export UNIRTM_PYTHON_VERSION=\"3.11.0\"")
	assert.Contains(t, script.Content, "export NODE_ENV=\"production\"")
	assert.Contains(t, script.Content, "export UNIRTM_ACTIVATION_SCOPE=\"global\"")
}

func TestActivationManager_GenerateActivationScript_Zsh(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:      ShellZsh,
		Scope:      ScopeProject,
		ShimsDir:   "/usr/local/unirtm/shims",
		ProjectDir: "/home/user/myproject",
		ToolVersions: map[string]string{
			"go": "1.21.0",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellZsh, script.Shell)
	assert.NotEmpty(t, script.Content)

	// Verify script content
	assert.Contains(t, script.Content, "export PATH=\"/usr/local/unirtm/shims:$PATH\"")
	assert.Contains(t, script.Content, "export UNIRTM_GO_VERSION=\"1.21.0\"")
	assert.Contains(t, script.Content, "export UNIRTM_ACTIVATION_SCOPE=\"project\"")
	assert.Contains(t, script.Content, "export UNIRTM_PROJECT_DIR=\"/home/user/myproject\"")
}

func TestActivationManager_GenerateActivationScript_Fish(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:    ShellFish,
		Scope:    ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"ruby": "3.2.0",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellFish, script.Shell)
	assert.NotEmpty(t, script.Content)

	// Verify script content (fish uses different syntax)
	assert.Contains(t, script.Content, "set -gx PATH \"/usr/local/unirtm/shims\" $PATH")
	assert.Contains(t, script.Content, "set -gx UNIRTM_RUBY_VERSION \"3.2.0\"")
	assert.Contains(t, script.Content, "set -gx UNIRTM_ACTIVATION_SCOPE \"global\"")
}

func TestActivationManager_GenerateActivationScript_PowerShell(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:    ShellPowerShell,
		Scope:    ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{
			"dotnet": "8.0.0",
		},
		EnvVars: map[string]string{
			"DOTNET_CLI_TELEMETRY_OPTOUT": "1",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellPowerShell, script.Shell)
	assert.NotEmpty(t, script.Content)

	// Verify script content (PowerShell uses different syntax)
	assert.Contains(t, script.Content, "$env:PATH =")
	assert.Contains(t, script.Content, "$env:UNIRTM_DOTNET_VERSION = \"8.0.0\"")
	assert.Contains(t, script.Content, "$env:DOTNET_CLI_TELEMETRY_OPTOUT = \"1\"")
	assert.Contains(t, script.Content, "$env:UNIRTM_ACTIVATION_SCOPE = \"global\"")
}

func TestActivationManager_GenerateActivationScript_InvalidConfig(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	tests := []struct {
		name        string
		config      ActivationConfig
		expectedErr string
	}{
		{
			name: "missing shell",
			config: ActivationConfig{
				Scope:    ScopeGlobal,
				ShimsDir: "/usr/local/unirtm/shims",
			},
			expectedErr: "shell type is required",
		},
		{
			name: "project scope without project directory",
			config: ActivationConfig{
				Shell:    ShellBash,
				Scope:    ScopeProject,
				ShimsDir: "/usr/local/unirtm/shims",
			},
			expectedErr: "project directory is required",
		},
		{
			name: "unsupported shell",
			config: ActivationConfig{
				Shell:    ShellType("unsupported"),
				Scope:    ScopeGlobal,
				ShimsDir: "/usr/local/unirtm/shims",
			},
			expectedErr: "unsupported shell type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := manager.GenerateActivationScript(ctx, tt.config)

			require.Error(t, err)
			assert.Nil(t, script)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestActivationManager_GenerateGlobalActivation(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	toolVersions := map[string]string{
		"node":   "20.0.0",
		"python": "3.11.0",
		"go":     "1.21.0",
	}

	script, err := manager.GenerateGlobalActivation(ctx, ShellBash, toolVersions)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellBash, script.Shell)
	assert.Contains(t, script.Content, "export UNIRTM_ACTIVATION_SCOPE=\"global\"")
	assert.Contains(t, script.Content, "export UNIRTM_NODE_VERSION=\"20.0.0\"")
	assert.Contains(t, script.Content, "export UNIRTM_PYTHON_VERSION=\"3.11.0\"")
	assert.Contains(t, script.Content, "export UNIRTM_GO_VERSION=\"1.21.0\"")
}

func TestActivationManager_GenerateProjectActivation(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	toolVersions := map[string]string{
		"node": "18.0.0",
	}

	envVars := map[string]string{
		"NODE_ENV": "development",
		"DEBUG":    "app:*",
	}

	script, err := manager.GenerateProjectActivation(ctx, ShellBash, "/home/user/myproject", toolVersions, envVars)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Equal(t, ShellBash, script.Shell)
	assert.Contains(t, script.Content, "export UNIRTM_ACTIVATION_SCOPE=\"project\"")
	assert.Contains(t, script.Content, "export UNIRTM_PROJECT_DIR=\"/home/user/myproject\"")
	assert.Contains(t, script.Content, "export UNIRTM_NODE_VERSION=\"18.0.0\"")
	assert.Contains(t, script.Content, "export NODE_ENV=\"development\"")
	assert.Contains(t, script.Content, "export DEBUG=\"app:*\"")
}

func TestActivationManager_ToolVersionEnvVar(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())

	tests := []struct {
		tool     string
		expected string
	}{
		{"node", "UNIRTM_NODE_VERSION"},
		{"python", "UNIRTM_PYTHON_VERSION"},
		{"go", "UNIRTM_GO_VERSION"},
		{"ruby", "UNIRTM_RUBY_VERSION"},
		{"node-js", "UNIRTM_NODE_JS_VERSION"},
		{"my-tool", "UNIRTM_MY_TOOL_VERSION"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			envVar := manager.toolVersionEnvVar(tt.tool)
			assert.Equal(t, tt.expected, envVar)
		})
	}
}

func TestActivationManager_GenerateActivationScript_EmptyToolVersions(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:        ShellBash,
		Scope:        ScopeGlobal,
		ShimsDir:     "/usr/local/unirtm/shims",
		ToolVersions: map[string]string{},
		EnvVars:      map[string]string{},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Contains(t, script.Content, "export PATH=\"/usr/local/unirtm/shims:$PATH\"")
	assert.Contains(t, script.Content, "export UNIRTM_ACTIVATION_SCOPE=\"global\"")
	// Should not contain tool version or env var sections
	assert.NotContains(t, script.Content, "Set active tool versions")
	assert.NotContains(t, script.Content, "Set additional environment variables")
}

func TestActivationManager_GenerateActivationScript_MultipleTools(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	toolVersions := map[string]string{
		"node":   "20.0.0",
		"python": "3.11.0",
		"go":     "1.21.0",
		"ruby":   "3.2.0",
		"java":   "17.0.0",
		"rust":   "1.70.0",
		"dotnet": "8.0.0",
		"php":    "8.2.0",
	}

	config := ActivationConfig{
		Shell:        ShellBash,
		Scope:        ScopeGlobal,
		ShimsDir:     "/usr/local/unirtm/shims",
		ToolVersions: toolVersions,
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)

	// Verify all tools are included
	for tool, version := range toolVersions {
		envVar := manager.toolVersionEnvVar(tool)
		assert.Contains(t, script.Content, fmt.Sprintf("export %s=\"%s\"", envVar, version))
	}
}

func TestActivationManager_GenerateActivationScript_PathModification(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	tests := []struct {
		name     string
		shell    ShellType
		shimsDir string
		expected string
	}{
		{
			name:     "bash with standard path",
			shell:    ShellBash,
			shimsDir: "/usr/local/unirtm/shims",
			expected: "export PATH=\"/usr/local/unirtm/shims:$PATH\"",
		},
		{
			name:     "zsh with custom path",
			shell:    ShellZsh,
			shimsDir: "/home/user/.unirtm/shims",
			expected: "export PATH=\"/home/user/.unirtm/shims:$PATH\"",
		},
		{
			name:     "fish with standard path",
			shell:    ShellFish,
			shimsDir: "/usr/local/unirtm/shims",
			expected: "set -gx PATH \"/usr/local/unirtm/shims\" $PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ActivationConfig{
				Shell:    tt.shell,
				Scope:    ScopeGlobal,
				ShimsDir: tt.shimsDir,
			}

			script, err := manager.GenerateActivationScript(ctx, config)

			require.NoError(t, err)
			require.NotNil(t, script)
			assert.Contains(t, script.Content, tt.expected)
		})
	}
}

func TestActivationManager_GenerateActivationScript_Instructions(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	tests := []struct {
		name     string
		shell    ShellType
		expected []string
	}{
		{
			name:  "bash instructions",
			shell: ShellBash,
			expected: []string{
				"source /path/to/activation.sh",
				"~/.bashrc",
			},
		},
		{
			name:  "zsh instructions",
			shell: ShellZsh,
			expected: []string{
				"source /path/to/activation.sh",
				"~/.zshrc",
			},
		},
		{
			name:  "fish instructions",
			shell: ShellFish,
			expected: []string{
				"source /path/to/activation.fish",
				"~/.config/fish/conf.d/unirtm.fish",
			},
		},
		{
			name:  "powershell instructions",
			shell: ShellPowerShell,
			expected: []string{
				". \\path\\to\\activation.ps1",
				"$PROFILE\\unirtm.ps1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ActivationConfig{
				Shell:    tt.shell,
				Scope:    ScopeGlobal,
				ShimsDir: "/usr/local/unirtm/shims",
			}

			script, err := manager.GenerateActivationScript(ctx, config)

			require.NoError(t, err)
			require.NotNil(t, script)
			assert.NotEmpty(t, script.Instructions)

			for _, expected := range tt.expected {
				assert.Contains(t, script.Instructions, expected)
			}
		})
	}
}

func TestActivationManager_GenerateActivationScript_DefaultShimsDir(t *testing.T) {
	defaultShimsDir := "/usr/local/unirtm/shims"
	manager := NewActivationManager(defaultShimsDir, "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell: ShellBash,
		Scope: ScopeGlobal,
		// ShimsDir not specified - should use default
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)
	assert.Contains(t, script.Content, fmt.Sprintf("export PATH=\"%s:$PATH\"", defaultShimsDir))
}

func TestDetectShell(t *testing.T) {
	shell, err := DetectShell()

	require.NoError(t, err)

	// On Windows, should detect PowerShell
	if runtime.GOOS == "windows" {
		assert.Equal(t, ShellPowerShell, shell)
	} else {
		// On Unix-like systems, should detect a POSIX shell or default to bash
		assert.Contains(t, []ShellType{ShellBash, ShellZsh, ShellFish}, shell)
	}
}

func TestActivationManager_GenerateActivationScript_SpecialCharacters(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:    ShellBash,
		Scope:    ScopeGlobal,
		ShimsDir: "/usr/local/unirtm/shims",
		EnvVars: map[string]string{
			"MY_VAR": "value with spaces",
			"QUOTED": "value\"with\"quotes",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)

	// Values should be properly quoted
	assert.Contains(t, script.Content, "export MY_VAR=\"value with spaces\"")
	assert.Contains(t, script.Content, "export QUOTED=\"value\"with\"quotes\"")
}

func TestActivationManager_GenerateActivationScript_ProjectScope(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	tests := []struct {
		name       string
		shell      ShellType
		projectDir string
	}{
		{
			name:       "bash project activation",
			shell:      ShellBash,
			projectDir: "/home/user/myproject",
		},
		{
			name:       "zsh project activation",
			shell:      ShellZsh,
			projectDir: "/home/user/another-project",
		},
		{
			name:       "fish project activation",
			shell:      ShellFish,
			projectDir: "/home/user/fish-project",
		},
		{
			name:       "powershell project activation",
			shell:      ShellPowerShell,
			projectDir: "C:\\Users\\user\\myproject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ActivationConfig{
				Shell:      tt.shell,
				Scope:      ScopeProject,
				ShimsDir:   "/usr/local/unirtm/shims",
				ProjectDir: tt.projectDir,
				ToolVersions: map[string]string{
					"node": "20.0.0",
				},
			}

			script, err := manager.GenerateActivationScript(ctx, config)

			require.NoError(t, err)
			require.NotNil(t, script)

			// Verify project-specific markers
			content := script.Content
			assert.Contains(t, content, "project")
			assert.Contains(t, content, "UNIRTM_PROJECT_DIR")
		})
	}
}

func TestActivationManager_GenerateActivationScript_Comments(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	config := ActivationConfig{
		Shell:      ShellBash,
		Scope:      ScopeProject,
		ShimsDir:   "/usr/local/unirtm/shims",
		ProjectDir: "/home/user/myproject",
		ToolVersions: map[string]string{
			"node": "20.0.0",
		},
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
	}

	script, err := manager.GenerateActivationScript(ctx, config)

	require.NoError(t, err)
	require.NotNil(t, script)

	// Verify comments are present
	assert.Contains(t, script.Content, "# UniRTM activation script")
	assert.Contains(t, script.Content, "# Shell: bash")
	assert.Contains(t, script.Content, "# Scope: project")
	assert.Contains(t, script.Content, "# Project: /home/user/myproject")
	assert.Contains(t, script.Content, "# Add UniRTM shims to PATH")
	assert.Contains(t, script.Content, "# Set active tool versions")
	assert.Contains(t, script.Content, "# Set additional environment variables")
}

func TestActivationManager_GenerateActivationScript_AllShells(t *testing.T) {
	manager := NewActivationManager("/usr/local/unirtm/shims", "/var/lib/unirtm", provider.NewRegistry())
	ctx := context.Background()

	shells := []ShellType{ShellBash, ShellZsh, ShellFish, ShellPowerShell}

	for _, shell := range shells {
		t.Run(string(shell), func(t *testing.T) {
			config := ActivationConfig{
				Shell:    shell,
				Scope:    ScopeGlobal,
				ShimsDir: "/usr/local/unirtm/shims",
				ToolVersions: map[string]string{
					"node": "20.0.0",
				},
			}

			script, err := manager.GenerateActivationScript(ctx, config)

			require.NoError(t, err)
			require.NotNil(t, script)
			assert.Equal(t, shell, script.Shell)
			assert.NotEmpty(t, script.Content)
			assert.NotEmpty(t, script.Instructions)

			// All scripts should set the activation scope
			assert.True(t, strings.Contains(script.Content, "UNIRTM_ACTIVATION_SCOPE") ||
				strings.Contains(script.Content, "$env:UNIRTM_ACTIVATION_SCOPE"))

			// All scripts should set the tool version
			assert.True(t, strings.Contains(script.Content, "UNIRTM_NODE_VERSION") ||
				strings.Contains(script.Content, "$env:UNIRTM_NODE_VERSION"))
		})
	}
}
