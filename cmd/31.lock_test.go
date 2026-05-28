// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockCommandStructure(t *testing.T) {
	assert.Equal(t, "lock [tool...]", lockCmd.Use)
}

func TestRunLockCheck(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	// Since there is no config file, it should just return no error
	err := lockCmd.RunE(lockCmd, []string{})
	assert.NoError(t, err)
}
