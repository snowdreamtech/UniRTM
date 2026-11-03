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

func TestConfigStructure(t *testing.T) {
	assert.Contains(t, configCmd.Use, "config", "configCmd command use should contain 'config'")
	assert.NotEmpty(t, configCmd.Short, "configCmd command short description should not be empty")
	assert.True(t, configCmd.Run != nil || configCmd.RunE != nil, "Run or RunE function should be set for configCmd")
}

func TestRunConfigGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "unirtm.toml")

	cmd := configGenerateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runConfigGenerate(cmd, []string{targetFile})
	assert.NoError(t, err)

	_, err = os.Stat(targetFile)
	assert.NoError(t, err)

	// test without args (should write to .unirtm.toml in cwd)
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	err = runConfigGenerate(cmd, []string{})
	assert.NoError(t, err)

	_, err = os.Stat(".unirtm.toml")
	assert.NoError(t, err)
}

func TestRunConfigSetGet(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	// Set config Path for the runConfigSet
	configPath = filepath.Join(tmpDir, ".unirtm.toml")
	defer func() { configPath = "" }()

	cmdSet := configSetCmd
	var buf bytes.Buffer
	cmdSet.SetOut(&buf)

	err := runConfigSet(cmdSet, []string{"settings.cache_ttl", "48h"})
	assert.NoError(t, err)

	cmdGet := configGetCmd
	cmdGet.SetOut(&buf)
	err = runConfigGet(cmdGet, []string{"settings.cache_ttl"})
	assert.NoError(t, err)
}

func TestRunConfigValidate(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	targetFile := filepath.Join(tmpDir, "unirtm.toml")
	err := os.WriteFile(targetFile, []byte("[tools]\ndummy = \"1.0.0\"\n"), 0644)
	assert.NoError(t, err)

	cmd := configValidateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runConfigValidate(cmd, []string{targetFile})
	assert.NoError(t, err)

	err = runConfigValidate(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunConfigShow(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	targetFile := filepath.Join(tmpDir, ".unirtm.toml")
	err := os.WriteFile(targetFile, []byte("[tools]\ndummy = \"1.0.0\"\n[env]\nMY_ENV = \"test\"\n[settings]\ncache_ttl = \"168h\""), 0644)
	assert.NoError(t, err)

	cmd := configShowCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runConfigShow(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunConfigJson(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	targetFile := filepath.Join(tmpDir, ".unirtm.toml")
	err := os.WriteFile(targetFile, []byte("[tools]\ndummy = \"1.0.0\"\n[env]\nMY_ENV = \"test\"\n[settings]\ncache_ttl = \"168h\""), 0644)
	assert.NoError(t, err)

	jsonOutput = true
	defer func() { jsonOutput = false }()

	cmdShow := configShowCmd
	var buf bytes.Buffer
	cmdShow.SetOut(&buf)
	err = runConfigShow(cmdShow, []string{})
	assert.NoError(t, err)

	cmdValidate := configValidateCmd
	cmdValidate.SetOut(&buf)
	err = runConfigValidate(cmdValidate, []string{})
	assert.NoError(t, err)

	cmdGet := configGetCmd
	cmdGet.SetOut(&buf)
	err = runConfigGet(cmdGet, []string{"tools"})
	assert.NoError(t, err)
}

func TestConfigGlobalBrowser(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	// 1. Setting browser locally (without --global) must return error
	configGlobal = false
	err := runConfigSet(configSetCmd, []string{"settings.browser", "chrome"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only allowed globally")

	// 2. Setting browser with --global must succeed
	configGlobal = true
	defer func() { configGlobal = false }()
	err = runConfigSet(configSetCmd, []string{"settings.browser", "chrome"})
	assert.NoError(t, err)

	// 3. Getting browser with --global must succeed
	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err = runConfigGet(configGetCmd, []string{"settings.browser"})
	assert.NoError(t, err)

	// 4. Test root configCmd routing: unirtm config --global settings.browser firefox
	err = configCmd.RunE(configCmd, []string{"settings.browser", "firefox"})
	assert.NoError(t, err)
}

func TestPluginConfigNamespaceIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	configPath = filepath.Join(tmpDir, ".unirtm.toml")
	defer func() { configPath = "" }()

	// 1. Unreserved key must be rejected
	err := runConfigSet(configSetCmd, []string{"myplugin.setting", "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not reserved for core")

	// 2. Setting Plugin A config succeeds
	err = runConfigSet(configSetCmd, []string{`plugins."snowdreamtech/docker".config.port`, "8080"})
	assert.NoError(t, err)

	// 3. Setting Plugin B config succeeds and does not clobber Plugin A
	err = runConfigSet(configSetCmd, []string{`plugins."community/deploy".config.cluster`, "staging"})
	assert.NoError(t, err)

	// 4. Reading back Plugin A config
	var buf bytes.Buffer
	configGetCmd.SetOut(&buf)
	err = runConfigGet(configGetCmd, []string{`plugins."snowdreamtech/docker".config.port`})
	assert.NoError(t, err)

	// 5. Reading back Plugin B config
	err = runConfigGet(configGetCmd, []string{`plugins."community/deploy".config.cluster`})
	assert.NoError(t, err)

	// 6. Config show renders without error
	cmdShow := configShowCmd
	cmdShow.SetOut(&buf)
	err = runConfigShow(cmdShow, []string{})
	assert.NoError(t, err)
}

func TestPluginManifestSchemaValidation(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0755))

	manifestContent := `{
  "name": "snowdreamtech/docker",
  "version": "1.0.0",
  "contributes": {
    "configuration": {
      "port": {
        "type": "integer"
      },
      "driver": {
        "type": "string",
        "enum": ["overlay2", "btrfs"]
      }
    }
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "snowdreamtech-docker.manifest.json"), []byte(manifestContent), 0644))

	configPath = filepath.Join(tmpDir, ".unirtm.toml")
	defer func() { configPath = "" }()

	// 1. Valid integer port
	err := runConfigSet(configSetCmd, []string{`plugins."snowdreamtech/docker".config.port`, "9090"})
	assert.NoError(t, err)

	// 2. Invalid non-integer port rejected
	err = runConfigSet(configSetCmd, []string{`plugins."snowdreamtech/docker".config.port`, "not_a_number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")

	// 3. Valid enum
	err = runConfigSet(configSetCmd, []string{`plugins."snowdreamtech/docker".config.driver`, "overlay2"})
	assert.NoError(t, err)

	// 4. Invalid enum rejected
	err = runConfigSet(configSetCmd, []string{`plugins."snowdreamtech/docker".config.driver`, "unsupported_driver"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowed enums")
}
