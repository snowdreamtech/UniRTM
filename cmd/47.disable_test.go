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

func TestDisableCommandStructure(t *testing.T) {
	assert.Equal(t, "disable [unirtm|mise]", disableCmd.Use)
	assert.NotEmpty(t, disableCmd.Short)
	assert.NotNil(t, disableCmd.RunE)
}

func TestRunDisable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a dummy shell config
	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte("# init\n# unirtm config here\n"), 0644)
	require.NoError(t, err)

	cmd := disableCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// We'll set the shell env for DetectShell
	t.Setenv("SHELL", "/bin/zsh")

	err = runDisable(cmd, []string{"unirtm"})
	assert.NoError(t, err)
}

func TestRunDisable_All(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte("# init\n# unirtm config here\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, ".bashrc"), []byte("# init\n# unirtm config here\n"), 0644)
	require.NoError(t, err)

	disableAll = true
	defer func() { disableAll = false }()

	cmd := disableCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runDisable(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunDisable_InvalidTool(t *testing.T) {
	cmd := disableCmd
	err := runDisable(cmd, []string{"invalid"})
	assert.Error(t, err)
}
