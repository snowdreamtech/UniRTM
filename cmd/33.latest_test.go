// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatestCommandStructure(t *testing.T) {
	assert.Contains(t, latestCmd.Use, "latest")
	assert.NotEmpty(t, latestCmd.Short)
	assert.NotNil(t, latestCmd.RunE)
}

func TestLatestCommandFlags(t *testing.T) {
	assert.NotNil(t, latestCmd.Flags().Lookup("backend"), "latest should have --backend flag")
}

func TestLatestCommandArgs(t *testing.T) {
	// Must accept 1–2 args.
	err := latestCmd.Args(latestCmd, []string{"cli/cli"})
	assert.NoError(t, err)

	err = latestCmd.Args(latestCmd, []string{"cli/cli", "2.70"})
	assert.NoError(t, err)

	err = latestCmd.Args(latestCmd, []string{})
	assert.Error(t, err, "0 args should fail")

	err = latestCmd.Args(latestCmd, []string{"a", "b", "c"})
	assert.Error(t, err, "3 args should fail")
}
