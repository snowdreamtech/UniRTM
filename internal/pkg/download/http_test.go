// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPDownloader_NewHTTPDownloader verifies that NewHTTPDownloader creates a valid instance.
func TestHTTPDownloader_NewHTTPDownloader(t *testing.T) {
	downloader := download.NewHTTPDownloader()
	require.NotNil(t, downloader, "NewHTTPDownloader should return a non-nil instance")
}

// TestHTTPDownloader_Download_Success verifies successful download.
func TestHTTPDownloader_Download_Success(t *testing.T) {
	// Create test server
	content := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions()
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.NoError(t, err, "Download should succeed")

	// Verify file content
	downloaded, err := os.ReadFile(destination)
	require.NoError(t, err, "Should be able to read downloaded file")
	assert.Equal(t, content, downloaded, "Downloaded content should match")
}

// TestHTTPDownloader_Download_WithChecksum verifies download with checksum verification.
func TestHTTPDownloader_Download_WithChecksum(t *testing.T) {
	// Create test server
	content := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Compute expected checksum
	hasher := sha256.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with checksum
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithChecksum("sha256:" + expectedChecksum)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.NoError(t, err, "Download with valid checksum should succeed")

	// Verify file exists
	_, err = os.Stat(destination)
	require.NoError(t, err, "Downloaded file should exist")
}

// TestHTTPDownloader_Download_ChecksumMismatch verifies checksum mismatch handling.
func TestHTTPDownloader_Download_ChecksumMismatch(t *testing.T) {
	// Create test server
	content := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Use wrong checksum
	wrongChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with wrong checksum
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithChecksum(wrongChecksum)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.Error(t, err, "Download with wrong checksum should fail")
	assert.ErrorIs(t, err, errors.ErrChecksumMismatch, "Should return ErrChecksumMismatch")

	// Verify file was deleted
	_, err = os.Stat(destination)
	assert.True(t, os.IsNotExist(err), "File should be deleted after checksum mismatch")
}

// TestHTTPDownloader_Download_WithProgress verifies progress reporting.
func TestHTTPDownloader_Download_WithProgress(t *testing.T) {
	// Create test server
	content := []byte(strings.Repeat("x", 10000)) // 10KB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Track progress
	var progressUpdates []int64
	callback := func(downloaded, total int64) {
		progressUpdates = append(progressUpdates, downloaded)
	}

	// Download file with progress callback
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithProgressCallback(callback)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.NoError(t, err, "Download should succeed")
	assert.NotEmpty(t, progressUpdates, "Progress callback should be called")
	assert.Equal(t, int64(len(content)), progressUpdates[len(progressUpdates)-1], "Final progress should equal content length")
}

// TestHTTPDownloader_Download_ContextCancellation verifies context cancellation handling.
func TestHTTPDownloader_Download_ContextCancellation(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Download file with cancelled context
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions()
	err := downloader.Download(ctx, server.URL, destination, opts)

	require.Error(t, err, "Download with cancelled context should fail")
	assert.True(t, errors.IsExternalError(err), "Should be an external error")
}

// TestHTTPDownloader_Download_ContextTimeout verifies context timeout handling.
func TestHTTPDownloader_Download_ContextTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Download file with timeout
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions()
	err := downloader.Download(ctx, server.URL, destination, opts)

	require.Error(t, err, "Download with timeout should fail")
	assert.True(t, errors.IsExternalError(err), "Should be an external error")
}

// TestHTTPDownloader_Download_HTTPError verifies HTTP error handling.
func TestHTTPDownloader_Download_HTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"403 Forbidden", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that returns error
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Create temporary destination
			tmpDir := t.TempDir()
			destination := filepath.Join(tmpDir, "test.txt")

			// Download file
			downloader := download.NewHTTPDownloader()
			opts := download.DefaultDownloadOptions()
			err := downloader.Download(context.Background(), server.URL, destination, opts)

			require.Error(t, err, "Download should fail with HTTP error")
			assert.True(t, errors.IsExternalError(err), "Should be an external error")
		})
	}
}

// TestHTTPDownloader_Download_InvalidURL verifies invalid URL handling.
func TestHTTPDownloader_Download_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"invalid scheme", "ftp://example.com/file.txt"},
		{"no host", "http:///file.txt"},
		{"malformed URL", "ht!tp://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary destination
			tmpDir := t.TempDir()
			destination := filepath.Join(tmpDir, "test.txt")

			// Download file with invalid URL
			downloader := download.NewHTTPDownloader()
			opts := download.DefaultDownloadOptions()
			err := downloader.Download(context.Background(), tt.url, destination, opts)

			require.Error(t, err, "Download with invalid URL should fail")
			assert.True(t, errors.IsUserError(err), "Should be a user error")
		})
	}
}

// TestHTTPDownloader_Download_RetryLogic verifies retry with exponential backoff.
func TestHTTPDownloader_Download_RetryLogic(t *testing.T) {
	// Track attempts
	attempts := 0
	maxAttempts := 3

	// Create test server that fails first attempts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < maxAttempts {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with retries
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithMaxRetries(5)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.NoError(t, err, "Download should succeed after retries")
	assert.Equal(t, maxAttempts, attempts, "Should retry until success")

	// Verify file content
	downloaded, err := os.ReadFile(destination)
	require.NoError(t, err, "Should be able to read downloaded file")
	assert.Equal(t, []byte("success"), downloaded, "Downloaded content should match")
}

// TestHTTPDownloader_Download_RetryExhausted verifies behavior when retries are exhausted.
func TestHTTPDownloader_Download_RetryExhausted(t *testing.T) {
	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with limited retries
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithMaxRetries(2)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.Error(t, err, "Download should fail after exhausting retries")
	assert.True(t, errors.IsExternalError(err), "Should be an external error")
	assert.Contains(t, err.Error(), "3 attempts", "Error should mention number of attempts")
}

// TestHTTPDownloader_Download_NoRetryOnUserError verifies that user errors are not retried.
func TestHTTPDownloader_Download_NoRetryOnUserError(t *testing.T) {
	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with invalid URL (user error)
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithMaxRetries(5)
	err := downloader.Download(context.Background(), "ftp://invalid.com", destination, opts)

	require.Error(t, err, "Download should fail immediately")
	assert.True(t, errors.IsUserError(err), "Should be a user error")
}

// TestHTTPDownloader_VerifyChecksum_Success verifies successful checksum verification.
func TestHTTPDownloader_VerifyChecksum_Success(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(file, content, 0644)
	require.NoError(t, err)

	// Compute expected checksum
	hasher := sha256.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Verify checksum
	downloader := download.NewHTTPDownloader()
	err = downloader.VerifyChecksum(context.Background(), file, "sha256:"+expectedChecksum)

	require.NoError(t, err, "Checksum verification should succeed")

	// Verify file still exists
	_, err = os.Stat(file)
	require.NoError(t, err, "File should still exist after successful verification")
}

// TestHTTPDownloader_VerifyChecksum_WithoutAlgorithm verifies checksum without algorithm prefix.
func TestHTTPDownloader_VerifyChecksum_WithoutAlgorithm(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(file, content, 0644)
	require.NoError(t, err)

	// Compute expected checksum
	hasher := sha256.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Verify checksum without algorithm prefix (should assume SHA-256)
	downloader := download.NewHTTPDownloader()
	err = downloader.VerifyChecksum(context.Background(), file, expectedChecksum)

	require.NoError(t, err, "Checksum verification should succeed without algorithm prefix")
}

// TestHTTPDownloader_VerifyChecksum_Mismatch verifies checksum mismatch handling.
func TestHTTPDownloader_VerifyChecksum_Mismatch(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(file, content, 0644)
	require.NoError(t, err)

	// Use wrong checksum
	wrongChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	// Verify checksum
	downloader := download.NewHTTPDownloader()
	err = downloader.VerifyChecksum(context.Background(), file, wrongChecksum)

	require.Error(t, err, "Checksum verification should fail")
	assert.ErrorIs(t, err, errors.ErrChecksumMismatch, "Should return ErrChecksumMismatch")

	// Verify file was deleted
	_, err = os.Stat(file)
	assert.True(t, os.IsNotExist(err), "File should be deleted after checksum mismatch")
}

// TestHTTPDownloader_VerifyChecksum_InvalidFormat verifies invalid checksum format handling.
func TestHTTPDownloader_VerifyChecksum_InvalidFormat(t *testing.T) {
	tests := []struct {
		name     string
		checksum string
	}{
		{"empty checksum", ""},
		{"invalid format", "sha256:"},
		{"unsupported algorithm", "md5:abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			file := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(file, []byte("content"), 0644)
			require.NoError(t, err)

			// Verify checksum with invalid format
			downloader := download.NewHTTPDownloader()
			err = downloader.VerifyChecksum(context.Background(), file, tt.checksum)

			require.Error(t, err, "Checksum verification should fail with invalid format")
			assert.True(t, errors.IsUserError(err), "Should be a user error")
		})
	}
}

// TestHTTPDownloader_VerifyChecksum_FileNotFound verifies file not found handling.
func TestHTTPDownloader_VerifyChecksum_FileNotFound(t *testing.T) {
	// Use non-existent file
	file := "/nonexistent/file.txt"
	checksum := "sha256:abc123"

	// Verify checksum
	downloader := download.NewHTTPDownloader()
	err := downloader.VerifyChecksum(context.Background(), file, checksum)

	require.Error(t, err, "Checksum verification should fail for non-existent file")
	assert.True(t, errors.IsSystemError(err), "Should be a system error")
}

// TestHTTPDownloader_Download_OptionsTimeout verifies timeout from options.
func TestHTTPDownloader_Download_OptionsTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file with short timeout in options
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithTimeout(50 * time.Millisecond)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.Error(t, err, "Download should fail with timeout")
	assert.True(t, errors.IsExternalError(err), "Should be an external error")
}

// TestHTTPDownloader_Download_CleanupOnFailure verifies partial download cleanup.
func TestHTTPDownloader_Download_CleanupOnFailure(t *testing.T) {
	// Create test server that fails mid-download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		// Connection closes abruptly
	}))
	defer server.Close()

	// Create temporary destination
	tmpDir := t.TempDir()
	destination := filepath.Join(tmpDir, "test.txt")

	// Download file (should fail)
	downloader := download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().WithMaxRetries(0)
	err := downloader.Download(context.Background(), server.URL, destination, opts)

	require.Error(t, err, "Download should fail")

	// Verify file was cleaned up
	_, err = os.Stat(destination)
	assert.True(t, os.IsNotExist(err), "Partial download should be cleaned up")
}
