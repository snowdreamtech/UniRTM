// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecStructure(t *testing.T) {
	assert.Contains(t, execCmd.Use, "exec", "execCmd command use should contain 'exec'")
	assert.NotEmpty(t, execCmd.Short, "execCmd command short description should not be empty")
	assert.True(t, execCmd.Run != nil || execCmd.RunE != nil, "Run or RunE function should be set for execCmd")
}
