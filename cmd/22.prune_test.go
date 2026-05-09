// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPruneStructure(t *testing.T) {
	assert.Contains(t, pruneCmd.Use, "prune", "pruneCmd command use should contain 'prune'")
	assert.NotEmpty(t, pruneCmd.Short, "pruneCmd command short description should not be empty")
	assert.True(t, pruneCmd.Run != nil || pruneCmd.RunE != nil, "Run or RunE function should be set for pruneCmd")
}
