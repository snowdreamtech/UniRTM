// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinPathsCommandStructure(t *testing.T) {
	assert.Contains(t, binPathsCmd.Use, "bin-paths")
	assert.NotEmpty(t, binPathsCmd.Short)
	assert.NotNil(t, binPathsCmd.RunE)
}
