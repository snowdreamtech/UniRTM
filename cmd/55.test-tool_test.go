// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestExecutable_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy executable that always succeeds
	exePath := filepath.Join(tmpDir, "dummy")
	script := `#!/bin/sh
exit 0
`
	err := os.WriteFile(exePath, []byte(script), 0755)
	assert.NoError(t, err)

	err = testExecutable(exePath, os.Environ())
	assert.NoError(t, err)
}

func TestTestExecutable_Failure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy executable that always fails
	exePath := filepath.Join(tmpDir, "dummy_fail")
	script := `#!/bin/sh
echo "some error output"
exit 1
`
	err := os.WriteFile(exePath, []byte(script), 0755)
	assert.NoError(t, err)

	err = testExecutable(exePath, os.Environ())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some error output")
}

// Test RunTestTool with no tools
func TestRunTestTool_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")

	// Empty config
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(""), 0644)

	err := runTestTool(testToolCmd, []string{})
	assert.NoError(t, err)
}
