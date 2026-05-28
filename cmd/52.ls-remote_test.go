// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsRemoteStructure(t *testing.T) {
	assert.Contains(t, lsRemoteCmd.Use, "ls-remote", "lsRemoteCmd command use should contain 'ls-remote'")
	assert.NotEmpty(t, lsRemoteCmd.Short, "lsRemoteCmd command short description should not be empty")
	assert.True(t, lsRemoteCmd.Run != nil || lsRemoteCmd.RunE != nil, "Run or RunE function should be set for lsRemoteCmd")
}

func TestRunLsRemote(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := lsRemoteCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Prevent network calls by using an unknown backend
	lsRemoteBackend = "unknown"
	defer func() { lsRemoteBackend = "" }()

	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{"dummy"})
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{"dummy"})
	}
}
