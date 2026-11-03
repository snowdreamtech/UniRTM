// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPDownloader_VerifyGPG_NoKeyring(t *testing.T) {
	downloader := NewHTTPDownloader()

	// Set a custom env dir so keyring.gpg doesn't exist
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	err := downloader.verifyGPGSignature(context.Background(), "http://example.com/file", "/tmp/file")
	require.Error(t, err)
	require.Contains(t, err.Error(), "keyring not found")
}

func TestHTTPDownloader_VerifyGPG_InvalidKeyring(t *testing.T) {
	downloader := NewHTTPDownloader()

	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Create an invalid keyring.gpg
	keyringPath := filepath.Join(tmpDir, "keyring.gpg")
	err := os.WriteFile(keyringPath, []byte("invalid data"), 0644)
	require.NoError(t, err)

	err = downloader.verifyGPGSignature(context.Background(), "http://example.com/file", "/tmp/file")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse keyring")
}
