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
