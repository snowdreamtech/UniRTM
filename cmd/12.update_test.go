// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateStructure(t *testing.T) {
	assert.Contains(t, updateCmd.Use, "update", "updateCmd command use should contain 'update'")
	assert.NotEmpty(t, updateCmd.Short, "updateCmd command short description should not be empty")
	assert.True(t, updateCmd.Run != nil || updateCmd.RunE != nil, "Run or RunE function should be set for updateCmd")
}

func TestRunUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	cmd := updateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// test no args
	updateAll = false
	updatePreview = false
	err := runUpdate(cmd, []string{})
	assert.Error(t, err)

	// test --preview
	updatePreview = true
	updateForce = true
	err = runUpdate(cmd, []string{})
	assert.NoError(t, err)

	// test --all
	updatePreview = false
	updateAll = true
	err = runUpdate(cmd, []string{})
	assert.NoError(t, err)

	// test tool update (this will fail in tests as tool cannot be fully downloaded but that is ok for coverage if we mock or expect err)
	updateAll = false
	err = runUpdate(cmd, []string{"dummy"})
	assert.Error(t, err)
}
