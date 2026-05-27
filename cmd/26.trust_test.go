// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustStructure(t *testing.T) {
	assert.Contains(t, trustCmd.Use, "trust", "trustCmd command use should contain 'trust'")
	assert.NotEmpty(t, trustCmd.Short, "trustCmd command short description should not be empty")
	assert.True(t, trustCmd.Run != nil || trustCmd.RunE != nil, "Run or RunE function should be set for trustCmd")
}

func TestRunTrust(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	// Create a dummy config file to trust
	dummyFile := filepath.Join(tmpDir, "unirtm.toml")
	os.WriteFile(dummyFile, []byte(""), 0644)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{dummyFile})
		assert.NoError(t, err)
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{dummyFile})
	}
}
