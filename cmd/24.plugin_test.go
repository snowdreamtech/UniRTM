// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginStructure(t *testing.T) {
	assert.Contains(t, pluginCmd.Use, "plugin", "pluginCmd command use should contain 'plugin'")
	assert.NotEmpty(t, pluginCmd.Short, "pluginCmd command short description should not be empty")
	assert.True(t, pluginCmd.Run != nil || pluginCmd.RunE != nil, "Run or RunE function should be set for pluginCmd")
}
