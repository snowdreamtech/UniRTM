// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunStructure(t *testing.T) {
	assert.Contains(t, runCmd.Use, "run", "runCmd command use should contain 'run'")
	assert.NotEmpty(t, runCmd.Short, "runCmd command short description should not be empty")
	assert.True(t, runCmd.Run != nil || runCmd.RunE != nil, "Run or RunE function should be set for runCmd")
}

func TestRunRun(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := runCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
