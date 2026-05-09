// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateStructure(t *testing.T) {
	assert.Contains(t, migrateCmd.Use, "migrate", "migrateCmd command use should contain 'migrate'")
	assert.NotEmpty(t, migrateCmd.Short, "migrateCmd command short description should not be empty")
	assert.True(t, migrateCmd.Run != nil || migrateCmd.RunE != nil, "Run or RunE function should be set for migrateCmd")
}
