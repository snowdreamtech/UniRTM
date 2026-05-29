// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseStructure(t *testing.T) {
	assert.Contains(t, licenseCmd.Use, "license", "licenseCmd command use should contain 'license'")
	assert.NotEmpty(t, licenseCmd.Short, "licenseCmd command short description should not be empty")
	assert.True(t, licenseCmd.HasSubCommands(), "licenseCmd should have subcommands")
}

func TestRunLicenseAdd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	cmd := licenseAddCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Since add needs args, we'll pass a dummy argument that causes it to error quickly
	// if we don't pass --file or --license. That proves it's wired up.
	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{"dummy"})
		assert.Error(t, err)
	}
}
