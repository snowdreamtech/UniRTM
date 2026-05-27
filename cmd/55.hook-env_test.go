// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookEnvStructure(t *testing.T) {
	assert.Contains(t, hookEnvCmd.Use, "hook-env", "hookEnvCmd command use should contain 'hook-env'")
	assert.NotEmpty(t, hookEnvCmd.Short, "hookEnvCmd command short description should not be empty")
	assert.True(t, hookEnvCmd.Run != nil || hookEnvCmd.RunE != nil, "Run or RunE function should be set for hookEnvCmd")
}

func TestRunHookEnv(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := hookEnvCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runHookEnv(cmd, []string{})
	assert.NoError(t, err)
}
