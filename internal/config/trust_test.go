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

func TestFileTrustManager(t *testing.T) {
	// Create a temp dir to act as our config dir containing the trusted_configs file
	tmpDir := t.TempDir()
	trustFilePath := filepath.Join(tmpDir, "trusted_configs")

	// Initialize our trust manager pointing to the temp file
	manager := &fileTrustManager{
		trustFilePath: trustFilePath,
	}

	// Create a dummy project config file to trust
	projectDir := t.TempDir()
	projectConfig := filepath.Join(projectDir, "unirtm.toml")
	content1 := `[tools]
node = { version = "18.0.0" }
`
	err := os.WriteFile(projectConfig, []byte(content1), 0644)
	require.NoError(t, err)

	// Test 1: Initially Untrusted
	assert.Equal(t, TrustStatusUntrusted, manager.TrustStatus(projectConfig), "Initially the file should be untrusted")

	// Test 2: Trust the file
	err = manager.Trust(projectConfig)
	require.NoError(t, err)

	// Verify status is Trusted
	assert.Equal(t, TrustStatusTrusted, manager.TrustStatus(projectConfig), "After trusting, status should be Trusted")

	// Verify the hash was written to the file
	fileBytes, err := os.ReadFile(trustFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(fileBytes), projectConfig, "Trust file should contain the absolute path")

	hash1, err := calculateHash(projectConfig)
	require.NoError(t, err)
	assert.Contains(t, string(fileBytes), hash1, "Trust file should contain the hash")

	// Test 3: Modify the file (Hash mismatch)
	content2 := `[tools]
node = { version = "20.0.0" }
`
	err = os.WriteFile(projectConfig, []byte(content2), 0644)
	require.NoError(t, err)

	assert.Equal(t, TrustStatusModified, manager.TrustStatus(projectConfig), "After modification, status should be Modified")

	// Test 4: Re-trust the modified file
	err = manager.Trust(projectConfig)
	require.NoError(t, err)
	assert.Equal(t, TrustStatusTrusted, manager.TrustStatus(projectConfig), "After re-trusting, status should be Trusted again")

	// Test 5: Untrust the file
	err = manager.Untrust(projectConfig)
	require.NoError(t, err)
	assert.Equal(t, TrustStatusUntrusted, manager.TrustStatus(projectConfig), "After untrusting, status should be Untrusted")

	// Test 6: Legacy format compatibility
	// Write a line without a hash
	absPath, _ := filepath.Abs(projectConfig)
	legacyLine := absPath + "\n"
	err = os.WriteFile(trustFilePath, []byte(legacyLine), 0644)
	require.NoError(t, err)

	// Legacy format (no hash) should fall back to TrustStatusModified so user is prompted to re-trust
	assert.Equal(t, TrustStatusModified, manager.TrustStatus(projectConfig), "Legacy format without hash should return Modified")
}

func TestFileTrustManager_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	trustFilePath := filepath.Join(tmpDir, "trusted_configs_err")
	manager := &fileTrustManager{
		trustFilePath: trustFilePath,
	}

	dummyPath := filepath.Join(tmpDir, "dummy")
	os.WriteFile(dummyPath, []byte("test"), 0644)

	// Test 1: OsMkdirAll error
	origMkdirAll := OsMkdirAll
	defer func() { OsMkdirAll = origMkdirAll }()
	OsMkdirAll = func(path string, perm os.FileMode) error {
		return os.ErrPermission
	}
	err := manager.Trust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Restore
	OsMkdirAll = origMkdirAll

	// Test 2: OsOpenFile error in ensureTrustFileExists
	origOpenFile := OsOpenFile
	defer func() { OsOpenFile = origOpenFile }()
	OsOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, os.ErrPermission
	}
	err = manager.Trust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Restore
	OsOpenFile = origOpenFile

	// Ensure file exists for subsequent tests
	manager.ensureTrustFileExists()

	// Test 3: OsOpen error in loadTrustedPaths
	origOpen := OsOpen
	defer func() { OsOpen = origOpen }()
	// Only fail when opening the trust config file
	OsOpen = func(name string) (*os.File, error) {
		if name == trustFilePath {
			return nil, os.ErrPermission
		}
		return origOpen(name)
	}
	err = manager.Trust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load trusted paths")

	paths, err := manager.List()
	assert.Error(t, err)
	assert.Nil(t, paths)

	// Restore
	OsOpen = origOpen

	// Test 4: calculateHash error (file doesn't exist)
	origAbs := FilepathAbs
	defer func() { FilepathAbs = origAbs }()
	FilepathAbs = func(path string) (string, error) {
		return "/nonexistent/dummy", nil
	}
	err = manager.Trust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to calculate file hash")

	// Test Untrust error due to loadTrustedPaths
	OsOpen = func(name string) (*os.File, error) {
		if name == trustFilePath {
			return nil, os.ErrPermission
		}
		return origOpen(name)
	}
	err = manager.Untrust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load trusted paths")
	OsOpen = origOpen

	// Test 5: TrustStatus FilepathAbs error
	FilepathAbs = func(path string) (string, error) {
		return "", os.ErrPermission
	}
	status := manager.TrustStatus(dummyPath)
	assert.Equal(t, TrustStatusUntrusted, status)

	err = manager.Trust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve absolute path")

	err = manager.Untrust(dummyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve absolute path")
	FilepathAbs = origAbs

	// Test 6: TrustStatus loadTrustedPaths error
	OsOpen = func(name string) (*os.File, error) {
		if name == trustFilePath {
			return nil, os.ErrPermission
		}
		return origOpen(name)
	}
	status = manager.TrustStatus(dummyPath)
	assert.Equal(t, TrustStatusUntrusted, status)
	OsOpen = origOpen

	// Test 7: Untrust with file not in trusted list
	err = manager.Untrust("/tmp/nonexistent")
	assert.NoError(t, err)

	// Test 8: SaveTrustedPaths write error
	manager.Trust(dummyPath) // Trust it first so Untrust will try to save

	// We can't easily mock bufio or os.File write methods, but we can make OsOpenFile fail on write
	OsOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		if name == trustFilePath && flag == os.O_WRONLY|os.O_TRUNC {
			return nil, os.ErrPermission
		}
		return origOpenFile(name, flag, perm)
	}
	err = manager.Untrust(dummyPath) // Untrust calls saveTrustedPaths
	assert.Error(t, err)
	OsOpenFile = origOpenFile

	// Test 9: calculateHash error
	OsOpen = func(name string) (*os.File, error) {
		return nil, os.ErrPermission
	}
	_, err = calculateHash(dummyPath)
	assert.Error(t, err)
	OsOpen = origOpen
}
