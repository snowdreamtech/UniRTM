// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendsCommandStructure(t *testing.T) {
	assert.Contains(t, backendsCmd.Use, "backends")
	assert.NotEmpty(t, backendsCmd.Short)
}

func TestBackendsListSubcommand(t *testing.T) {
	assert.NotNil(t, backendsListCmd.RunE)
}

func TestBackendsInfoSubcommand(t *testing.T) {
	assert.NotNil(t, backendsInfoCmd.RunE)
	err := backendsInfoCmd.Args(backendsInfoCmd, []string{"github"})
	assert.NoError(t, err)
	err = backendsInfoCmd.Args(backendsInfoCmd, []string{})
	assert.Error(t, err)
}

func TestBackendEntry_Struct(t *testing.T) {
	e := backendEntry{Name: "github", SupportsChecksum: true, SupportsGPG: false}
	assert.Equal(t, "github", e.Name)
	assert.True(t, e.SupportsChecksum)
	assert.False(t, e.SupportsGPG)
}
