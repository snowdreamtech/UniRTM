// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActivateCommandStructure(t *testing.T) {
	assert.Contains(t, activateCmd.Use, "activate", "activateCmd command use should contain 'activate'")
	assert.NotEmpty(t, activateCmd.Short, "activate command short description should not be empty")
	assert.True(t, activateCmd.Run != nil || activateCmd.RunE != nil, "Run or RunE function should be set for activateCmd")
}
