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

func TestUseStructure(t *testing.T) {
	assert.Contains(t, useCmd.Use, "use", "useCmd command use should contain 'use'")
	assert.NotEmpty(t, useCmd.Short, "useCmd command short description should not be empty")
	assert.True(t, useCmd.Run != nil || useCmd.RunE != nil, "Run or RunE function should be set for useCmd")
}

func TestUseCommandFlags(t *testing.T) {
	globalFlag := useCmd.Flags().Lookup("global")
	require.NotNil(t, globalFlag)
	assert.Equal(t, "g", globalFlag.Shorthand)
	assert.Equal(t, "false", globalFlag.DefValue)

	pathFlag := useCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag)
	assert.Equal(t, "p", pathFlag.Shorthand)
	assert.Equal(t, "", pathFlag.DefValue)

	forceFlag := useCmd.Flags().Lookup("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
	assert.Equal(t, "false", forceFlag.DefValue)

	envFlag := useCmd.Flags().Lookup("env")
	require.NotNil(t, envFlag)
	assert.Equal(t, "e", envFlag.Shorthand)
	assert.Equal(t, "", envFlag.DefValue)
}

func TestFindOrCreateConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Default (no env) when file doesn't exist
	p1 := findOrCreateConfigFile(tmpDir, "")
	assert.Equal(t, filepath.Join(tmpDir, "unirtm.toml"), p1)

	// 2. Default (no env) when .unirtm.toml exists
	dotFile := filepath.Join(tmpDir, ".unirtm.toml")
	err := os.WriteFile(dotFile, []byte(""), 0644)
	require.NoError(t, err)
	p2 := findOrCreateConfigFile(tmpDir, "")
	assert.Equal(t, dotFile, p2)
	_ = os.Remove(dotFile)

	// 3. With env "prod" when file doesn't exist
	p3 := findOrCreateConfigFile(tmpDir, "prod")
	assert.Equal(t, filepath.Join(tmpDir, "unirtm.prod.toml"), p3)

	// 4. With env "prod" when .unirtm.prod.toml exists
	dotProdFile := filepath.Join(tmpDir, ".unirtm.prod.toml")
	err = os.WriteFile(dotProdFile, []byte(""), 0644)
	require.NoError(t, err)
	p4 := findOrCreateConfigFile(tmpDir, "prod")
	assert.Equal(t, dotProdFile, p4)
	_ = os.Remove(dotProdFile)
}
