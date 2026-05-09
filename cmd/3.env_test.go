// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvCommandStructure(t *testing.T) {
	assert.Contains(t, envCmd.Use, "env", "envCmd command use should contain 'env'")
	assert.NotEmpty(t, envCmd.Short, "env command short description should not be empty")
	assert.True(t, envCmd.Run != nil || envCmd.RunE != nil, "Run or RunE function should be set for envCmd")
}
