// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommandStructure(t *testing.T) {
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)
	assert.NotEmpty(t, completionCmd.Short)
	assert.NotNil(t, completionCmd.RunE)
}

// TestRunCompletion_Generate verifies that a single shell's completion is printed to stdout.
func TestRunCompletion_Generate(t *testing.T) {
	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{"bash"})
	assert.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}

// TestRunCompletion_Install verifies that a single shell installs its completion file.
func TestRunCompletion_Install(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Create dummy shell config to allow source injection.
	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte(""), 0644)
	require.NoError(t, err)

	completionInstall = true
	defer func() { completionInstall = false }()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runCompletion(cmd, []string{"zsh"})
	assert.NoError(t, err)

	// Check that the completion file was created.
	compFile := filepath.Join(tmpDir, "completions", "unirtm.zsh")
	_, err = os.Stat(compFile)
	assert.NoError(t, err)
}

// TestRunCompletion_Uninstall verifies that a single shell's completion is removed.
func TestRunCompletion_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Create dummy completion file.
	compDir := filepath.Join(tmpDir, "completions")
	err := os.MkdirAll(compDir, 0755)
	require.NoError(t, err)
	compFile := filepath.Join(compDir, "unirtm.zsh")
	err = os.WriteFile(compFile, []byte(""), 0644)
	require.NoError(t, err)

	completionUninstall = true
	defer func() { completionUninstall = false }()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runCompletion(cmd, []string{"zsh"})
	assert.NoError(t, err)

	// Verify the completion file was removed.
	_, err = os.Stat(compFile)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

// TestRunCompletion_All_Install verifies that --all -i generates all 4 scripts and only
// injects source for shells with existing config files.
func TestRunCompletion_All_Install(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Only create zsh config; bash/fish/powershell configs are absent.
	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte(""), 0644)
	require.NoError(t, err)

	completionAll = true
	completionInstall = true
	defer func() {
		completionAll = false
		completionInstall = false
	}()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runCompletion(cmd, []string{})
	assert.NoError(t, err)

	compDir := filepath.Join(tmpDir, "completions")

	// All 4 completion scripts must be generated regardless of the environment.
	for _, filename := range []string{"unirtm.zsh", "unirtm.bash", "unirtm.ps1"} {
		_, statErr := os.Stat(filepath.Join(compDir, filename))
		assert.NoError(t, statErr, "completion script %s should always be generated", filename)
	}

	// Zsh config should have the activation line injected.
	zshrc, err := os.ReadFile(filepath.Join(tmpDir, ".zshrc"))
	require.NoError(t, err)
	assert.Contains(t, string(zshrc), "unirtm.zsh", "zshrc should contain completion source line")

	// Bash config should NOT be created automatically (it was absent).
	_, err = os.Stat(filepath.Join(tmpDir, ".bashrc"))
	assert.True(t, os.IsNotExist(err), ".bashrc should NOT be auto-created by --all install")
}

// TestRunCompletion_All_Uninstall verifies that --all -u removes completion for all shells.
func TestRunCompletion_All_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	// Pre-create completion files for all shells.
	compDir := filepath.Join(tmpDir, "completions")
	require.NoError(t, os.MkdirAll(compDir, 0755))
	for _, f := range []string{"unirtm.zsh", "unirtm.bash", "unirtm.ps1"} {
		require.NoError(t, os.WriteFile(filepath.Join(compDir, f), []byte(""), 0644))
	}

	completionAll = true
	completionUninstall = true
	defer func() {
		completionAll = false
		completionUninstall = false
	}()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{})
	assert.NoError(t, err)

	// All completion files should be removed.
	for _, f := range []string{"unirtm.zsh", "unirtm.bash", "unirtm.ps1"} {
		_, statErr := os.Stat(filepath.Join(compDir, f))
		assert.True(t, os.IsNotExist(statErr), "%s should be removed after uninstall", f)
	}
}

// TestRunCompletion_Dir exports all four completion scripts to a specified directory.
func TestRunCompletion_Dir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	destDir := filepath.Join(tmpDir, "exported-completions")

	completionDir = destDir
	defer func() { completionDir = "" }()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{})
	assert.NoError(t, err)

	// All 4 scripts must exist in the destination directory.
	for _, filename := range []string{"unirtm.zsh", "unirtm.bash", "unirtm.fish", "unirtm.ps1"} {
		info, statErr := os.Stat(filepath.Join(destDir, filename))
		assert.NoError(t, statErr, "%s should be exported to %s", filename, destDir)
		if info != nil {
			assert.Greater(t, info.Size(), int64(0), "%s should not be empty", filename)
		}
	}
}

// TestRunCompletion_Dir_DryRun verifies dry-run mode for the -d flag.
func TestRunCompletion_Dir_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dry-run-completions")

	completionDir = destDir
	dryRun = true
	defer func() {
		completionDir = ""
		dryRun = false
	}()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{})
	assert.NoError(t, err)

	// Directory should NOT be created in dry-run mode.
	_, err = os.Stat(destDir)
	assert.True(t, os.IsNotExist(err), "directory should not be created in dry-run mode")
}

// TestRunCompletion_Dir_MutuallyExclusive_Install verifies that -d and -i cannot be combined.
func TestRunCompletion_Dir_MutuallyExclusive_Install(t *testing.T) {
	tmpDir := t.TempDir()
	completionDir = filepath.Join(tmpDir, "out")
	completionInstall = true
	defer func() {
		completionDir = ""
		completionInstall = false
	}()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestRunCompletion_Dir_MutuallyExclusive_Uninstall verifies that -d and -u cannot be combined.
func TestRunCompletion_Dir_MutuallyExclusive_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	completionDir = filepath.Join(tmpDir, "out")
	completionUninstall = true
	defer func() {
		completionDir = ""
		completionUninstall = false
	}()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestAllShells_Coverage ensures allShells and completionFileNames are consistent.
func TestAllShells_Coverage(t *testing.T) {
	assert.Len(t, allShells, 4, "allShells should list exactly 4 supported shells")
	for _, st := range allShells {
		filename, ok := completionFileNames[st]
		assert.True(t, ok, "completionFileNames should have an entry for %s", st)
		assert.NotEmpty(t, filename)
	}
}
