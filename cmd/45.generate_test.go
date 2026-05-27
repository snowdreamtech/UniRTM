// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStructure(t *testing.T) {
	assert.Contains(t, generateCmd.Use, "generate", "generateCmd command use should contain 'generate'")
	assert.NotEmpty(t, generateCmd.Short, "generateCmd command short description should not be empty")
	assert.True(t, generateCmd.Run != nil || generateCmd.RunE != nil, "Run or RunE function should be set for generateCmd")
}

func TestRunGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := generateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{})
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
