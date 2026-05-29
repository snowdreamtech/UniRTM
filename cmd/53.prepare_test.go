// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareStructure(t *testing.T) {
	assert.Contains(t, prepareCmd.Use, "prepare", "prepareCmd command use should contain 'prepare'")
	assert.NotEmpty(t, prepareCmd.Short, "prepareCmd command short description should not be empty")
	assert.True(t, prepareCmd.Run != nil || prepareCmd.RunE != nil, "Run or RunE function should be set for prepareCmd")
}

func TestRunPrepare(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := prepareCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runPrepare(cmd, []string{})
	assert.NoError(t, err)
}
