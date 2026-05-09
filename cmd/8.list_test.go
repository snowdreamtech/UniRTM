// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommandStructure(t *testing.T) {
	assert.Contains(t, listCmd.Use, "list", "listCmd command use should contain 'list'")
	assert.NotEmpty(t, listCmd.Short, "list command short description should not be empty")
	assert.True(t, listCmd.Run != nil || listCmd.RunE != nil, "Run or RunE function should be set for listCmd")
}
