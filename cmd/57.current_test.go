// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentStructure(t *testing.T) {
	assert.Contains(t, currentCmd.Use, "current", "currentCmd command use should contain 'current'")
	assert.NotEmpty(t, currentCmd.Short, "currentCmd command short description should not be empty")
	assert.True(t, currentCmd.Run != nil || currentCmd.RunE != nil, "Run or RunE function should be set for currentCmd")
}

func TestRunCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := currentCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCurrent(cmd, []string{})
	assert.NoError(t, err)
}
