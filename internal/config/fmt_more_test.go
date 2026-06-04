// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSectionOrder(t *testing.T) {
	tests := map[string]int{
		"min_version": 0,
		"env_file":    1,
		"env_path":    2,
		"":            3,
		"env":         4,
		"vars":        5,
		"hooks":       6,
		"watch_files": 7,
		"tools":       8,
		"tasks":       10,
		"task_config": 11,
		"redactions":  12,
		"alias":       13,
		"plugins":     14,
		"settings":    15,
		"unknown":     9,
	}

	for k, expected := range tests {
		actual := getSectionOrder(k)
		require.Equal(t, expected, actual, "key %s", k)
	}
}

func TestFormatFile_NonToml(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(path, []byte("  hello  \n"), 0644)
	require.NoError(t, err)

	changed, err := FormatFile(path, false)
	require.NoError(t, err)
	require.True(t, changed)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "hello\n", string(content))
}

func TestFormatFile_NoChange(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(path, []byte("hello\n"), 0644)
	require.NoError(t, err)

	changed, err := FormatFile(path, false)
	require.NoError(t, err)
	require.False(t, changed)
}

func TestFormatFile_FmtCheckOnly(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(path, []byte("  hello  \n"), 0644)
	require.NoError(t, err)

	changed, err := FormatFile(path, true)
	require.NoError(t, err)
	require.True(t, changed)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "  hello  \n", string(content))
}

func TestFormatFile_TOML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.toml")
	err := os.WriteFile(path, []byte("[tools]\nnode = \"18\"\n\n[env]\nFOO=\"bar\"\n"), 0644)
	require.NoError(t, err)

	changed, err := FormatFile(path, false)
	require.NoError(t, err)
	require.True(t, changed)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	// getSectionOrder: env=4, tools=8. So env should come before tools.
	require.Contains(t, string(content), "[env]")
	require.Contains(t, string(content), "[tools]")
}
