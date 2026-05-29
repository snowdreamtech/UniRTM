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

func TestEnableCommandStructure(t *testing.T) {
	assert.Equal(t, "enable [unirtm|mise]", enableCmd.Use)
	assert.NotEmpty(t, enableCmd.Short)
	assert.NotNil(t, enableCmd.RunE)
}

func TestRunEnable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a dummy shell config
	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte("# init\n"), 0644)
	require.NoError(t, err)

	cmd := enableCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// We'll set the shell env for DetectShell
	t.Setenv("SHELL", "/bin/zsh")

	err = runEnable(cmd, []string{"unirtm"})
	assert.NoError(t, err)
}

func TestRunEnable_All(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte("# init\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, ".bashrc"), []byte("# init\n"), 0644)
	require.NoError(t, err)

	enableAll = true
	defer func() { enableAll = false }()

	cmd := enableCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runEnable(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunEnable_InvalidTool(t *testing.T) {
	cmd := enableCmd
	err := runEnable(cmd, []string{"invalid"})
	assert.Error(t, err)
}
