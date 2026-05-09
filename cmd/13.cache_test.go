// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheStructure(t *testing.T) {
	assert.Contains(t, cacheCmd.Use, "cache", "cacheCmd command use should contain 'cache'")
	assert.NotEmpty(t, cacheCmd.Short, "cacheCmd command short description should not be empty")
	assert.True(t, cacheCmd.Run != nil || cacheCmd.RunE != nil, "Run or RunE function should be set for cacheCmd")
}
