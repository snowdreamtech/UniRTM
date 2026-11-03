// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDfStructure(t *testing.T) {
	assert.Equal(t, "df", dfCmd.Use, "dfCmd command use should be 'df'")
	assert.NotEmpty(t, dfCmd.Short, "dfCmd command short description should not be empty")

	// Ensure the human-readable flag is mapped to 'h'
	flag := dfCmd.Flags().Lookup("human-readable")
	require.NotNil(t, flag, "flag human-readable should be registered")
	assert.Equal(t, "h", flag.Shorthand, "flag human-readable should have shorthand h")
}

func TestRunDf(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	cmd := dfCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
	}
}
