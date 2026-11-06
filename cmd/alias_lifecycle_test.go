// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_AliasResolutionLifecycle(t *testing.T) {
	h := NewE2EHarness(t)

	// Create a .unirtm.toml with an alias configured
	cfgContent := `
[tools]
"github:mvdan/sh" = { version = "3.14.1", alias = "shfmt" }
`
	cfgPath := filepath.Join(h.TmpDir, ".unirtm.toml")
	err := os.WriteFile(cfgPath, []byte(cfgContent), 0644)
	assert.NoError(t, err)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	_ = os.Chdir(h.TmpDir)

	_, _, err = h.Run("trust", cfgPath)
	assert.NoError(t, err)

	// 1. Test shell command with alias
	t.Run("shell with alias", func(t *testing.T) {
		stdout, _, err := h.Run("shell", "shfmt@3.14.1")
		assert.NoError(t, err)
		assert.Contains(t, stdout, "UNIRTM_SHFMT_VERSION=\"3.14.1\"")
		assert.Contains(t, stdout, "UNIRTM_GITHUB_MVDAN_SH_VERSION=\"3.14.1\"")
	})

	// 2. Test config get with alias
	t.Run("config get with alias", func(t *testing.T) {
		stdout, _, err := h.Run("config", "get", "tools.shfmt")
		assert.NoError(t, err)
		assert.Contains(t, stdout, "3.14.1")
	})

	// 3. Test current with alias
	t.Run("current with alias", func(t *testing.T) {
		stdout, _, err := h.Run("current", "shfmt")
		assert.NoError(t, err)
		assert.Contains(t, stdout, "3.14.1")
	})

	// 4. Test list with --tool alias
	t.Run("list with --tool alias", func(t *testing.T) {
		_, _, err := h.Run("list", "--tool", "shfmt")
		assert.NoError(t, err)
	})

	// 5. Test bin-paths (all active tools and with alias)
	t.Run("bin-paths", func(t *testing.T) {
		stdout, _, err := h.Run("bin-paths")
		assert.NoError(t, err)
		assert.NotEmpty(t, stdout)

		_, _, err = h.Run("bin-paths", "shfmt")
		assert.NoError(t, err)
	})

	// 6. Test unuse with alias
	t.Run("unuse with alias", func(t *testing.T) {
		_, _, err := h.Run("unuse", "shfmt")
		assert.NoError(t, err)

		// Read back configuration: github:mvdan/sh should be removed
		updatedContent, err := os.ReadFile(cfgPath)
		assert.NoError(t, err)
		assert.NotContains(t, string(updatedContent), "github:mvdan/sh")
	})
}
