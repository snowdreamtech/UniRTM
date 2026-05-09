// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateStructure(t *testing.T) {
	assert.Contains(t, updateCmd.Use, "update", "updateCmd command use should contain 'update'")
	assert.NotEmpty(t, updateCmd.Short, "updateCmd command short description should not be empty")
	assert.True(t, updateCmd.Run != nil || updateCmd.RunE != nil, "Run or RunE function should be set for updateCmd")
}
