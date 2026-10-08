// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/service"
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
	t.Setenv("USERPROFILE", tmpDir)

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
	t.Setenv("USERPROFILE", tmpDir)

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

func TestGetActivationCmd_CrossPlatform(t *testing.T) {
	// 1. Bash: Must never contain unescaped backslashes
	cmdBash, err := getActivationCmd("unirtm", service.ShellBash, false)
	require.NoError(t, err)
	assert.Contains(t, cmdBash, `activate bash)"`)
	assert.NotContains(t, cmdBash, `\`)

	// 2. PowerShell: Must be safe for execution
	cmdPS, err := getActivationCmd("unirtm", service.ShellPowerShell, false)
	require.NoError(t, err)
	assert.Contains(t, cmdPS, `activate powershell | Out-String | Invoke-Expression`)

	// 3. Zsh: Must never contain unescaped backslashes
	cmdZsh, err := getActivationCmd("unirtm", service.ShellZsh, false)
	require.NoError(t, err)
	assert.Contains(t, cmdZsh, `activate zsh)"`)
	assert.NotContains(t, cmdZsh, `\`)
}
