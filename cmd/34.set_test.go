// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCommandStructure(t *testing.T) {
	assert.Contains(t, setCmd.Use, "set")
	assert.NotNil(t, setCmd.RunE)
	assert.NotNil(t, setCmd.Flags().Lookup("global"))
}

func TestUnsetCommandStructure(t *testing.T) {
	assert.Contains(t, unsetCmd.Use, "unset")
	assert.NotNil(t, unsetCmd.RunE)
}

func TestLoadRawTOML_Missing(t *testing.T) {
	m, err := loadRawTOML("/nonexistent/path/unirtm.toml")
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Empty(t, m)
}

func TestLoadRawTOML_Valid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "unirtm.toml")
	require.NoError(t, os.WriteFile(path, []byte("[env]\nFOO = \"bar\"\n"), 0o644))

	m, err := loadRawTOML(path)
	require.NoError(t, err)
	envMap, ok := m["env"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "bar", envMap["FOO"])
}

func TestSaveRawTOML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "unirtm.toml")

	m := map[string]interface{}{
		"env": map[string]interface{}{"NODE_ENV": "test"},
	}
	require.NoError(t, saveRawTOML(path, m))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "NODE_ENV")
}

func TestSetUnsetRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "unirtm.toml")

	// Set FOO=bar
	m, _ := loadRawTOML(path)
	envSection(m)["FOO"] = "bar"
	require.NoError(t, saveRawTOML(path, m))

	// Reload and verify
	m2, _ := loadRawTOML(path)
	envMap := envSection(m2)
	assert.Equal(t, "bar", envMap["FOO"])

	// Unset FOO
	delete(envMap, "FOO")
	if len(envMap) == 0 {
		delete(m2, "env")
	}
	require.NoError(t, saveRawTOML(path, m2))

	// Verify removed
	m3, _ := loadRawTOML(path)
	_, hasEnv := m3["env"]
	assert.False(t, hasEnv)
}

func TestResolveConfigFilePath_Default(t *testing.T) {
	// In a temp dir with no config files, should return "unirtm.toml"
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	defer os.Chdir(orig)

	got := resolveConfigFilePath(false)
	assert.Equal(t, "unirtm.toml", got)
}

func TestResolveConfigFilePath_ExistingFile(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	defer os.Chdir(orig)

	// Create .unirtm.toml — should be preferred over default unirtm.toml
	require.NoError(t, os.WriteFile(".unirtm.toml", []byte(""), 0o644))
	got := resolveConfigFilePath(false)
	assert.Equal(t, ".unirtm.toml", got)
}
