// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellStructure(t *testing.T) {
	assert.Contains(t, shellCmd.Use, "shell", "shellCmd command use should contain 'shell'")
	assert.NotEmpty(t, shellCmd.Short, "shellCmd command short description should not be empty")
	assert.True(t, shellCmd.Run != nil || shellCmd.RunE != nil, "Run or RunE function should be set for shellCmd")
}
