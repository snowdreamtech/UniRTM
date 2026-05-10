// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkCommandStructure(t *testing.T) {
	assert.Contains(t, linkCmd.Use, "link")
	assert.NotNil(t, linkCmd.RunE)
	assert.NotNil(t, linkCmd.Flags().Lookup("backend"))
}

func TestLinkCommandArgs(t *testing.T) {
	err := linkCmd.Args(linkCmd, []string{"node", "22.14.0", "/usr/local"})
	assert.NoError(t, err)
	err = linkCmd.Args(linkCmd, []string{"node", "22.14.0"})
	assert.Error(t, err)
}

func TestUnuseCommandStructure(t *testing.T) {
	assert.Contains(t, unuseCmd.Use, "unuse")
	assert.Contains(t, unuseCmd.Aliases, "rm")
	assert.Contains(t, unuseCmd.Aliases, "remove")
	assert.NotNil(t, unuseCmd.RunE)
}

func TestUnuseCommandArgs(t *testing.T) {
	err := unuseCmd.Args(unuseCmd, []string{"node"})
	assert.NoError(t, err)
	err = unuseCmd.Args(unuseCmd, []string{})
	assert.Error(t, err)
}
