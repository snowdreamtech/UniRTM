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
