// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheStructure(t *testing.T) {
	assert.Contains(t, cacheCmd.Use, "cache", "cacheCmd command use should contain 'cache'")
	assert.NotEmpty(t, cacheCmd.Short, "cacheCmd command short description should not be empty")
	assert.True(t, cacheCmd.Run != nil || cacheCmd.RunE != nil, "Run or RunE function should be set for cacheCmd")
}

func TestRunCacheList(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cacheListCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCacheList(cmd, []string{})
	assert.NoError(t, err)

	// test json
	jsonOutput = true
	defer func() { jsonOutput = false }()
	err = runCacheList(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunCacheStats(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cacheStatsCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCacheStats(cmd, []string{})
	assert.NoError(t, err)

	jsonOutput = true
	defer func() { jsonOutput = false }()
	err = runCacheStats(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cacheClearCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCacheClear(cmd, []string{})
	assert.NoError(t, err)

	err = runCacheClear(cmd, []string{"dummy"})
	assert.NoError(t, err)
}

func TestRunCachePurge(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cachePurgeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runCachePurge(cmd, []string{})
	assert.NoError(t, err)
}

func TestCachePathCmd(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cachePathCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

func TestCacheCmd(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := cacheCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}
