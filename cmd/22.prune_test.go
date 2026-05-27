// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPruneStructure(t *testing.T) {
	assert.Contains(t, pruneCmd.Use, "prune", "pruneCmd command use should contain 'prune'")
	assert.NotEmpty(t, pruneCmd.Short, "pruneCmd command short description should not be empty")
	assert.True(t, pruneCmd.Run != nil || pruneCmd.RunE != nil, "Run or RunE function should be set for pruneCmd")
}

func TestRunPrune(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := pruneCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// prune without confirm should fail in non-interactive if we don't mock it, but actually here since it tries to read stdin, we'll set pruneYes = true
	pruneYes = true
	err := runPrune(cmd, []string{})
	assert.NoError(t, err)

	pruneTool = "dummy"
	err = runPrune(cmd, []string{})
	assert.NoError(t, err)
}
