// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginStructure(t *testing.T) {
	assert.Contains(t, pluginCmd.Use, "plugin", "pluginCmd command use should contain 'plugin'")
	assert.NotEmpty(t, pluginCmd.Short, "pluginCmd command short description should not be empty")
	assert.True(t, pluginCmd.Run != nil || pluginCmd.RunE != nil, "Run or RunE function should be set for pluginCmd")
}

func TestRunPluginList(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := pluginListCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPluginList(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunPluginInstallAndRemove(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	srcFile := filepath.Join(tmpDir, "unirtm-plugin-dummy")
	err := os.WriteFile(srcFile, []byte("dummy plugin"), 0755)
	assert.NoError(t, err)

	cmdInstall := pluginInstallCmd
	var buf bytes.Buffer
	cmdInstall.SetOut(&buf)

	err = runPluginInstall(cmdInstall, []string{srcFile})
	assert.NoError(t, err)

	// List again
	cmdList := pluginListCmd
	cmdList.SetOut(&buf)
	err = runPluginList(cmdList, []string{})
	assert.NoError(t, err)

	// Remove
	cmdRemove := pluginRemoveCmd
	cmdRemove.SetOut(&buf)
	err = runPluginRemove(cmdRemove, []string{"dummy"})
	assert.NoError(t, err)

	// Remove again to test error
	err = runPluginRemove(cmdRemove, []string{"dummy"})
	assert.Error(t, err)
}

func TestRunPluginInstall_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	srcFile := filepath.Join(tmpDir, "unirtm-plugin-dummy")
	err := os.WriteFile(srcFile, []byte("dummy plugin"), 0755)
	assert.NoError(t, err)

	dryRun = true
	defer func() { dryRun = false }()

	cmdInstall := pluginInstallCmd
	var buf bytes.Buffer
	cmdInstall.SetOut(&buf)

	err = runPluginInstall(cmdInstall, []string{srcFile})
	assert.NoError(t, err)

	cmdRemove := pluginRemoveCmd
	cmdRemove.SetOut(&buf)
	err = runPluginRemove(cmdRemove, []string{"dummy"})
	assert.NoError(t, err)
}

func TestRunPluginList_Json(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	jsonOutput = true
	defer func() { jsonOutput = false }()

	cmd := pluginListCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPluginList(cmd, []string{})
	assert.NoError(t, err)
}
