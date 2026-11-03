// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommandStructure(t *testing.T) {
	assert.Contains(t, listCmd.Use, "list")
	assert.NotEmpty(t, listCmd.Short)
	assert.NotNil(t, listCmd.RunE)
}

func TestRunList_EmptyDB(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	err := listCmd.RunE(listCmd, []string{})
	assert.NoError(t, err)
}
