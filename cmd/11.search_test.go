// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchStructure(t *testing.T) {
	assert.Contains(t, searchCmd.Use, "search", "searchCmd command use should contain 'search'")
	assert.NotEmpty(t, searchCmd.Short, "searchCmd command short description should not be empty")
	assert.True(t, searchCmd.Run != nil || searchCmd.RunE != nil, "Run or RunE function should be set for searchCmd")
}
