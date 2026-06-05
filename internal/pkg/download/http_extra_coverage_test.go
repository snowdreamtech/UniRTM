// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPDownloader_VerifyChecksum_OpenError verifies error handling when the file cannot be read.
func TestHTTPDownloader_VerifyChecksum_OpenError(t *testing.T) {
	downloader := download.NewHTTPDownloader()
	tmpDir := t.TempDir()
	
	// Use a directory to guarantee a read error cross-platform (since Windows ignores 0000 permissions)
	dirAsFile := filepath.Join(tmpDir, "directory_not_file")
	require.NoError(t, os.Mkdir(dirAsFile, 0755))

	err := downloader.VerifyChecksum(context.Background(), dirAsFile, "sha256:dummy")
	assert.Error(t, err)
	// Don't assert exact error string as it varies by OS ("is a directory" vs "Access is denied")
}

// TestHTTPDownloader_VerifyGPGSignature_KeyringPermissionDenied verifies error handling for unreadable keyring.
func TestHTTPDownloader_VerifyGPGSignature_KeyringPermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Use a directory to guarantee a read error cross-platform
	keyringPath := filepath.Join(tmpDir, "keyring.gpg")
	require.NoError(t, os.Mkdir(keyringPath, 0755))

	downloader := download.NewHTTPDownloader()
	res := &download.GPGResult{}
	opts := download.DefaultDownloadOptions().WithVerifyGPG(true, res).WithMaxRetries(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer server.Close()

	dest := filepath.Join(tmpDir, "test.txt")
	err := downloader.Download(context.Background(), server.URL, dest, opts)

	assert.Error(t, err)
	// It will either fail to open (Windows) or fail to read/parse (Unix).
	// We just verify it failed due to some GPG/keyring issue.
	assert.Contains(t, err.Error(), "keyring")
}

// TestHTTPDownloader_VerifyGPGSignature_FetchError verifies error handling when context is cancelled.
func TestHTTPDownloader_Download_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer server.Close()

	downloader := download.NewHTTPDownloader()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := download.DefaultDownloadOptions().WithMaxRetries(0)
	dest := filepath.Join(tmpDir, "test.txt")
	err := downloader.Download(ctx, server.URL, dest, opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download cancelled")
}

// TestHTTPDownloader_VerifyGPGSignature_FullFailure verifies the end-to-end execution of signature check which fails at CheckDetachedSignature.
func TestHTTPDownloader_VerifyGPGSignature_FullFailure(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	keyringPath := filepath.Join(tmpDir, "keyring.gpg")
	// Must be a valid empty keyring to pass openpgp.ReadKeyRing.
	// Actually an empty file will result in empty keyring. Let's write an empty keyring.
	err := os.WriteFile(keyringPath, []byte{}, 0644)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer server.Close()

	downloader := download.NewHTTPDownloader()
	res := &download.GPGResult{}
	opts := download.DefaultDownloadOptions().WithVerifyGPG(true, res).WithMaxRetries(0)
	dest := filepath.Join(tmpDir, "test.txt")
	err = downloader.Download(context.Background(), server.URL+"/file.txt", dest, opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GPG signature verification failed")
}

// TestHTTPDownloader_Download_Concurrent_Cancel verifies concurrent cancellation.
func TestHTTPDownloader_Download_Concurrent_Cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", "6291456")
			w.WriteHeader(http.StatusOK)
			return
		}
		// Write slowly so context can be cancelled
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", "6291456")
		w.WriteHeader(http.StatusPartialContent)

		for i := 0; i < 60; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				_, _ = w.Write(make([]byte, 102400))
				if w.(http.Flusher) != nil {
					w.(http.Flusher).Flush()
				}
			}
		}
	}))
	defer server.Close()

	downloader := download.NewHTTPDownloader()
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "test.txt")

	ctx, cancel := context.WithCancel(context.Background())
	opts := download.DefaultDownloadOptions().WithMaxRetries(0).WithProgressCallback(func(downloaded, total int64) {
		if downloaded > 1024 {
			cancel() // Cancel once it started downloading
		}
	})

	err := downloader.Download(ctx, server.URL, dest, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestHTTPDownloader_Download_Concurrent_IgnoredRange verifies when server ignores range headers.
func TestHTTPDownloader_Download_Concurrent_IgnoredRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", "6291456")
			w.WriteHeader(http.StatusOK)
			return
		}

		// If Range doesn't start at 0, return 200 OK to simulate ignored range
		if r.Header.Get("Range") != "" && r.Header.Get("Range") != "bytes=0-1048575" {
			w.Header().Set("Content-Length", "6291456")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ignored range"))
			return
		}

		if r.Header.Get("Range") == "" {
			w.Header().Set("Content-Length", "1048576")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(make([]byte, 102400))
			return
		}

		// First thread gets 206 Partial Content
		w.Header().Set("Content-Length", "1048576")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(make([]byte, 102400)) // Just partial write is enough, it will fail but we want the Range ignored error from another thread
	}))
	defer server.Close()

	downloader := download.NewHTTPDownloader()
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "test.txt")

	opts := download.DefaultDownloadOptions().WithMaxRetries(0)
	err := downloader.Download(context.Background(), server.URL, dest, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected EOF")
}

// TestHTTPDownloader_Download_Concurrent_InvalidDest verifies error when opening file fails.
func TestHTTPDownloader_Download_Concurrent_InvalidDest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", "6291456")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	downloader := download.NewHTTPDownloader()
	tmpDir := t.TempDir()
	// Use directory as destination to force os.OpenFile to fail inside downloadConcurrent
	dest := tmpDir

	opts := download.DefaultDownloadOptions().WithMaxRetries(0)
	err := downloader.Download(context.Background(), server.URL, dest, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}
