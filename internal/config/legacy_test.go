// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadLegacyVersionFiles_IndividualFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-legacy-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 1. Create .nvmrc with leading 'v' and comments
	err = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("v20.18.0 # LTS version\n"), 0644)
	require.NoError(t, err)

	// 2. Create .go-version with leading 'go'
	err = os.WriteFile(filepath.Join(tempDir, ".go-version"), []byte("go1.23.2\n"), 0644)
	require.NoError(t, err)

	// 3. Create .python-version
	err = os.WriteFile(filepath.Join(tempDir, ".python-version"), []byte("3.12.7\n"), 0644)
	require.NoError(t, err)

	tools, sources := ReadLegacyVersionFiles(tempDir)

	assert.Equal(t, "20.18.0", tools["node"].Version)
	assert.Equal(t, ".nvmrc", sources["node"])

	assert.Equal(t, "1.23.2", tools["go"].Version)
	assert.Equal(t, ".go-version", sources["go"])

	assert.Equal(t, "3.12.7", tools["python"].Version)
	assert.Equal(t, ".python-version", sources["python"])
}

func TestReadLegacyVersionFiles_NodeVersionFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-legacy-node-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Only .node-version exists
	err = os.WriteFile(filepath.Join(tempDir, ".node-version"), []byte("22.1.0\n"), 0644)
	require.NoError(t, err)

	tools, sources := ReadLegacyVersionFiles(tempDir)
	assert.Equal(t, "22.1.0", tools["node"].Version)
	assert.Equal(t, ".node-version", sources["node"])

	// If .nvmrc is added, it should take precedence
	err = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("20.18.0\n"), 0644)
	require.NoError(t, err)

	tools2, sources2 := ReadLegacyVersionFiles(tempDir)
	assert.Equal(t, "20.18.0", tools2["node"].Version)
	assert.Equal(t, ".nvmrc", sources2["node"])
}

func TestReadLegacyVersionFiles_ToolVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-legacy-tv-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tvContent := `# Asdf / mise version configuration
nodejs 20.11.0
golang 1.22.1
python 3.11.8
rust 1.76.0
`
	err = os.WriteFile(filepath.Join(tempDir, ".tool-versions"), []byte(tvContent), 0644)
	require.NoError(t, err)

	tools, sources := ReadLegacyVersionFiles(tempDir)

	assert.Equal(t, "20.11.0", tools["node"].Version)
	assert.Equal(t, ".tool-versions", sources["node"])

	assert.Equal(t, "1.22.1", tools["go"].Version)
	assert.Equal(t, ".tool-versions", sources["go"])

	assert.Equal(t, "3.11.8", tools["python"].Version)
	assert.Equal(t, ".tool-versions", sources["python"])

	assert.Equal(t, "1.76.0", tools["rust"].Version)
	assert.Equal(t, ".tool-versions", sources["rust"])
}

func TestLoadFromDir_LegacyOnlyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-legacy-loaddir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// No unirtm.toml, only .nvmrc and .go-version
	err = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("v18.20.0\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, ".go-version"), []byte("1.21.5\n"), 0644)
	require.NoError(t, err)

	cfg, err := LoadFromDir(tempDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "18.20.0", cfg.Tools["node"].Version)
	assert.Equal(t, ".nvmrc", cfg.LegacySources["node"])
	assert.Equal(t, "1.21.5", cfg.Tools["go"].Version)
	assert.Equal(t, ".go-version", cfg.LegacySources["go"])
}

func TestLoadFromDir_PrimaryOverridesLegacy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-legacy-override-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Primary .unirtm.toml pins node to 22.0.0
	unirtmContent := `[tools]
node = "22.0.0"
`
	err = os.WriteFile(filepath.Join(tempDir, ".unirtm.toml"), []byte(unirtmContent), 0644)
	require.NoError(t, err)

	// Legacy .nvmrc has 20.0.0, and .go-version has 1.23.0
	err = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("20.0.0\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, ".go-version"), []byte("1.23.0\n"), 0644)
	require.NoError(t, err)

	cfg, err := LoadFromDir(tempDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// node comes from .unirtm.toml
	assert.Equal(t, "22.0.0", cfg.Tools["node"].Version)
	assert.Empty(t, cfg.LegacySources["node"])

	// go was missing in .unirtm.toml, so it is filled from .go-version
	assert.Equal(t, "1.23.0", cfg.Tools["go"].Version)
	assert.Equal(t, ".go-version", cfg.LegacySources["go"])
}
