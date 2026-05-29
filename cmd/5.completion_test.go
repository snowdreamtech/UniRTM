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

func TestRunCompletion_Generate(t *testing.T) {
	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCompletion(cmd, []string{"bash"})
	assert.NoError(t, err)
}

func TestRunCompletion_Install(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create dummy shell config
	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte(""), 0644)
	require.NoError(t, err)

	completionInstall = true
	defer func() { completionInstall = false }()

	cmd := completionCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runCompletion(cmd, []string{"zsh"})
	assert.NoError(t, err)

	// Check completion file was created
	compFile := filepath.Join(tmpDir, "completions", "unirtm.zsh")
	_, err = os.Stat(compFile)
	assert.NoError(t, err)
}

func TestRunCompletion_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create dummy completion file
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

	// Check completion file was removed
	_, err = os.Stat(compFile)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestRunCompletion_All(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Create dummy shell config
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

	// Check completion file was created
	compFile := filepath.Join(tmpDir, "completions", "unirtm.zsh")
	_, err = os.Stat(compFile)
	assert.NoError(t, err)
}
