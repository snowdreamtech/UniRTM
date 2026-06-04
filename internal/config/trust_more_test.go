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
