// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrustManager_TrustUntrustList(t *testing.T) {
	tmpDir := t.TempDir()
	trustFile := filepath.Join(tmpDir, "trusted.toml")
	tm := &fileTrustManager{trustFilePath: trustFile}

	// List empty
	list, err := tm.List()
	require.NoError(t, err)
	require.Empty(t, list)

	// Trust a file
	cfgFile := filepath.Join(tmpDir, "unirtm.toml")
	err = os.WriteFile(cfgFile, []byte("min_version = '1.0.0'"), 0644)
	require.NoError(t, err)

	err = tm.Trust(cfgFile)
	require.NoError(t, err)

	// Check status
	trusted := tm.TrustStatus(cfgFile)
	require.Equal(t, TrustStatusTrusted, trusted)

	// List should contain the file
	list, err = tm.List()
	require.NoError(t, err)
	require.Contains(t, list, cfgFile)

	// Untrust
	err = tm.Untrust(cfgFile)
	require.NoError(t, err)

	trusted = tm.TrustStatus(cfgFile)
	require.Equal(t, TrustStatusUntrusted, trusted)
}

func TestTrustManager_TrustStatus_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	trustFile := filepath.Join(tmpDir, "trusted.toml")
	tm := &fileTrustManager{trustFilePath: trustFile}

	trusted := tm.TrustStatus("/invalid/path/that/does/not/exist")
	require.Equal(t, TrustStatusUntrusted, trusted)
}

func TestTrustManager_List_CleansUpDeletedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	trustFile := filepath.Join(tmpDir, "trusted.toml")
	tm := &fileTrustManager{trustFilePath: trustFile}

	cfgFile := filepath.Join(tmpDir, "unirtm.toml")
	err := os.WriteFile(cfgFile, []byte(""), 0644)
	require.NoError(t, err)

	err = tm.Trust(cfgFile)
	require.NoError(t, err)

	// Delete file
	os.Remove(cfgFile)

	list, err := tm.List()
	require.NoError(t, err)
	require.Empty(t, list) // should be cleaned up
}

func TestNewTrustManager(t *testing.T) {
	tm := NewTrustManager()
	require.NotNil(t, tm)
}

// TestUntrust_CleansStaleRecordWhenFileDeleted verifies that calling Untrust on a
// file that has been deleted still removes its stale entry from the database.
func TestUntrust_CleansStaleRecordWhenFileDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	trustFile := filepath.Join(tmpDir, "trusted.toml")
	tm := &fileTrustManager{trustFilePath: trustFile}

	cfgFile := filepath.Join(tmpDir, "unirtm.toml")
	os.WriteFile(cfgFile, []byte(""), 0644)

	// Trust it first
	require.NoError(t, tm.Trust(cfgFile))

	// Delete the file
	os.Remove(cfgFile)

	// Untrust should still remove the stale record
	require.NoError(t, tm.Untrust(cfgFile))

	// Verify record is gone
	paths, err := tm.loadTrustedPaths()
	require.NoError(t, err)
	absConfig, _ := filepath.Abs(cfgFile)
	require.NotContains(t, paths, absConfig, "stale record should be removed by Untrust even when file is deleted")
}

// TestStaleRecordCleanup_AllOperations verifies that each trust operation
// (Trust, Untrust, List) correctly removes stale records for deleted files.
func TestStaleRecordCleanup_AllOperations(t *testing.T) {
	setup := func(t *testing.T) (*fileTrustManager, string) {
		t.Helper()
		tmpDir := t.TempDir()
		tm := &fileTrustManager{trustFilePath: filepath.Join(tmpDir, "trusted.toml")}
		cfgFile := filepath.Join(tmpDir, "unirtm.toml")
		os.WriteFile(cfgFile, []byte(""), 0644)
		require.NoError(t, tm.Trust(cfgFile))
		return tm, cfgFile
	}

	t.Run("Trust cleans stale record", func(t *testing.T) {
		tm, cfgFile := setup(t)
		os.Remove(cfgFile)

		err := tm.Trust(cfgFile)
		require.Error(t, err) // Trust fails because file is gone

		paths, _ := tm.loadTrustedPaths()
		absConfig, _ := filepath.Abs(cfgFile)
		require.NotContains(t, paths, absConfig)
	})

	t.Run("Untrust cleans stale record", func(t *testing.T) {
		tm, cfgFile := setup(t)
		os.Remove(cfgFile)

		err := tm.Untrust(cfgFile)
		require.NoError(t, err) // Untrust succeeds even when file is gone

		paths, _ := tm.loadTrustedPaths()
		absConfig, _ := filepath.Abs(cfgFile)
		require.NotContains(t, paths, absConfig)
	})

	t.Run("List cleans stale record", func(t *testing.T) {
		tm, cfgFile := setup(t)
		os.Remove(cfgFile)

		list, err := tm.List()
		require.NoError(t, err)
		absConfig, _ := filepath.Abs(cfgFile)
		require.NotContains(t, list, absConfig)

		// Verify record is also gone from raw DB
		paths, _ := tm.loadTrustedPaths()
		require.NotContains(t, paths, absConfig)
	})
}
