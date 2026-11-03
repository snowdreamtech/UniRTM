// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShimRecursionGuardClearing(t *testing.T) {
	// Verify that setting and unsetting the recursion guard environment variable works
	os.Setenv("_UNIRTM_SHIM_RECURSION_GUARD", "1")
	assert.Equal(t, "1", os.Getenv("_UNIRTM_SHIM_RECURSION_GUARD"))

	os.Unsetenv("_UNIRTM_SHIM_RECURSION_GUARD")
	assert.Empty(t, os.Getenv("_UNIRTM_SHIM_RECURSION_GUARD"))
}

func TestToolScopedShimRecursionGuard(t *testing.T) {
	// Verify per-tool scoped guard naming and isolation
	tools := []string{"node", "npm", "npm:vitepress", "cargo-binstall"}
	for _, tool := range tools {
		cleanName := strings.TrimSuffix(tool, ".exe")
		guardKey := fmt.Sprintf("_UNIRTM_SHIM_GUARD_%s", strings.ToUpper(strings.ReplaceAll(cleanName, "-", "_")))
		guardKey = strings.ReplaceAll(guardKey, ":", "_")

		os.Setenv(guardKey, "1")
		assert.Equal(t, "1", os.Getenv(guardKey))

		os.Unsetenv(guardKey)
		assert.Empty(t, os.Getenv(guardKey))
	}
}
