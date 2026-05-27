// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncStructure(t *testing.T) {
	assert.Contains(t, syncCmd.Use, "sync", "syncCmd command use should contain 'sync'")
	assert.NotEmpty(t, syncCmd.Short, "syncCmd command short description should not be empty")
	assert.True(t, syncCmd.Run != nil || syncCmd.RunE != nil, "Run or RunE function should be set for syncCmd")
}

func TestRunSync(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := syncCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runSync(cmd, []string{})
	assert.NoError(t, err)
}
