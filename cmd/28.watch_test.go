// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchStructure(t *testing.T) {
	assert.Contains(t, watchCmd.Use, "watch", "watchCmd command use should contain 'watch'")
	assert.NotEmpty(t, watchCmd.Short, "watchCmd command short description should not be empty")
	assert.True(t, watchCmd.Run != nil || watchCmd.RunE != nil, "Run or RunE function should be set for watchCmd")
}
