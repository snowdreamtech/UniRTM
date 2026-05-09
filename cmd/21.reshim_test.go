// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReshimStructure(t *testing.T) {
	assert.Contains(t, reshimCmd.Use, "reshim", "reshimCmd command use should contain 'reshim'")
	assert.NotEmpty(t, reshimCmd.Short, "reshimCmd command short description should not be empty")
	assert.True(t, reshimCmd.Run != nil || reshimCmd.RunE != nil, "Run or RunE function should be set for reshimCmd")
}
