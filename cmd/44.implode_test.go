// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImplodeStructure(t *testing.T) {
	assert.Contains(t, implodeCmd.Use, "implode", "implodeCmd command use should contain 'implode'")
	assert.NotEmpty(t, implodeCmd.Short, "implodeCmd command short description should not be empty")
	assert.True(t, implodeCmd.Run != nil || implodeCmd.RunE != nil, "Run or RunE function should be set for implodeCmd")
}

func TestRunImplode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := implodeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	implodeYes = true
	err := runImplode(cmd, []string{})
	assert.NoError(t, err)
}
