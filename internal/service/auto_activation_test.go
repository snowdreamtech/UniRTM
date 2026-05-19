// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"github.com/snowdreamtech/unirtm/internal/provider"

	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAutoActivationManager(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	require.NotNil(t, autoMgr)
	assert.NotNil(t, autoMgr.activationManager)
	assert.NotEmpty(t, autoMgr.configFileNames)
	assert.Contains(t, autoMgr.configFileNames, "unirtm.toml")
	assert.Contains(t, autoMgr.configFileNames, ".tool-versions")
}

func TestFindProjectDirectory(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create nested directories
	projectDir := filepath.Join(tmpDir, "projects", "myproject")
	subDir := filepath.Join(projectDir, "src", "components")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Create config file in project root
	configPath := filepath.Join(projectDir, "unirtm.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("# test config"), 0644))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	tests := []struct {
		name     string
		startDir string
		want     string
	}{
		{
			name:     "finds config in current directory",
			startDir: projectDir,
			want:     projectDir,
		},
		{
			name:     "finds config in parent directory",
			startDir: subDir,
			want:     projectDir,
		},
		{
			name:     "returns empty when no config found",
			startDir: tmpDir,
			want:     "",
		},
		{
			name:     "handles empty start directory",
			startDir: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := autoMgr.findProjectDirectory(tt.startDir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindProjectDirectory_MultipleConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Test each config file name
	configFiles := []string{
		"unirtm.toml",
		".unirtm.toml",
		"mise.toml",
		".mise.toml",
		".tool-versions",
	}

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	for _, configFile := range configFiles {
		t.Run(configFile, func(t *testing.T) {
			// Create a subdirectory for this test
			testDir := filepath.Join(tmpDir, configFile)
			require.NoError(t, os.MkdirAll(testDir, 0755))

			// Create the config file
			configPath := filepath.Join(testDir, configFile)
			require.NoError(t, os.WriteFile(configPath, []byte("# test"), 0644))

			// Should find the project directory
			got := autoMgr.findProjectDirectory(testDir)
			assert.Equal(t, testDir, got)
		})
	}
}

func TestDetermineAction(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	tests := []struct {
		name              string
		oldProjectDir     string
		newProjectDir     string
		currentProjectDir string
		want              ActivationAction
	}{
		{
			name:              "no project before or after",
			oldProjectDir:     "",
			newProjectDir:     "",
			currentProjectDir: "",
			want:              ActionNone,
		},
		{
			name:              "entering a project",
			oldProjectDir:     "",
			newProjectDir:     "/home/user/project",
			currentProjectDir: "",
			want:              ActionActivate,
		},
		{
			name:              "leaving a project",
			oldProjectDir:     "/home/user/project",
			newProjectDir:     "",
			currentProjectDir: "/home/user/project",
			want:              ActionDeactivate,
		},
		{
			name:              "same project",
			oldProjectDir:     "/home/user/project",
			newProjectDir:     "/home/user/project",
			currentProjectDir: "/home/user/project",
			want:              ActionNone,
		},
		{
			name:              "switching projects",
			oldProjectDir:     "/home/user/project1",
			newProjectDir:     "/home/user/project2",
			currentProjectDir: "/home/user/project1",
			want:              ActionSwitch,
		},
		{
			name:              "uses current project when old is empty",
			oldProjectDir:     "",
			newProjectDir:     "/home/user/project2",
			currentProjectDir: "/home/user/project1",
			want:              ActionSwitch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := autoMgr.determineAction(tt.oldProjectDir, tt.newProjectDir, tt.currentProjectDir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandleDirectoryChange_NoChange(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	ctx := context.Background()
	event := DirectoryChangeEvent{
		OldDir: "/home/user",
		NewDir: "/home/user/documents",
		Shell:  ShellBash,
	}
	currentState := &EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
	}

	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	require.NoError(t, err)
	assert.Equal(t, ActionNone, change.Action)
	assert.Empty(t, change.Script)
	assert.Equal(t, currentState, change.NewState)
}

func TestHandleDirectoryChange_Activate(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	configPath := filepath.Join(projectDir, "unirtm.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("# test"), 0644))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	ctx := context.Background()
	event := DirectoryChangeEvent{
		OldDir: tmpDir,
		NewDir: projectDir,
		Shell:  ShellBash,
	}
	currentState := &EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	require.NoError(t, err)
	assert.Equal(t, ActionActivate, change.Action)
	assert.NotEmpty(t, change.Script)
	assert.Equal(t, projectDir, change.NewState.ProjectDir)
	assert.Equal(t, "/usr/bin:/bin", change.NewState.PreviousPath)
}

func TestHandleDirectoryChange_Deactivate(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	configPath := filepath.Join(projectDir, "unirtm.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("# test"), 0644))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	ctx := context.Background()
	event := DirectoryChangeEvent{
		OldDir: projectDir,
		NewDir: tmpDir,
		Shell:  ShellBash,
	}
	currentState := &EnvironmentState{
		ProjectDir: projectDir,
		ToolVersions: map[string]string{
			"node": "20.0.0",
		},
		EnvVars: map[string]string{
			"NODE_ENV": "development",
		},
		PreviousPath: "/usr/bin:/bin",
	}

	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	require.NoError(t, err)
	assert.Equal(t, ActionDeactivate, change.Action)
	assert.NotEmpty(t, change.Script)
	assert.Empty(t, change.NewState.ProjectDir)
	assert.Empty(t, change.NewState.ToolVersions)
	assert.Empty(t, change.NewState.EnvVars)
}

func TestHandleDirectoryChange_Switch(t *testing.T) {
	// Create two temporary project directories
	tmpDir := t.TempDir()
	project1Dir := filepath.Join(tmpDir, "project1")
	project2Dir := filepath.Join(tmpDir, "project2")
	require.NoError(t, os.MkdirAll(project1Dir, 0755))
	require.NoError(t, os.MkdirAll(project2Dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(project1Dir, "unirtm.toml"), []byte("# test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(project2Dir, "unirtm.toml"), []byte("# test"), 0644))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	ctx := context.Background()
	event := DirectoryChangeEvent{
		OldDir: project1Dir,
		NewDir: project2Dir,
		Shell:  ShellBash,
	}
	currentState := &EnvironmentState{
		ProjectDir: project1Dir,
		ToolVersions: map[string]string{
			"node": "18.0.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	require.NoError(t, err)
	assert.Equal(t, ActionSwitch, change.Action)
	assert.NotEmpty(t, change.Script)
	assert.Contains(t, change.Script, "Deactivating previous project")
	assert.Contains(t, change.Script, "Activating new project")
	assert.Equal(t, project2Dir, change.NewState.ProjectDir)
}

func TestGenerateDeactivationScript_Bash(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	state := &EnvironmentState{
		ProjectDir: "/home/user/project",
		ToolVersions: map[string]string{
			"node":   "20.0.0",
			"python": "3.11.0",
		},
		EnvVars: map[string]string{
			"NODE_ENV": "development",
		},
		PreviousPath: "/usr/bin:/bin",
	}

	script := autoMgr.generateDeactivationScript(ShellBash, state)

	assert.Contains(t, script, "# UniRTM deactivation script")
	assert.Contains(t, script, "export PATH=\"/usr/bin:/bin\"")
	assert.Contains(t, script, "unset UNIRTM_NODE_VERSION")
	assert.Contains(t, script, "unset UNIRTM_PYTHON_VERSION")
	assert.Contains(t, script, "unset NODE_ENV")
	assert.Contains(t, script, "unset UNIRTM_ACTIVATION_SCOPE")
	assert.Contains(t, script, "unset UNIRTM_PROJECT_DIR")
}

func TestGenerateDeactivationScript_Fish(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	state := &EnvironmentState{
		ProjectDir: "/home/user/project",
		ToolVersions: map[string]string{
			"go": "1.21.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	script := autoMgr.generateDeactivationScript(ShellFish, state)

	assert.Contains(t, script, "# UniRTM deactivation script")
	assert.Contains(t, script, "set -gx PATH /usr/bin /bin")
	assert.Contains(t, script, "set -e UNIRTM_GO_VERSION")
	assert.Contains(t, script, "set -e UNIRTM_ACTIVATION_SCOPE")
	assert.Contains(t, script, "set -e UNIRTM_PROJECT_DIR")
}

func TestGenerateDeactivationScript_PowerShell(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	state := &EnvironmentState{
		ProjectDir: "C:\\Users\\user\\project",
		ToolVersions: map[string]string{
			"ruby": "3.2.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "C:\\Windows\\System32",
	}

	script := autoMgr.generateDeactivationScript(ShellPowerShell, state)

	assert.Contains(t, script, "# UniRTM deactivation script")
	assert.Contains(t, script, "$env:PATH = \"C:\\Windows\\System32\"")
	assert.Contains(t, script, "Remove-Item Env:UNIRTM_RUBY_VERSION")
	assert.Contains(t, script, "Remove-Item Env:UNIRTM_ACTIVATION_SCOPE")
	assert.Contains(t, script, "Remove-Item Env:UNIRTM_PROJECT_DIR")
}

func TestGenerateHookEnvScript_Bash(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	script, err := autoMgr.GenerateHookEnvScript(ShellBash, "/usr/local/bin/unirtm")
	require.NoError(t, err)

	assert.Contains(t, script, "# UniRTM auto-activation hook")
	assert.Contains(t, script, "_unirtm_hook()")
	assert.Contains(t, script, "UNIRTM_OLD_PWD")
	assert.Contains(t, script, "unirtm hook-env --shell bash")
	assert.Contains(t, script, "PROMPT_COMMAND")
}

func TestGenerateHookEnvScript_Zsh(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	script, err := autoMgr.GenerateHookEnvScript(ShellZsh, "/usr/local/bin/unirtm")
	require.NoError(t, err)

	assert.Contains(t, script, "# UniRTM auto-activation hook")
	assert.Contains(t, script, "_unirtm_hook()")
	assert.Contains(t, script, "UNIRTM_OLD_PWD")
	assert.Contains(t, script, "unirtm hook-env --shell zsh")
	assert.Contains(t, script, "add-zsh-hook precmd")
}

func TestGenerateHookEnvScript_Fish(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	script, err := autoMgr.GenerateHookEnvScript(ShellFish, "/usr/local/bin/unirtm")
	require.NoError(t, err)

	assert.Contains(t, script, "# UniRTM auto-activation hook for fish")
	assert.Contains(t, script, "function _unirtm_hook --on-variable PWD")
	assert.Contains(t, script, "unirtm hook-env --shell fish")
}

func TestGenerateHookEnvScript_PowerShell(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	script, err := autoMgr.GenerateHookEnvScript(ShellPowerShell, "/usr/local/bin/unirtm")
	require.NoError(t, err)

	assert.Contains(t, script, "# UniRTM auto-activation hook for PowerShell")
	assert.Contains(t, script, "function Invoke-UnirtmHook")
	assert.Contains(t, script, "$env:UNIRTM_OLD_PWD")
	assert.Contains(t, script, "unirtm hook-env --shell powershell")
	assert.Contains(t, script, "function prompt")
}

func TestGenerateHookEnvScript_UnsupportedShell(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	_, err := autoMgr.GenerateHookEnvScript("unsupported", "/usr/local/bin/unirtm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported shell type")
}

func TestEnvironmentState_SavesPath(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	configPath := filepath.Join(projectDir, "unirtm.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("# test"), 0644))

	ctx := context.Background()
	event := DirectoryChangeEvent{
		OldDir: tmpDir,
		NewDir: projectDir,
		Shell:  ShellBash,
	}

	// First activation - should save PATH
	currentState := &EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
		PreviousPath: "",
	}

	change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
	require.NoError(t, err)
	assert.NotEmpty(t, change.NewState.PreviousPath)

	// Second activation - should preserve saved PATH
	event2 := DirectoryChangeEvent{
		OldDir: projectDir,
		NewDir: filepath.Join(projectDir, "subdir"),
		Shell:  ShellBash,
	}

	savedPath := change.NewState.PreviousPath
	change2, err := autoMgr.HandleDirectoryChange(ctx, event2, change.NewState)
	require.NoError(t, err)
	assert.Equal(t, savedPath, change2.NewState.PreviousPath)
}

func TestFindProjectDirectory_SymlinkHandling(t *testing.T) {
	// Skip on Windows as symlink creation requires admin privileges
	if env.Get("GOOS") == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	linkDir := filepath.Join(tmpDir, "link")

	require.NoError(t, os.MkdirAll(realDir, 0755))
	configPath := filepath.Join(realDir, "unirtm.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("# test"), 0644))

	// Create symlink
	err := os.Symlink(realDir, linkDir)
	if err != nil {
		t.Skipf("Cannot create symlink: %v", err)
	}

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	// Should find project through symlink
	got := autoMgr.findProjectDirectory(linkDir)
	assert.NotEmpty(t, got)
}

func TestGenerateDeactivationScript_EmptyState(t *testing.T) {
	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	state := &EnvironmentState{
		ProjectDir:   "",
		ToolVersions: make(map[string]string),
		EnvVars:      make(map[string]string),
		PreviousPath: "",
	}

	script := autoMgr.generateDeactivationScript(ShellBash, state)

	// Should still unset UniRTM variables even with empty state
	assert.Contains(t, script, "unset UNIRTM_ACTIVATION_SCOPE")
	assert.Contains(t, script, "unset UNIRTM_PROJECT_DIR")
	// Should not try to restore PATH if it wasn't saved
	assert.NotContains(t, script, "export PATH=\"\"")
}

func TestActivationAction_String(t *testing.T) {
	tests := []struct {
		action ActivationAction
		want   string
	}{
		{ActionActivate, "activate"},
		{ActionDeactivate, "deactivate"},
		{ActionSwitch, "switch"},
		{ActionNone, "none"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.action))
		})
	}
}

func TestFindProjectDirectory_NestedProjects(t *testing.T) {
	// Test that we find the nearest project, not a parent project
	tmpDir := t.TempDir()

	// Create outer project
	outerProject := filepath.Join(tmpDir, "outer")
	require.NoError(t, os.MkdirAll(outerProject, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(outerProject, "unirtm.toml"), []byte("# outer"), 0644))

	// Create inner project
	innerProject := filepath.Join(outerProject, "inner")
	require.NoError(t, os.MkdirAll(innerProject, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(innerProject, "unirtm.toml"), []byte("# inner"), 0644))

	// Create subdirectory in inner project
	subDir := filepath.Join(innerProject, "src")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	// Should find inner project, not outer
	got := autoMgr.findProjectDirectory(subDir)
	assert.Equal(t, innerProject, got)
}

func TestGenerateSwitch_CombinesScripts(t *testing.T) {
	// Create two temporary project directories
	tmpDir := t.TempDir()
	project1Dir := filepath.Join(tmpDir, "project1")
	project2Dir := filepath.Join(tmpDir, "project2")
	require.NoError(t, os.MkdirAll(project1Dir, 0755))
	require.NoError(t, os.MkdirAll(project2Dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(project1Dir, "unirtm.toml"), []byte("# test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(project2Dir, "unirtm.toml"), []byte("# test"), 0644))

	activationMgr := NewActivationManager("/tmp/shims", "/tmp/data", provider.NewRegistry())
	autoMgr := NewAutoActivationManager(activationMgr)

	ctx := context.Background()
	currentState := &EnvironmentState{
		ProjectDir: project1Dir,
		ToolVersions: map[string]string{
			"node": "18.0.0",
		},
		EnvVars:      make(map[string]string),
		PreviousPath: "/usr/bin:/bin",
	}

	change, err := autoMgr.generateSwitch(ctx, ShellBash, project1Dir, project2Dir, currentState)
	require.NoError(t, err)

	// Verify the script contains both deactivation and activation
	lines := strings.Split(change.Script, "\n")
	var foundDeactivation, foundActivation bool
	for _, line := range lines {
		if strings.Contains(line, "Deactivating previous project") {
			foundDeactivation = true
		}
		if strings.Contains(line, "Activating new project") {
			foundActivation = true
		}
	}

	assert.True(t, foundDeactivation, "Script should contain deactivation comment")
	assert.True(t, foundActivation, "Script should contain activation comment")
	assert.Equal(t, project2Dir, change.NewState.ProjectDir)
}
