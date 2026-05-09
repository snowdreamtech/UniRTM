// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigStructure(t *testing.T) {
	assert.Contains(t, configCmd.Use, "config", "configCmd command use should contain 'config'")
	assert.NotEmpty(t, configCmd.Short, "configCmd command short description should not be empty")
	assert.True(t, configCmd.Run != nil || configCmd.RunE != nil, "Run or RunE function should be set for configCmd")
}
