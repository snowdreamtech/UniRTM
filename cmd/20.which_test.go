// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhichStructure(t *testing.T) {
	assert.Contains(t, whichCmd.Use, "which", "whichCmd command use should contain 'which'")
	assert.NotEmpty(t, whichCmd.Short, "whichCmd command short description should not be empty")
	assert.True(t, whichCmd.Run != nil || whichCmd.RunE != nil, "Run or RunE function should be set for whichCmd")
}
