// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides high-level business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
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
	Sources []string
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
}

// NewActivationManager creates a new ActivationManager.
func NewActivationManager(shimsDir, dataDir string) *ActivationManager {
	return &ActivationManager{
		shimsDir: shimsDir,
		dataDir:  dataDir,
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

	logger.Info("Generated activation script", map[string]interface{}{
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
	}

	return m.GenerateActivationScript(ctx, config)
}

// GenerateProjectActivation generates activation for project-specific tool versions.
//
// Requirements: 15.4
func (m *ActivationManager) GenerateProjectActivation(ctx context.Context, shell ShellType, projectDir string, toolVersions map[string]string, envVars map[string]string) (*ActivationScript, error) {
	config := ActivationConfig{
		Shell:        shell,
		Scope:        ScopeProject,
		ShimsDir:     m.shimsDir,
		ProjectDir:   projectDir,
		ToolVersions: toolVersions,
		EnvVars:      envVars,
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

	// Add shims directory to PATH
	sb.WriteString("# Add UniRTM shims to PATH\n")
	sb.WriteString(fmt.Sprintf("export PATH=\"%s:$PATH\"\n", config.ShimsDir))
	sb.WriteString("\n")

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

	// Set scope indicator
	sb.WriteString(fmt.Sprintf("export UNIRTM_ACTIVATION_SCOPE=\"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("export UNIRTM_PROJECT_DIR=\"%s\"\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	// Hot-reloading hook
	sb.WriteString("# UniRTM Hot Reloading Hook\n")
	sb.WriteString("if [[ \"$OSTYPE\" == \"darwin\"* ]]; then\n")
	sb.WriteString("  export _UNIRTM_LAST_MTIME=$(stat -f \"%m\" unirtm.toml 2>/dev/null || echo \"\")\n")
	sb.WriteString("else\n")
	sb.WriteString("  export _UNIRTM_LAST_MTIME=$(stat -c \"%Y\" unirtm.toml 2>/dev/null || echo \"\")\n")
	sb.WriteString("fi\n\n")

	sb.WriteString("_unirtm_hook() {\n")
	sb.WriteString("  local config_file=\"unirtm.toml\"\n")
	sb.WriteString("  if [ -f \"$config_file\" ]; then\n")
	sb.WriteString("    local current_mtime\n")
	sb.WriteString("    if [[ \"$OSTYPE\" == \"darwin\"* ]]; then\n")
	sb.WriteString("      current_mtime=$(stat -f \"%m\" \"$config_file\" 2>/dev/null)\n")
	sb.WriteString("    else\n")
	sb.WriteString("      current_mtime=$(stat -c \"%Y\" \"$config_file\" 2>/dev/null)\n")
	sb.WriteString("    fi\n")
	sb.WriteString("    if [ \"$current_mtime\" != \"$_UNIRTM_LAST_MTIME\" ]; then\n")
	sb.WriteString("      export _UNIRTM_LAST_MTIME=\"$current_mtime\"\n")
	sb.WriteString(fmt.Sprintf("      eval \"$(unirtm activate --shell %s)\"\n", config.Shell))
	sb.WriteString("    fi\n")
	sb.WriteString("  fi\n")
	sb.WriteString("}\n\n")

	sb.WriteString("if [[ -n \"${ZSH_VERSION-}\" ]]; then\n")
	sb.WriteString("  autoload -Uz add-zsh-hook\n")
	sb.WriteString("  add-zsh-hook -d precmd _unirtm_hook 2>/dev/null\n")
	sb.WriteString("  add-zsh-hook precmd _unirtm_hook\n")
	sb.WriteString("elif [[ -n \"${BASH_VERSION-}\" ]]; then\n")
	sb.WriteString("  if [[ ! \"$PROMPT_COMMAND\" =~ \"_unirtm_hook\" ]]; then\n")
	sb.WriteString("    PROMPT_COMMAND=\"_unirtm_hook; ${PROMPT_COMMAND:-}\"\n")
	sb.WriteString("  fi\n")
	sb.WriteString("fi\n\n")

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

	// Add shims directory to PATH
	sb.WriteString("# Add UniRTM shims to PATH\n")
	sb.WriteString(fmt.Sprintf("set -gx PATH \"%s\" $PATH\n", config.ShimsDir))
	sb.WriteString("\n")

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

	// Set scope indicator
	sb.WriteString(fmt.Sprintf("set -gx UNIRTM_ACTIVATION_SCOPE \"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		sb.WriteString(fmt.Sprintf("set -gx UNIRTM_PROJECT_DIR \"%s\"\n", config.ProjectDir))
	}
	sb.WriteString("\n")

	// Hot-reloading hook
	sb.WriteString("# UniRTM Hot Reloading Hook\n")
	sb.WriteString("if string match -q \"darwin*\" $OSTYPE\n")
	sb.WriteString("  set -gx _UNIRTM_LAST_MTIME (stat -f \"%m\" unirtm.toml 2>/dev/null)\n")
	sb.WriteString("else\n")
	sb.WriteString("  set -gx _UNIRTM_LAST_MTIME (stat -c \"%Y\" unirtm.toml 2>/dev/null)\n")
	sb.WriteString("end\n\n")

	sb.WriteString("function _unirtm_hook --on-event fish_prompt\n")
	sb.WriteString("  set -l config_file \"unirtm.toml\"\n")
	sb.WriteString("  if test -f \"$config_file\"\n")
	sb.WriteString("    set -l current_mtime\n")
	sb.WriteString("    if string match -q \"darwin*\" $OSTYPE\n")
	sb.WriteString("      set current_mtime (stat -f \"%m\" \"$config_file\" 2>/dev/null)\n")
	sb.WriteString("    else\n")
	sb.WriteString("      set current_mtime (stat -c \"%Y\" \"$config_file\" 2>/dev/null)\n")
	sb.WriteString("    end\n")
	sb.WriteString("    if test \"$current_mtime\" != \"$_UNIRTM_LAST_MTIME\"\n")
	sb.WriteString("      set -gx _UNIRTM_LAST_MTIME \"$current_mtime\"\n")
	sb.WriteString("      unirtm activate --shell fish | source\n")
	sb.WriteString("    end\n")
	sb.WriteString("  end\n")
	sb.WriteString("end\n\n")

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

	// Add shims directory to PATH
	sb.WriteString("# Add UniRTM shims to PATH\n")
	shimsDir := config.ShimsDir
	if runtime.GOOS == "windows" {
		// Convert forward slashes to backslashes on Windows
		shimsDir = filepath.FromSlash(shimsDir)
	}
	sb.WriteString(fmt.Sprintf("$env:PATH = \"%s;$env:PATH\"\n", shimsDir))
	sb.WriteString("\n")

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

	// Set scope indicator
	sb.WriteString(fmt.Sprintf("$env:UNIRTM_ACTIVATION_SCOPE = \"%s\"\n", config.Scope))
	if config.ProjectDir != "" {
		projectDir := config.ProjectDir
		if runtime.GOOS == "windows" {
			projectDir = filepath.FromSlash(projectDir)
		}
		sb.WriteString(fmt.Sprintf("$env:UNIRTM_PROJECT_DIR = \"%s\"\n", projectDir))
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
	configFile := "~/.bashrc"
	if shell == ShellZsh {
		configFile = "~/.zshrc"
	}

	return fmt.Sprintf("To activate this environment, run:\n\n"+
		"    source /path/to/activation.sh\n\n"+
		"Or save the script to a file and source it in your %s config:\n\n"+
		"    unirtm activate --shell %s > ~/unirtm-activation.sh\n"+
		"    echo 'source ~/unirtm-activation.sh' >> %s",
		shellName, shellName, configFile)
}

// toolVersionEnvVar returns the environment variable name for a tool version.
//
// The environment variable name is UNIRTM_<TOOL>_VERSION, where <TOOL> is
// the uppercase tool name with hyphens replaced by underscores.
func (m *ActivationManager) toolVersionEnvVar(tool string) string {
	// Convert tool name to uppercase and replace hyphens with underscores
	envVar := strings.ToUpper(tool)
	envVar = strings.ReplaceAll(envVar, "-", "_")
	return fmt.Sprintf("UNIRTM_%s_VERSION", envVar)
}

// DetectShell detects the current shell from the environment.
//
// It checks the SHELL environment variable and returns the corresponding ShellType.
// If the shell cannot be detected, it returns an error.
func DetectShell() (ShellType, error) {
	// On Windows, default to PowerShell
	if runtime.GOOS == "windows" {
		return ShellPowerShell, nil
	}

	// On Unix-like systems, check SHELL environment variable
	shell := filepath.Base(runtime.GOOS)
	switch {
	case strings.Contains(shell, "bash"):
		return ShellBash, nil
	case strings.Contains(shell, "zsh"):
		return ShellZsh, nil
	case strings.Contains(shell, "fish"):
		return ShellFish, nil
	default:
		// Default to bash on Unix-like systems
		return ShellBash, nil
	}
}
