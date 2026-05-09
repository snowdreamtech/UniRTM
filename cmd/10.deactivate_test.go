// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeactivateCommandStructure(t *testing.T) {
	assert.Contains(t, deactivateCmd.Use, "deactivate", "deactivateCmd command use should contain 'deactivate'")
	assert.NotEmpty(t, deactivateCmd.Short, "deactivate command short description should not be empty")
	assert.True(t, deactivateCmd.Run != nil || deactivateCmd.RunE != nil, "Run or RunE function should be set for deactivateCmd")
}
