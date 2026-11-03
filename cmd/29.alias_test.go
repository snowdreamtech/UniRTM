// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliasStructure(t *testing.T) {
	assert.Contains(t, aliasCmd.Use, "alias", "aliasCmd command use should contain 'alias'")
	assert.NotEmpty(t, aliasCmd.Short, "aliasCmd command short description should not be empty")
	assert.True(t, aliasCmd.Run != nil || aliasCmd.RunE != nil, "Run or RunE function should be set for aliasCmd")
}

func TestRunAlias(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := aliasCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err) // It might fail with wrong args, that's fine
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
