// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
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
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := settingsCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{})
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}
