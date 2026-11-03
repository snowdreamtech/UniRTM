// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestExecutable_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy executable that always succeeds
	var script, ext string
	if runtime.GOOS == "windows" {
		ext = ".bat"
		script = "@echo off\r\nexit 0\r\n"
	} else {
		script = "#!/bin/sh\nexit 0\n"
	}
	exePath := filepath.Join(tmpDir, "dummy"+ext)
	err := os.WriteFile(exePath, []byte(script), 0755)
	assert.NoError(t, err)

	err = testExecutable(exePath, os.Environ())
	assert.NoError(t, err)
}

func TestTestExecutable_Failure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy executable that always fails
	var script, ext string
	if runtime.GOOS == "windows" {
		ext = ".bat"
		script = "@echo off\r\necho some error output\r\nexit 1\r\n"
	} else {
		script = "#!/bin/sh\necho \"some error output\"\nexit 1\n"
	}
	exePath := filepath.Join(tmpDir, "dummy_fail"+ext)
	err := os.WriteFile(exePath, []byte(script), 0755)
	assert.NoError(t, err)

	err = testExecutable(exePath, os.Environ())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some error output")
}

// Test RunTestTool with no tools
func TestRunTestTool_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	// Empty config
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(""), 0644)

	err := runTestTool(testToolCmd, []string{})
	assert.NoError(t, err)
}
