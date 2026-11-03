// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryCommandStructure(t *testing.T) {
	assert.Contains(t, registryCmd.Use, "registry")
	assert.NotNil(t, registryCmd.RunE)
	assert.NotNil(t, registryCmd.Flags().Lookup("search"))
}

func TestRegistryEntry_Struct(t *testing.T) {
	e := registryEntry{Tool: "node", Backend: "native", Provider: "node"}
	assert.Equal(t, "node", e.Tool)
	assert.Equal(t, "native", e.Backend)
}
