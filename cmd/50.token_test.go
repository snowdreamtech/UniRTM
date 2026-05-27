// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenStructure(t *testing.T) {
	assert.Contains(t, tokenCmd.Use, "token", "tokenCmd command use should contain 'token'")
	assert.NotEmpty(t, tokenCmd.Short, "tokenCmd command short description should not be empty")
	assert.True(t, tokenCmd.Run != nil || tokenCmd.RunE != nil, "Run or RunE function should be set for tokenCmd")
}

func TestRunToken(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := tokenCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runToken(cmd, []string{})
	assert.NoError(t, err)
}
