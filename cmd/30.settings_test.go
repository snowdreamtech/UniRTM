// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingsStructure(t *testing.T) {
	assert.Contains(t, settingsCmd.Use, "settings", "settingsCmd command use should contain 'settings'")
	assert.NotEmpty(t, settingsCmd.Short, "settingsCmd command short description should not be empty")
	assert.True(t, settingsCmd.Run != nil || settingsCmd.RunE != nil, "Run or RunE function should be set for settingsCmd")
}

func TestRunSettings(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := settingsCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{})
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
