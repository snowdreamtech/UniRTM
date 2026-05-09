// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUseStructure(t *testing.T) {
	assert.Contains(t, useCmd.Use, "use", "useCmd command use should contain 'use'")
	assert.NotEmpty(t, useCmd.Short, "useCmd command short description should not be empty")
	assert.True(t, useCmd.Run != nil || useCmd.RunE != nil, "Run or RunE function should be set for useCmd")
}
