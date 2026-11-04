// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookCmd_Structure(t *testing.T) {
	assert.Equal(t, "hook", hookCmd.Use)
	assert.NotEmpty(t, hookCmd.Short)

	// Check subcommands
	subcommands := hookCmd.Commands()
	subNames := make([]string, 0, len(subcommands))
	for _, sc := range subcommands {
		subNames = append(subNames, sc.Name())
	}
	assert.Contains(t, subNames, "install")
	assert.Contains(t, subNames, "uninstall")
	assert.Contains(t, subNames, "run")
}

func TestHookInstallCmd_GlobalFlag(t *testing.T) {
	flag := hookInstallCmd.Flags().Lookup("global")
	assert.NotNil(t, flag)
	assert.Equal(t, "g", flag.Shorthand)
}

func TestHookInstallCmd_Execution_Global(t *testing.T) {
	tmpConfig := t.TempDir()
	t.Setenv("UNIRTM_CONFIG_DIR", tmpConfig)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(tmpConfig, "gitconfig"))

	hookGlobal = true
	defer func() { hookGlobal = false }()

	err := hookInstallCmd.RunE(hookInstallCmd, []string{})
	assert.NoError(t, err)

	// Verify global hooks directory was populated
	preCommitPath := filepath.Join(tmpConfig, "hooks", "pre-commit")
	fi, err := os.Stat(preCommitPath)
	assert.NoError(t, err)
	assert.False(t, fi.IsDir())

	// Test uninstall global
	err = hookUninstallCmd.RunE(hookUninstallCmd, []string{})
	assert.NoError(t, err)
}
