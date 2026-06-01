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

func TestEditCommandStructure(t *testing.T) {
	assert.Equal(t, "edit [FILE]", editCmd.Use)
	assert.NotEmpty(t, editCmd.Short)
	assert.NotNil(t, editCmd.RunE)
}

func TestDiscoverConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some dummy config files
	err := os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(""), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "mise.toml"), []byte(""), 0644)
	require.NoError(t, err)

	// Change working dir for discoverConfigFiles
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	candidates := discoverConfigFiles()

	// Since we created 2 local files, we expect at least 2
	assert.GreaterOrEqual(t, len(candidates), 2)

	foundUnirtm := false
	foundMise := false
	for _, c := range candidates {
		if c.Name == ".unirtm.toml" {
			foundUnirtm = true
		}
		if c.Name == "mise.toml" {
			foundMise = true
		}
	}
	assert.True(t, foundUnirtm)
	assert.True(t, foundMise)
}

func TestRunEdit_SpecificFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Create a non-blocking cross-platform dummy editor
	var dummyEditor string
	if runtime.GOOS == "windows" {
		dummyEditor = filepath.Join(tmpDir, "dummy_editor.bat")
		_ = os.WriteFile(dummyEditor, []byte("@echo off\nexit 0"), 0755)
	} else {
		dummyEditor = filepath.Join(tmpDir, "dummy_editor.sh")
		_ = os.WriteFile(dummyEditor, []byte("#!/bin/sh\nexit 0"), 0755)
	}
	os.Setenv("UNIRTM_EDITOR", dummyEditor)

	targetFile := filepath.Join(tmpDir, "unirtm.toml")

	cmd := editCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEdit(cmd, []string{targetFile})
	assert.NoError(t, err)

	// The file should have been created since it didn't exist
	_, err = os.Stat(targetFile)
	assert.NoError(t, err)
}

func TestRunEdit_GlobalFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	var dummyEditor string
	if runtime.GOOS == "windows" {
		dummyEditor = filepath.Join(tmpDir, "dummy_editor.bat")
		_ = os.WriteFile(dummyEditor, []byte("@echo off\nexit 0"), 0755)
	} else {
		dummyEditor = filepath.Join(tmpDir, "dummy_editor.sh")
		_ = os.WriteFile(dummyEditor, []byte("#!/bin/sh\nexit 0"), 0755)
	}
	t.Setenv("UNIRTM_EDITOR", dummyEditor)

	editGlobal = true
	defer func() { editGlobal = false }()

	cmd := editCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runEdit(cmd, []string{})
	assert.NoError(t, err)

	// It should create the global config
	globalPath := filepath.Join(tmpDir, "unirtm.toml")
	_, err = os.Stat(globalPath)
	assert.NoError(t, err)
}
