// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutdatedCommandStructure(t *testing.T) {
	assert.Contains(t, outdatedCmd.Use, "outdated")
	assert.NotEmpty(t, outdatedCmd.Short)
	assert.NotNil(t, outdatedCmd.RunE)
}

func TestOutdatedResult_Struct(t *testing.T) {
	r := outdatedResult{
		Tool:     "cli/cli",
		Backend:  "github",
		Current:  "2.70.0",
		Latest:   "2.72.0",
		Outdated: true,
	}
	assert.True(t, r.Outdated)
	assert.Equal(t, "2.72.0", r.Latest)
}

func TestOutdatedResult_UpToDate(t *testing.T) {
	r := outdatedResult{
		Tool:     "cli/cli",
		Backend:  "github",
		Current:  "2.72.0",
		Latest:   "2.72.0",
		Outdated: false,
	}
	assert.False(t, r.Outdated)
}

func TestOutdatedRun_EmptyDB(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	err := outdatedCmd.RunE(outdatedCmd, []string{})
	assert.NoError(t, err)
}

