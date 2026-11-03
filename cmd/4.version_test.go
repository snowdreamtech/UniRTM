// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommandStructure(t *testing.T) {
	assert.Contains(t, versionCmd.Use, "version", "versionCmd command use should contain 'version'")
	assert.NotEmpty(t, versionCmd.Short, "version command short description should not be empty")
	assert.True(t, versionCmd.Run != nil || versionCmd.RunE != nil, "Run or RunE function should be set for versionCmd")
}
