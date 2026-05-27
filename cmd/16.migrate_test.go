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

func TestMigrateStructure(t *testing.T) {
	assert.Contains(t, migrateCmd.Use, "migrate", "migrateCmd command use should contain 'migrate'")
	assert.NotEmpty(t, migrateCmd.Short, "migrateCmd command short description should not be empty")
	assert.True(t, migrateCmd.Run != nil || migrateCmd.RunE != nil, "Run or RunE function should be set for migrateCmd")
}

func TestRunMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := migrateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// test no args, no files -> should fail
	err := runMigrate(cmd, []string{})
	assert.Error(t, err)

	// test with valid file
	testFile := filepath.Join(tmpDir, ".tool-versions")
	err = os.WriteFile(testFile, []byte("nodejs 18.0.0\n"), 0644)
	assert.NoError(t, err)

	migrateDryRun = true
	migrateOutput = filepath.Join(tmpDir, ".unirtm.toml")
	err = runMigrate(cmd, []string{testFile})
	assert.NoError(t, err)

	migrateDryRun = false
	err = runMigrate(cmd, []string{testFile})
	assert.NoError(t, err)

	jsonOutput = true
	defer func() { jsonOutput = false }()
	err = runMigrate(cmd, []string{testFile})
	assert.NoError(t, err)
}
