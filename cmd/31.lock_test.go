// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockStructure(t *testing.T) {
	assert.Contains(t, lockCmd.Use, "lock", "lockCmd command use should contain 'lock'")
	assert.NotEmpty(t, lockCmd.Short, "lockCmd command short description should not be empty")
	assert.True(t, lockCmd.Run != nil || lockCmd.RunE != nil, "Run or RunE function should be set for lockCmd")
}

func TestRunLock(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := lockCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runLock(cmd, []string{})
	assert.NoError(t, err)
}
