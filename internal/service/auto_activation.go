// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides high-level business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// AutoActivationManager manages automatic environment activation based on directory changes.
//
// It detects when the user enters a directory with a UniRTM configuration file
// and automatically activates the project's toolchain. When leaving the directory,
// it restores the previous environment.
//
// Requirements: 15.6, 15.7
type AutoActivationManager struct {
	activationManager *ActivationManager
	configFileNames   []string
}

// NewAutoActivationManager creates a new AutoActivationManager.
func NewAutoActivationManager(activationManager *ActivationManager) *AutoActivationManager {
	return &AutoActivationManager{
		activationManager: activationManager,
		configFileNames: []string{
			"unirtm.toml",
			".unirtm.toml",
			"mise.toml",
			".mise.toml",
			".tool-versions",
		},
	}
}

// EnvironmentState represents the state of the environment at a point in time.
type EnvironmentState struct {
	// ProjectDir is the project directory (empty if no project is active)
	ProjectDir string
	// ToolVersions maps tool names to their active versions
	ToolVersions map[string]string
	// EnvVars contains environment variables set by the activation
	EnvVars map[string]string
	// PreviousPath is the PATH before activation
	PreviousPath string
}

// DirectoryChangeEvent represents a directory change event.
type DirectoryChangeEvent struct {
	// OldDir is the previous working directory
	OldDir string
	// NewDir is the new working directory
	NewDir string
	// Shell is the current shell type
	Shell ShellType
}

// ActivationChange represents the changes needed to update the environment.
type ActivationChange struct {
	// Action is the type of change (activate, deactivate, switch)
	Action ActivationAction
	// Script is the shell script to execute the change
	Script string
	// NewState is the new environment state after the change
	NewState *EnvironmentState
}

// ActivationAction represents the type of activation change.
type ActivationAction string

const (
	// ActionActivate indicates activating a new project environment
	ActionActivate ActivationAction = "activate"
	// ActionDeactivate indicates deactivating the current project environment
	ActionDeactivate ActivationAction = "deactivate"
	// ActionSwitch indicates switching from one project to another
	ActionSwitch ActivationAction = "switch"
	// ActionNone indicates no change is needed
	ActionNone ActivationAction = "none"
)

// HandleDirectoryChange handles a directory change event and returns the activation changes needed.
//
// This is the main entry point for auto-activation. It detects whether the directory
// change requires activating, deactivating, or switching project environments.
//
// Requirements: 15.6, 15.7
func (m *AutoActivationManager) HandleDirectoryChange(ctx context.Context, event DirectoryChangeEvent, currentState *EnvironmentState) (*ActivationChange, error) {
	logger.Debug("Handling directory change", map[string]interface{}{
		"old_dir":         event.OldDir,
		"new_dir":         event.NewDir,
		"shell":           event.Shell,
		"current_project": currentState.ProjectDir,
	})

	// Find project directories for old and new locations
	oldProjectDir := m.findProjectDirectory(event.OldDir)
	newProjectDir := m.findProjectDirectory(event.NewDir)

	// Determine the action needed
	action := m.determineAction(oldProjectDir, newProjectDir, currentState.ProjectDir)

	logger.Debug("Determined activation action", map[string]interface{}{
		"action":          action,
		"old_project":     oldProjectDir,
		"new_project":     newProjectDir,
		"current_project": currentState.ProjectDir,
	})

	switch action {
	case ActionNone:
		return &ActivationChange{
			Action:   ActionNone,
			Script:   "",
			NewState: currentState,
		}, nil

	case ActionActivate:
		return m.generateActivation(ctx, event.Shell, newProjectDir, currentState)

	case ActionDeactivate:
		return m.generateDeactivation(ctx, event.Shell, currentState)

	case ActionSwitch:
		return m.generateSwitch(ctx, event.Shell, oldProjectDir, newProjectDir, currentState)

	default:
		return nil, errors.NewSystemError(fmt.Sprintf("unknown activation action: %s", action), nil)
	}
}

// findProjectDirectory searches for a UniRTM configuration file starting from the given directory
// and walking up the directory tree. Returns the directory containing the config file, or empty string if not found.
//
// Requirements: 15.6
func (m *AutoActivationManager) findProjectDirectory(startDir string) string {
	if startDir == "" {
		return ""
	}

	// Clean the path to handle relative paths and symlinks
	startDir, err := filepath.Abs(startDir)
	if err != nil {
		logger.Debug("Failed to resolve absolute path", map[string]interface{}{
			"path":  startDir,
			"error": err.Error(),
		})
		return ""
	}

	currentDir := startDir

	// Walk up the directory tree
	for {
		// Check for each possible config file name
		for _, configFileName := range m.configFileNames {
			configPath := filepath.Join(currentDir, configFileName)
			if _, err := os.Stat(configPath); err == nil {
				logger.Debug("Found project configuration", map[string]interface{}{
					"project_dir": currentDir,
					"config_file": configFileName,
				})
				return currentDir
			}
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)

		// Stop if we've reached the root
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	logger.Debug("No project configuration found", map[string]interface{}{
		"start_dir": startDir,
	})

	return ""
}

// determineAction determines what activation action is needed based on the directory change.
func (m *AutoActivationManager) determineAction(oldProjectDir, newProjectDir, currentProjectDir string) ActivationAction {
	// Normalize empty strings for comparison
	if oldProjectDir == "" {
		oldProjectDir = currentProjectDir
	}

	// Case 1: No project before or after - no action needed
	if oldProjectDir == "" && newProjectDir == "" {
		return ActionNone
	}

	// Case 2: Entering a project from outside
	if oldProjectDir == "" && newProjectDir != "" {
		return ActionActivate
	}

	// Case 3: Leaving a project
	if oldProjectDir != "" && newProjectDir == "" {
		return ActionDeactivate
	}

	// Case 4: Same project - no action needed
	if oldProjectDir == newProjectDir {
		return ActionNone
	}

	// Case 5: Switching between different projects
	return ActionSwitch
}

// generateActivation generates the activation script for entering a project.
//
// Requirements: 15.6
func (m *AutoActivationManager) generateActivation(ctx context.Context, shell ShellType, projectDir string, currentState *EnvironmentState) (*ActivationChange, error) {
	logger.Info("Generating project activation", map[string]interface{}{
		"project_dir": projectDir,
		"shell":       shell,
	})

	// Load project configuration
	toolVersions, envVars, sources, err := m.LoadConfigByDir(projectDir)
	if err != nil {
		logger.Error("Failed to load project config", map[string]interface{}{
			"project_dir": projectDir,
			"error":       err.Error(),
		})
		// Continue with empty config if loading fails
	}

	// Generate activation script
	activationScript, err := m.activationManager.GenerateProjectActivation(
		ctx,
		shell,
		projectDir,
		toolVersions,
		envVars,
	)
	if err != nil {
		return nil, errors.Wrap(err, "generate project activation")
	}

	// Add sources to script
	var scriptContent strings.Builder
	scriptContent.WriteString(activationScript.Content)
	if len(sources) > 0 {
		scriptContent.WriteString("\n# Project-specific sources\n")
		for _, s := range sources {
			switch shell {
			case ShellFish:
				scriptContent.WriteString(fmt.Sprintf("source \"%s\"\n", s))
			case ShellPowerShell:
				scriptContent.WriteString(fmt.Sprintf(". \"%s\"\n", s))
			default: // POSIX
				scriptContent.WriteString(fmt.Sprintf("source \"%s\"\n", s))
			}
		}
	}

	// Create new environment state
	newState := &EnvironmentState{
		ProjectDir:   projectDir,
		ToolVersions: toolVersions,
		EnvVars:      envVars,
		PreviousPath: currentState.PreviousPath,
	}

	// If this is the first activation, save the current PATH
	if newState.PreviousPath == "" {
		newState.PreviousPath = os.Getenv("PATH")
	}

	return &ActivationChange{
		Action:   ActionActivate,
		Script:   scriptContent.String(),
		NewState: newState,
	}, nil
}

// LoadConfigByDir loads UniRTM configuration from the specified directory.
func (m *AutoActivationManager) LoadConfigByDir(projectDir string) (map[string]string, map[string]string, []string, error) {
	toolVersions := make(map[string]string)
	envVars := make(map[string]string)
	var sources []string

	// 1. Try to load unirtm.toml or .unirtm.toml
	configPath := ""
	for _, name := range []string{"unirtm.toml", ".unirtm.toml"} {
		p := filepath.Join(projectDir, name)
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			// Minimal parsing for speed
			var raw struct {
				Tools map[string]interface{} `toml:"tools"`
				Env   map[string]interface{} `toml:"env"`
			}
			if err := toml.Unmarshal(data, &raw); err == nil {
				// Extract tools
				for k, v := range raw.Tools {
					switch val := v.(type) {
					case string:
						toolVersions[k] = val
					case map[string]interface{}:
						if ver, ok := val["version"].(string); ok {
							toolVersions[k] = ver
						}
					}
				}
				// Extract simple env vars
				for k, v := range raw.Env {
					if s, ok := v.(string); ok {
						envVars[k] = s
					}
				}
			}
		}
	}

	// 2. Try to load .tool-versions (asdf compatibility)
	tvPath := filepath.Join(projectDir, ".tool-versions")
	if _, err := os.Stat(tvPath); err == nil {
		data, err := os.ReadFile(tvPath)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					toolVersions[parts[0]] = parts[1]
				}
			}
		}
	}

	return toolVersions, envVars, sources, nil
}

// generateDeactivation generates the deactivation script for leaving a project.
//
// Requirements: 15.7
func (m *AutoActivationManager) generateDeactivation(ctx context.Context, shell ShellType, currentState *EnvironmentState) (*ActivationChange, error) {
	logger.Info("Generating project deactivation", map[string]interface{}{
		"project_dir": currentState.ProjectDir,
		"shell":       shell,
	})

	// Generate deactivation script that restores the previous environment
	script := m.generateDeactivationScript(shell, currentState)

	// Create new environment state (no active project)
	newState := &EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
		PreviousPath: "",
	}

	return &ActivationChange{
		Action:   ActionDeactivate,
		Script:   script,
		NewState: newState,
	}, nil
}

// generateSwitch generates the script for switching between projects.
//
// Requirements: 15.6, 15.7
func (m *AutoActivationManager) generateSwitch(ctx context.Context, shell ShellType, oldProjectDir, newProjectDir string, currentState *EnvironmentState) (*ActivationChange, error) {
	logger.Info("Generating project switch", map[string]interface{}{
		"old_project": oldProjectDir,
		"new_project": newProjectDir,
		"shell":       shell,
	})

	// First deactivate the old project
	deactivation, err := m.generateDeactivation(ctx, shell, currentState)
	if err != nil {
		return nil, errors.Wrap(err, "generate deactivation for switch")
	}

	// Then activate the new project
	activation, err := m.generateActivation(ctx, shell, newProjectDir, deactivation.NewState)
	if err != nil {
		return nil, errors.Wrap(err, "generate activation for switch")
	}

	// Combine the scripts
	var combinedScript strings.Builder
	combinedScript.WriteString("# Deactivating previous project\n")
	combinedScript.WriteString(deactivation.Script)
	combinedScript.WriteString("\n")
	combinedScript.WriteString("# Activating new project\n")
	combinedScript.WriteString(activation.Script)

	return &ActivationChange{
		Action:   ActionSwitch,
		Script:   combinedScript.String(),
		NewState: activation.NewState,
	}, nil
}

// generateDeactivationScript generates a shell-specific deactivation script.
//
// Requirements: 15.7
func (m *AutoActivationManager) generateDeactivationScript(shell ShellType, state *EnvironmentState) string {
	var sb strings.Builder

	sb.WriteString("# UniRTM deactivation script\n")
	sb.WriteString(fmt.Sprintf("# Shell: %s\n", shell))
	sb.WriteString("\n")

	switch shell {
	case ShellBash, ShellZsh:
		m.generatePosixDeactivation(&sb, state)
	case ShellFish:
		m.generateFishDeactivation(&sb, state)
	case ShellPowerShell:
		m.generatePowerShellDeactivation(&sb, state)
	}

	return sb.String()
}

// generatePosixDeactivation generates deactivation script for POSIX shells.
func (m *AutoActivationManager) generatePosixDeactivation(sb *strings.Builder, state *EnvironmentState) {
	// Restore PATH
	if state.PreviousPath != "" {
		sb.WriteString("# Restore previous PATH\n")
		sb.WriteString(fmt.Sprintf("export PATH=\"%s\"\n", state.PreviousPath))
		sb.WriteString("\n")
	}

	// Unset tool version environment variables
	if len(state.ToolVersions) > 0 {
		sb.WriteString("# Unset tool version variables\n")
		for tool := range state.ToolVersions {
			envVar := m.activationManager.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("unset %s\n", envVar))
		}
		sb.WriteString("\n")
	}

	// Unset additional environment variables
	if len(state.EnvVars) > 0 {
		sb.WriteString("# Unset additional environment variables\n")
		for key := range state.EnvVars {
			sb.WriteString(fmt.Sprintf("unset %s\n", key))
		}
		sb.WriteString("\n")
	}

	// Unset UniRTM-specific variables
	sb.WriteString("# Unset UniRTM variables\n")
	sb.WriteString("unset UNIRTM_ACTIVATION_SCOPE\n")
	sb.WriteString("unset UNIRTM_PROJECT_DIR\n")
}

// generateFishDeactivation generates deactivation script for fish shell.
func (m *AutoActivationManager) generateFishDeactivation(sb *strings.Builder, state *EnvironmentState) {
	// Restore PATH
	if state.PreviousPath != "" {
		sb.WriteString("# Restore previous PATH\n")
		sb.WriteString(fmt.Sprintf("set -gx PATH \"%s\"\n", state.PreviousPath))
		sb.WriteString("\n")
	}

	// Unset tool version environment variables
	if len(state.ToolVersions) > 0 {
		sb.WriteString("# Unset tool version variables\n")
		for tool := range state.ToolVersions {
			envVar := m.activationManager.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("set -e %s\n", envVar))
		}
		sb.WriteString("\n")
	}

	// Unset additional environment variables
	if len(state.EnvVars) > 0 {
		sb.WriteString("# Unset additional environment variables\n")
		for key := range state.EnvVars {
			sb.WriteString(fmt.Sprintf("set -e %s\n", key))
		}
		sb.WriteString("\n")
	}

	// Unset UniRTM-specific variables
	sb.WriteString("# Unset UniRTM variables\n")
	sb.WriteString("set -e UNIRTM_ACTIVATION_SCOPE\n")
	sb.WriteString("set -e UNIRTM_PROJECT_DIR\n")
}

// generatePowerShellDeactivation generates deactivation script for PowerShell.
func (m *AutoActivationManager) generatePowerShellDeactivation(sb *strings.Builder, state *EnvironmentState) {
	// Restore PATH
	if state.PreviousPath != "" {
		sb.WriteString("# Restore previous PATH\n")
		sb.WriteString(fmt.Sprintf("$env:PATH = \"%s\"\n", state.PreviousPath))
		sb.WriteString("\n")
	}

	// Unset tool version environment variables
	if len(state.ToolVersions) > 0 {
		sb.WriteString("# Unset tool version variables\n")
		for tool := range state.ToolVersions {
			envVar := m.activationManager.toolVersionEnvVar(tool)
			sb.WriteString(fmt.Sprintf("Remove-Item Env:%s -ErrorAction SilentlyContinue\n", envVar))
		}
		sb.WriteString("\n")
	}

	// Unset additional environment variables
	if len(state.EnvVars) > 0 {
		sb.WriteString("# Unset additional environment variables\n")
		for key := range state.EnvVars {
			sb.WriteString(fmt.Sprintf("Remove-Item Env:%s -ErrorAction SilentlyContinue\n", key))
		}
		sb.WriteString("\n")
	}

	// Unset UniRTM-specific variables
	sb.WriteString("# Unset UniRTM variables\n")
	sb.WriteString("Remove-Item Env:UNIRTM_ACTIVATION_SCOPE -ErrorAction SilentlyContinue\n")
	sb.WriteString("Remove-Item Env:UNIRTM_PROJECT_DIR -ErrorAction SilentlyContinue\n")
}

// GenerateHookEnvScript generates a shell hook script that can be evaluated on every prompt.
//
// This script is designed to be evaluated by the shell on every prompt (e.g., via PROMPT_COMMAND
// in bash or precmd in zsh). It detects directory changes and outputs the appropriate activation
// or deactivation commands.
//
// Requirements: 15.6, 15.7
func (m *AutoActivationManager) GenerateHookEnvScript(shell ShellType) (string, error) {
	var sb strings.Builder

	switch shell {
	case ShellBash, ShellZsh:
		m.generatePosixHook(&sb, shell)
	case ShellFish:
		m.generateFishHook(&sb)
	case ShellPowerShell:
		m.generatePowerShellHook(&sb)
	default:
		return "", errors.NewUserError(fmt.Sprintf("unsupported shell type: %s", shell), nil)
	}

	return sb.String(), nil
}

// generatePosixHook generates the hook script for POSIX shells.
func (m *AutoActivationManager) generatePosixHook(sb *strings.Builder, shell ShellType) {
	sb.WriteString("# UniRTM auto-activation hook\n")
	sb.WriteString("_unirtm_hook() {\n")
	sb.WriteString("  local old_pwd=\"${UNIRTM_OLD_PWD:-}\"\n")
	sb.WriteString("  local new_pwd=\"$PWD\"\n")
	sb.WriteString("  \n")
	sb.WriteString("  # Only run if directory changed\n")
	sb.WriteString("  if [ \"$old_pwd\" != \"$new_pwd\" ]; then\n")
	sb.WriteString("    export UNIRTM_OLD_PWD=\"$new_pwd\"\n")
	sb.WriteString("    \n")
	sb.WriteString("    # Call unirtm hook-env to get activation changes\n")
	sb.WriteString(fmt.Sprintf("    eval \"$(unirtm hook-env --shell %s)\"\n", shell))
	sb.WriteString("  fi\n")
	sb.WriteString("}\n")
	sb.WriteString("\n")

	if shell == ShellBash {
		sb.WriteString("# Install the hook in bash\n")
		sb.WriteString("if [[ -z \"${PROMPT_COMMAND:-}\" ]]; then\n")
		sb.WriteString("  PROMPT_COMMAND=\"_unirtm_hook\"\n")
		sb.WriteString("else\n")
		sb.WriteString("  PROMPT_COMMAND=\"_unirtm_hook;$PROMPT_COMMAND\"\n")
		sb.WriteString("fi\n")
	} else if shell == ShellZsh {
		sb.WriteString("# Install the hook in zsh\n")
		sb.WriteString("autoload -U add-zsh-hook\n")
		sb.WriteString("add-zsh-hook precmd _unirtm_hook\n")
	}
}

// generateFishHook generates the hook script for fish shell.
func (m *AutoActivationManager) generateFishHook(sb *strings.Builder) {
	sb.WriteString("# UniRTM auto-activation hook for fish\n")
	sb.WriteString("function _unirtm_hook --on-variable PWD\n")
	sb.WriteString("  # Call unirtm hook-env to get activation changes\n")
	sb.WriteString("  unirtm hook-env --shell fish | source\n")
	sb.WriteString("end\n")
}

// generatePowerShellHook generates the hook script for PowerShell.
func (m *AutoActivationManager) generatePowerShellHook(sb *strings.Builder) {
	sb.WriteString("# UniRTM auto-activation hook for PowerShell\n")
	sb.WriteString("function Invoke-UnirtmHook {\n")
	sb.WriteString("  $oldPwd = $env:UNIRTM_OLD_PWD\n")
	sb.WriteString("  $newPwd = $PWD.Path\n")
	sb.WriteString("  \n")
	sb.WriteString("  if ($oldPwd -ne $newPwd) {\n")
	sb.WriteString("    $env:UNIRTM_OLD_PWD = $newPwd\n")
	sb.WriteString("    \n")
	sb.WriteString("    # Call unirtm hook-env to get activation changes\n")
	sb.WriteString("    $script = unirtm hook-env --shell powershell\n")
	sb.WriteString("    if ($script) {\n")
	sb.WriteString("      Invoke-Expression $script\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("\n")
	sb.WriteString("# Install the hook in PowerShell prompt\n")
	sb.WriteString("$global:_unirtm_original_prompt = $function:prompt\n")
	sb.WriteString("function prompt {\n")
	sb.WriteString("  Invoke-UnirtmHook\n")
	sb.WriteString("  & $global:_unirtm_original_prompt\n")
	sb.WriteString("}\n")
}
