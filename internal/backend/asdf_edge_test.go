// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsdfResolveToolName(t *testing.T) {
	assert.Equal(t, "nodejs", ResolveAsdfToolName("node"))
	assert.Equal(t, "golang", ResolveAsdfToolName("go"))
	assert.Equal(t, "unknown", ResolveAsdfToolName("unknown"))
}

func TestAsdfBackend_EdgeCases(t *testing.T) {
	b := NewAsdfBackend()

	// Use temporary directory to test file errors
	tempDir := t.TempDir()
	b.registryPath = filepath.Join(tempDir, "registry")
	b.pluginsPath = filepath.Join(tempDir, "plugins")

	// Pre-create the registry path and the FETCH_HEAD file to pretend we recently fetched
	// This avoids "git clone" in tests
	gitDir := filepath.Join(b.registryPath, ".git")
	err := os.MkdirAll(gitDir, 0o755)
	require.NoError(t, err)
	fetchHead := filepath.Join(gitDir, "FETCH_HEAD")
	err = os.WriteFile(fetchHead, []byte("recent"), 0o644)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("ResolveVersion ListVersions Error", func(t *testing.T) {
		// Mock pluginsPath to an unwritable directory to force ensurePlugin error -> ListVersions error
		b.pluginsPath = filepath.Join(tempDir, "unwritable_plugins")
		err := os.WriteFile(b.pluginsPath, []byte("file"), 0o644)
		require.NoError(t, err)

		_, err = b.ResolveVersion(ctx, "testtool", "latest", Platform{OS: "linux", Arch: "amd64"})
		assert.Error(t, err)
	})

	t.Run("GetDownloadInfo EnsurePlugin Error", func(t *testing.T) {
		// Same unwritable path will cause GetDownloadInfo ensurePlugin to fail
		_, err := b.GetDownloadInfo(ctx, "testtool", "1.0.0", Platform{OS: "linux", Arch: "amd64"})
		assert.Error(t, err)
	})

	t.Run("UpdateRegistry MkdirAll Error", func(t *testing.T) {
		oldRegistry := b.registryPath
		b.registryPath = filepath.Join(tempDir, "unwritable_registry", "registry")
		err := os.WriteFile(filepath.Join(tempDir, "unwritable_registry"), []byte("file"), 0o644)
		require.NoError(t, err)

		err = b.updateRegistry(ctx)
		assert.Error(t, err)

		b.registryPath = oldRegistry
	})

	t.Run("UpdateRegistry FETCH_HEAD recent", func(t *testing.T) {
		oldRegistry := b.registryPath
		b.registryPath = filepath.Join(tempDir, "registry_recent")
		gitDir := filepath.Join(b.registryPath, ".git")
		err := os.MkdirAll(gitDir, 0o755)
		require.NoError(t, err)

		fetchHead := filepath.Join(gitDir, "FETCH_HEAD")
		err = os.WriteFile(fetchHead, []byte("recent"), 0o644)
		require.NoError(t, err)

		// Set modification time to now
		err = os.Chtimes(fetchHead, time.Now(), time.Now())
		require.NoError(t, err)

		// It should return nil early without doing anything since it's < 24h
		err = b.updateRegistry(ctx)
		assert.NoError(t, err)

		b.registryPath = oldRegistry
	})

	t.Run("LookupPluginURL empty", func(t *testing.T) {
		oldRegistry := b.registryPath
		b.registryPath = filepath.Join(tempDir, "registry_empty")
		pluginsDir := filepath.Join(b.registryPath, "plugins")
		err := os.MkdirAll(pluginsDir, 0o755)
		require.NoError(t, err)

		toolFile := filepath.Join(pluginsDir, "emptytool")
		// Write an empty repository string
		err = os.WriteFile(toolFile, []byte("repository = \n"), 0o644)
		require.NoError(t, err)

		url, err := b.lookupPluginURL("emptytool")
		assert.Error(t, err)
		assert.Equal(t, "", url)

		b.registryPath = oldRegistry
	})
}
