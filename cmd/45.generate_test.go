// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
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
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := generateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{})
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
