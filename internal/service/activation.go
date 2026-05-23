// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides high-level business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

// ShellType represents the type of shell for activation scripts.
type ShellType string

const (
	// ShellBash represents bash shell
	ShellBash ShellType = "bash"
	// ShellZsh represents zsh shell
	ShellZsh ShellType = "zsh"
	// ShellFish represents fish shell
	ShellFish ShellType = "fish"
	// ShellPowerShell represents PowerShell
	ShellPowerShell ShellType = "powershell"
)

// ActivationScope represents the scope of activation.
type ActivationScope string

const (
	// ScopeGlobal represents global (system-wide) activation
	ScopeGlobal ActivationScope = "global"
	// ScopeProject represents project-specific activation
	ScopeProject ActivationScope = "project"
)

// ActivationConfig contains configuration for activation.
type ActivationConfig struct {
	// Shell is the target shell type
	Shell ShellType
	// Scope is the activation scope (global or project)
	Scope ActivationScope
	// ShimsDir is the directory containing shim scripts
	ShimsDir string
	// ProjectDir is the project directory (for project-specific activation)
	ProjectDir string
	// ToolVersions maps tool names to their active versions
	ToolVersions map[string]string
	// EnvVars contains additional environment variables to set
	EnvVars map[string]string
	// Sources contains shell scripts to source
	// ExePath is the absolute path to the UniRTM executable
	ExePath string
	Sources []string
	// UseShims indicates whether to use shims (true) or PATH mode (false)
	UseShims bool
	// InjectedPaths is the list of absolute paths to inject into PATH (if not using shims)
	InjectedPaths []string
}

// ActivationScript represents a generated activation script.
type ActivationScript struct {
	// Shell is the target shell type
	Shell ShellType
	// Content is the script content
	Content string
	// Instructions are human-readable instructions for using the script
	Instructions string
}

// ActivationManager manages environment activation for tools.
type ActivationManager struct {
	// shimsDir is the default directory for shim scripts
	shimsDir string
	// dataDir is the directory for storing activation state
	dataDir string
	// registry is the provider registry for tool discovery
	registry *provider.Registry
}

// NewActivationManager creates a new ActivationManager.
func NewActivationManager(shimsDir, dataDir string, registry *provider.Registry) *ActivationManager {
	return &ActivationManager{
		shimsDir: shimsDir,
		dataDir:  dataDir,
		registry: registry,
	}
}

// GenerateActivationScript generates a shell-specific activation script.
//
// The script modifies PATH to include the shims directory and sets environment
// variables for active tool versions. The script format depends on the target shell.
//
// Requirements: 15.1, 15.2, 15.3
func (m *ActivationManager) GenerateActivationScript(ctx context.Context, config ActivationConfig) (*ActivationScript, error) {
	if config.ShimsDir == "" {
		config.ShimsDir = m.shimsDir
	}

	// Validate configuration
	if err := m.validateConfig(config); err != nil {
		return nil, errors.Wrap(err, "invalid activation configuration")
	}

	logger.Debug("Generating activation script", map[string]interface{}{
		"shell":         config.Shell,
		"scope":         config.Scope,
		"shims_dir":     config.ShimsDir,
		"project_dir":   config.ProjectDir,
		"tool_count":    len(config.ToolVersions),
		"env_var_count": len(config.EnvVars),
	})

	var script *ActivationScript
	var err error

	switch config.Shell {
	case ShellBash, ShellZsh:
		script, err = m.generatePosixScript(config)
	case ShellFish:
		script, err = m.generateFishScript(config)
	case ShellPowerShell:
		script, err = m.generatePowerShellScript(config)
	default:
		return nil, errors.NewUserError(fmt.Sprintf("unsupported shell type: %s", config.Shell), nil)
	}

	if err != nil {
		return nil, errors.Wrap(err, "generate activation script")
	}

	logger.Debug("Generated activation script", map[string]interface{}{
		"shell":       config.Shell,
		"scope":       config.Scope,
		"script_size": len(script.Content),
	})

	return script, nil
}

// GenerateGlobalActivation generates activation for global (system-wide) tool versions.
//
// Requirements: 15.5
func (m *ActivationManager) GenerateGlobalActivation(ctx context.Context, shell ShellType, toolVersions map[string]string) (*ActivationScript, error) {
	config := ActivationConfig{
		Shell:        shell,
		Scope:        ScopeGlobal,
		ShimsDir:     m.shimsDir,
		ToolVersions: toolVersions,
		EnvVars:      make(map[string]string),
		UseShims:     true,
	}

	return m.GenerateActivationScript(ctx, config)
}

// GenerateProjectActivation generates activation for project-specific tool versions.
//
// Requirements: 15.4
func (m *ActivationManager) GenerateProjectActivation(ctx context.Context, shell ShellType, projectDir string, toolVersions map[string]string, envVars map[string]string) (*ActivationScript, error) {
	// If envVars is nil, initialize it
	if envVars == nil {
		envVars = make(map[string]string)
	}

	var injectedPaths []string
	installsDir := filepath.Join(m.dataDir, "installs")

	// Resolve tool binary paths and environment variables
	for toolNameKey, version := range toolVersions {
		toolName := toolNameKey
		backendName := ""

		// Resolve backend and tool name from key if not explicit
		if idx := strings.Index(toolNameKey, ":"); idx != -1 {
			backendName = toolNameKey[:idx]
			toolName = toolNameKey[idx+1:]
		} else if strings.Contains(toolNameKey, "/") {
			backendName = "github"
		}

		// Intercept go: prefix (align with installation manager)
		if backendName == "go" || strings.HasPrefix(toolNameKey, "go:") {
			backendName = "go-pkg"
			if strings.HasPrefix(toolNameKey, "go:") {
				toolName = strings.TrimPrefix(toolNameKey, "go:")
			}
		}

		// Standardize tool name for filesystem (Scheme B: provider-tool-name)
		fsToolName := env.GetFSToolName(toolName, backendName)

		p := m.registry.GetWithBackend(toolName, backendName)
		if p == nil {
			continue
		}
		installPath := filepath.Join(installsDir, fsToolName, version)

		// Get bin paths
		binPaths, err := p.GetBinPaths(toolName, installPath, version)
		if err == nil {
			injectedPaths = append(injectedPaths, binPaths...)
		}

		// Get env vars
		toolEnvVars, err := p.GetEnvVars(toolName, installPath, version)
		if err == nil {
			for k, v := range toolEnvVars {
				if existing, exists := envVars[k]; !exists {
					envVars[k] = v
				} else if k == "NODE_PATH" {
					// Concatenate NODE_PATH for multiple npm tools so they can share plugins
					sep := string(os.PathListSeparator)
					if !strings.Contains(existing+sep, v+sep) {
						envVars[k] = existing + sep + v
					}
				}
			}
		}
	}

	config := ActivationConfig{
		Shell:         shell,
		Scope:         ScopeProject,
		ShimsDir:      m.shimsDir,
		ProjectDir:    projectDir,
		ToolVersions:  toolVersions,
		EnvVars:       envVars,
		InjectedPaths: injectedPaths,
		UseShims:      false, // Default to PATH mode for project activation
	}

	return m.GenerateActivationScript(ctx, config)
}

// validateConfig validates the activation configuration.
func (m *ActivationManager) validateConfig(config ActivationConfig) error {
	if config.Shell == "" {
		return errors.NewUserError("shell type is required", nil)
	}

	if config.ShimsDir == "" {
		return errors.NewUserError("shims directory is required", nil)
	}

	if config.Scope == ScopeProject && config.ProjectDir == "" {
		return errors.NewUserError("project directory is required for project-specific activation", nil)
	}

	return nil
}

// generatePosixScript generates activation script for POSIX shells (bash, zsh).
//
// Requirements: 15.1, 15.2, 15.3
func (m *ActivationManager) generatePosixScript(config ActivationConfig) (*ActivationScript, error) {
	var sb strings.Builder

	// Header comment
	sb.WriteString("# UniRTM activation script\n")
	sb.WriteString(fmt.Sprintf("# Shell: %s\n", config.Shell))
	sb.WriteString(fmt.Sprintf("# Scope: %s\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("# Project: %s\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	if config.UseShims {
		// Add shims directory to PATH
		sb.WriteString("# Add UniRTM shims to PATH\n")
		// Clean up existing shims from PATH to avoid duplicates
		sb.WriteString(fmt.Sprintf(`export PATH="%s:$(echo "$PATH" | sed -E 's|%s:?||g' | sed 's|:$||')"`+"\n", config.ShimsDir, config.ShimsDir))
		sb.WriteString("\n")
	} else if len(config.InjectedPaths) > 0 {
		// PATH mode activation
		sb.WriteString("# UniRTM PATH mode activation\n")
		injectedPath := strings.Join(config.InjectedPaths, string(os.PathListSeparator))

		// Use UNIRTM_PATH to track injected paths.
		// Use a shell loop to filter out existing UNIRTM-managed entries from PATH,
		// avoiding sed command length limits when there are many tools.
		sb.WriteString(fmt.Sprintf("export UNIRTM_PATH=\"%s\"\n", injectedPath))
		sb.WriteString("_unirtm_clean_path() {\n")
		sb.WriteString("  local result=\"\"\n")
		sb.WriteString("  local IFS=:\n")
		sb.WriteString("  for _p in $PATH; do\n")
		sb.WriteString("    case \":$UNIRTM_PATH:\" in\n")
		sb.WriteString("      *\":$_p:\"*) ;;\n")
		sb.WriteString("      *) result=\"${result:+$result:}$_p\" ;;\n")
		sb.WriteString("    esac\n")
		sb.WriteString("  done\n")
		sb.WriteString("  echo \"$result\"\n")
		sb.WriteString("}\n")
		sb.WriteString("export PATH=\"$UNIRTM_PATH:$(_unirtm_clean_path)\"\n")
		sb.WriteString("unset -f _unirtm_clean_path\n")
		sb.WriteString("\n")
	}

	// Set tool version environment variables
	if len(config.ToolVersions) > 0 {
		sb.WriteString("# Set active tool versions\n")
		for tool, version := range config.ToolVersions {
			envVar := m.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n", envVar, version))
		}
		sb.WriteString("\n")
	}

	// Set additional environment variables
	if len(config.EnvVars) > 0 {
		sb.WriteString("# Set additional environment variables\n")
		for key, value := range config.EnvVars {
			sb.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Source additional scripts
	if len(config.Sources) > 0 {
		sb.WriteString("# Source additional scripts\n")
		for _, s := range config.Sources {
			sb.WriteString(fmt.Sprintf("source \"%s\"\n", s))
		}
		sb.WriteString("\n")
	}

	// Set activation markers
	sb.WriteString("export UNIRTM_ACTIVE=1\n")
	sb.WriteString(fmt.Sprintf("export UNIRTM_ACTIVATION_SCOPE=\"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("export UNIRTM_PROJECT_DIR=\"%s\"\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	// Hot-reloading hook (only in PATH mode)
	if !config.UseShims {
		aam := NewAutoActivationManager(m)
		hookScript, err := aam.GenerateHookEnvScript(config.Shell, config.ExePath)
		if err == nil {
			sb.WriteString("\n")
			sb.WriteString(hookScript)
			sb.WriteString("\n")
		}
	}

	instructions := m.generatePosixInstructions(config.Shell)

	return &ActivationScript{
		Shell:        config.Shell,
		Content:      sb.String(),
		Instructions: instructions,
	}, nil
}

// generateFishScript generates activation script for fish shell.
//
// Requirements: 15.1, 15.2, 15.3
func (m *ActivationManager) generateFishScript(config ActivationConfig) (*ActivationScript, error) {
	var sb strings.Builder

	// Header comment
	sb.WriteString("# UniRTM activation script\n")
	sb.WriteString("# Shell: fish\n")
	sb.WriteString(fmt.Sprintf("# Scope: %s\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("# Project: %s\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	if config.UseShims {
		// Add shims directory to PATH
		sb.WriteString("# Add UniRTM shims to PATH\n")
		sb.WriteString(fmt.Sprintf("set -gx PATH \"%s\" $PATH\n", config.ShimsDir))
		sb.WriteString("\n")
	} else if len(config.InjectedPaths) > 0 {
		// PATH mode activation
		sb.WriteString("# UniRTM PATH mode activation\n")
		// In fish, PATH is a list. We filter out existing UniRTM paths to avoid duplicates.
		injectedPath := strings.Join(config.InjectedPaths, " ")
		sb.WriteString(fmt.Sprintf("set -gx UNIRTM_PATH %s\n", injectedPath))
		sb.WriteString("set -l new_path\n")
		sb.WriteString("for p in $PATH\n")
		sb.WriteString("    if not contains $p $UNIRTM_PATH\n")
		sb.WriteString("        set -a new_path $p\n")
		sb.WriteString("    end\n")
		sb.WriteString("end\n")
		sb.WriteString("set -gx PATH $UNIRTM_PATH $new_path\n")
		sb.WriteString("\n")
	}

	// Set tool version environment variables
	if len(config.ToolVersions) > 0 {
		sb.WriteString("# Set active tool versions\n")
		for tool, version := range config.ToolVersions {
			envVar := m.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("set -gx %s \"%s\"\n", envVar, version))
		}
		sb.WriteString("\n")
	}

	// Set additional environment variables
	if len(config.EnvVars) > 0 {
		sb.WriteString("# Set additional environment variables\n")
		for key, value := range config.EnvVars {
			sb.WriteString(fmt.Sprintf("set -gx %s \"%s\"\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Source additional scripts
	if len(config.Sources) > 0 {
		sb.WriteString("# Source additional scripts\n")
		for _, s := range config.Sources {
			sb.WriteString(fmt.Sprintf("source \"%s\"\n", s))
		}
		sb.WriteString("\n")
	}

	// Set activation markers
	sb.WriteString("set -gx UNIRTM_ACTIVE 1\n")
	sb.WriteString(fmt.Sprintf("set -gx UNIRTM_ACTIVATION_SCOPE \"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("set -gx UNIRTM_PROJECT_DIR \"%s\"\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	// Hot-reloading hook (only in PATH mode)
	if !config.UseShims {
		aam := NewAutoActivationManager(m)
		hookScript, err := aam.GenerateHookEnvScript(ShellFish, config.ExePath)
		if err == nil {
			sb.WriteString("\n")
			sb.WriteString(hookScript)
			sb.WriteString("\n")
		}
	}

	instructions := "To activate this environment, run:\n\n" +
		"    source /path/to/activation.fish\n\n" +
		"Or save the script to a file and source it in your fish config:\n\n" +
		"    unirtm activate --shell fish > ~/.config/fish/conf.d/unirtm.fish"

	return &ActivationScript{
		Shell:        ShellFish,
		Content:      sb.String(),
		Instructions: instructions,
	}, nil
}

// generatePowerShellScript generates activation script for PowerShell.
//
// Requirements: 15.1, 15.2, 15.3
func (m *ActivationManager) generatePowerShellScript(config ActivationConfig) (*ActivationScript, error) {
	var sb strings.Builder

	// Header comment
	sb.WriteString("# UniRTM activation script\n")
	sb.WriteString("# Shell: PowerShell\n")
	sb.WriteString(fmt.Sprintf("# Scope: %s\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("# Project: %s\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	if config.UseShims {
		// Add shims directory to PATH
		sb.WriteString("# Add UniRTM shims to PATH\n")
		shimsDir := config.ShimsDir
		if runtime.GOOS == "windows" {
			// Convert forward slashes to backslashes on Windows
			shimsDir = filepath.FromSlash(shimsDir)
		}
		sb.WriteString(fmt.Sprintf("$shimsDir = \"%s\"\n", shimsDir))
		sb.WriteString("$env:PATH = \"$shimsDir;\" + (($env:PATH -split ';') | Where-Object { $_ -ne $shimsDir } -join ';')\n")
		sb.WriteString("\n")
	} else if len(config.InjectedPaths) > 0 {
		// PATH mode activation
		sb.WriteString("# UniRTM PATH mode activation\n")
		var paths []string
		for _, p := range config.InjectedPaths {
			path := p
			if runtime.GOOS == "windows" {
				path = filepath.FromSlash(p)
			}
			paths = append(paths, path)
		}
		injectedPath := strings.Join(paths, ";")
		sb.WriteString(fmt.Sprintf("$unirtmPaths = \"%s\" -split ';'\n", injectedPath))
		sb.WriteString("$env:UNIRTM_PATH = $unirtmPaths -join ';'\n")
		sb.WriteString("$env:PATH = ($env:UNIRTM_PATH + ';' + (($env:PATH -split ';') | Where-Object { $unirtmPaths -notcontains $_ } -join ';'))\n")
		sb.WriteString("\n")
	}

	// Set tool version environment variables
	if len(config.ToolVersions) > 0 {
		sb.WriteString("# Set active tool versions\n")
		for tool, version := range config.ToolVersions {
			envVar := m.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("$env:%s = \"%s\"\n", envVar, version))
		}
		sb.WriteString("\n")
	}

	// Set additional environment variables
	if len(config.EnvVars) > 0 {
		sb.WriteString("# Set additional environment variables\n")
		for key, value := range config.EnvVars {
			sb.WriteString(fmt.Sprintf("$env:%s = \"%s\"\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Source additional scripts
	if len(config.Sources) > 0 {
		sb.WriteString("# Source additional scripts\n")
		for _, s := range config.Sources {
			// In PowerShell, use dot-sourcing
			path := s
			if runtime.GOOS == "windows" {
				path = filepath.FromSlash(s)
			}
			sb.WriteString(fmt.Sprintf(". \"%s\"\n", path))
		}
		sb.WriteString("\n")
	}

	// Set activation markers
	sb.WriteString("$env:UNIRTM_ACTIVE = \"1\"\n")
	sb.WriteString(fmt.Sprintf("$env:UNIRTM_ACTIVATION_SCOPE = \"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		projectDir := config.ProjectDir
		if runtime.GOOS == "windows" {
			projectDir = filepath.FromSlash(projectDir)
		}
		sb.WriteString(fmt.Sprintf("$env:UNIRTM_PROJECT_DIR = \"%s\"\n", projectDir))
	}
	sb.WriteString("\n")

	// Hot-reloading hook (only in PATH mode)
	if !config.UseShims {
		aam := NewAutoActivationManager(m)
		hookScript, err := aam.GenerateHookEnvScript(ShellPowerShell, config.ExePath)
		if err == nil {
			sb.WriteString("\n")
			sb.WriteString(hookScript)
			sb.WriteString("\n")
		}
	}

	instructions := "To activate this environment, run:\n\n" +
		"    . \\path\\to\\activation.ps1\n\n" +
		"Or save the script to a file and dot-source it in your PowerShell profile:\n\n" +
		"    unirtm activate --shell powershell | Out-File -FilePath $PROFILE\\unirtm.ps1\n" +
		"    . $PROFILE\\unirtm.ps1"

	return &ActivationScript{
		Shell:        ShellPowerShell,
		Content:      sb.String(),
		Instructions: instructions,
	}, nil
}

// generatePosixInstructions generates usage instructions for POSIX shells.
func (m *ActivationManager) generatePosixInstructions(shell ShellType) string {
	shellName := string(shell)

	return fmt.Sprintf("UniRTM environment for %s is ready.\n\n"+
		"To persist this, add the following to your %s config:\n\n"+
		"    eval \"$(unirtm activate %s)\"",
		shellName, shellName, shellName)
}

func (m *ActivationManager) toolVersionEnvVar(tool string) string {
	// Convert tool name to uppercase and replace all non-alphanumeric characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	envVar := strings.ToUpper(tool)
	envVar = reg.ReplaceAllString(envVar, "_")
	return fmt.Sprintf("UNIRTM_%s_VERSION", envVar)
}

// DetectShell detects the current shell from the environment.
//
// It checks the SHELL environment variable and returns the corresponding ShellType.
// If the shell cannot be detected, it returns a sensible default for the platform.
func DetectShell() (ShellType, error) {
	// On Unix-like systems, check SHELL environment variable first
	if shellPath := env.Get("SHELL"); shellPath != "" {
		shell := filepath.Base(shellPath)
		switch {
		case strings.Contains(shell, "bash"):
			return ShellBash, nil
		case strings.Contains(shell, "zsh"):
			return ShellZsh, nil
		case strings.Contains(shell, "fish"):
			return ShellFish, nil
		}
	}

	// On Windows, default to PowerShell if SHELL is not set or not recognized
	if runtime.GOOS == "windows" {
		return ShellPowerShell, nil
	}

	// Default to bash on Unix-like systems
	return ShellBash, nil
}
