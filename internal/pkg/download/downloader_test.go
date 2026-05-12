// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download_test

import (
	"context"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultDownloadOptions verifies that DefaultDownloadOptions returns sensible defaults.
func TestDefaultDownloadOptions(t *testing.T) {
	opts := download.DefaultDownloadOptions()

	assert.Equal(t, 5, opts.MaxRetries, "MaxRetries should default to 5")
	assert.Equal(t, 15*time.Minute, opts.Timeout, "Timeout should default to 5 minutes")
	assert.Empty(t, opts.Checksum, "Checksum should be empty by default")
	assert.Nil(t, opts.ProgressCallback, "ProgressCallback should be nil by default")
}

// TestDownloadOptions_WithChecksum verifies the WithChecksum fluent method.
func TestDownloadOptions_WithChecksum(t *testing.T) {
	opts := download.DefaultDownloadOptions()
	checksum := "sha256:abc123def456"

	newOpts := opts.WithChecksum(checksum)

	assert.Equal(t, checksum, newOpts.Checksum, "Checksum should be set")
	assert.Empty(t, opts.Checksum, "Original options should be unchanged")
}

// TestDownloadOptions_WithMaxRetries verifies the WithMaxRetries fluent method.
func TestDownloadOptions_WithMaxRetries(t *testing.T) {
	opts := download.DefaultDownloadOptions()
	retries := 3

	newOpts := opts.WithMaxRetries(retries)

	assert.Equal(t, retries, newOpts.MaxRetries, "MaxRetries should be set")
	assert.Equal(t, 5, opts.MaxRetries, "Original options should be unchanged")
}

// TestDownloadOptions_WithTimeout verifies the WithTimeout fluent method.
func TestDownloadOptions_WithTimeout(t *testing.T) {
	opts := download.DefaultDownloadOptions()
	timeout := 10 * time.Second

	newOpts := opts.WithTimeout(timeout)

	assert.Equal(t, timeout, newOpts.Timeout, "Timeout should be set")
	assert.Equal(t, 15*time.Minute, opts.Timeout, "Original options should be unchanged")
}

// TestDownloadOptions_WithProgressCallback verifies the WithProgressCallback fluent method.
func TestDownloadOptions_WithProgressCallback(t *testing.T) {
	opts := download.DefaultDownloadOptions()
	called := false
	callback := func(downloaded, total int64) {
		called = true
	}

	newOpts := opts.WithProgressCallback(callback)

	require.NotNil(t, newOpts.ProgressCallback, "ProgressCallback should be set")
	newOpts.ProgressCallback(100, 1000)
	assert.True(t, called, "Callback should be invoked")
	assert.Nil(t, opts.ProgressCallback, "Original options should be unchanged")
}

// TestDownloadOptions_FluentChaining verifies that fluent methods can be chained.
func TestDownloadOptions_FluentChaining(t *testing.T) {
	progressCalled := false
	callback := func(downloaded, total int64) {
		progressCalled = true
	}

	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123").
		WithMaxRetries(3).
		WithTimeout(30 * time.Second).
		WithProgressCallback(callback)

	assert.Equal(t, "sha256:abc123", opts.Checksum)
	assert.Equal(t, 3, opts.MaxRetries)
	assert.Equal(t, 30*time.Second, opts.Timeout)
	require.NotNil(t, opts.ProgressCallback)

	opts.ProgressCallback(50, 100)
	assert.True(t, progressCalled, "Callback should be invoked")
}

// MockDownloader is a mock implementation of the Downloader interface for testing.
type MockDownloader struct {
	DownloadFunc       func(ctx context.Context, url, destination string, opts download.DownloadOptions) error
	VerifyChecksumFunc func(ctx context.Context, file, expectedChecksum string) error
}

// Download implements the Downloader interface.
func (m *MockDownloader) Download(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, url, destination, opts)
	}
	return nil
}

// VerifyChecksum implements the Downloader interface.
func (m *MockDownloader) VerifyChecksum(ctx context.Context, file, expectedChecksum string) error {
	if m.VerifyChecksumFunc != nil {
		return m.VerifyChecksumFunc(ctx, file, expectedChecksum)
	}
	return nil
}

// TestMockDownloader_Download verifies the mock downloader can be used in tests.
func TestMockDownloader_Download(t *testing.T) {
	downloadCalled := false
	mock := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			downloadCalled = true
			assert.Equal(t, "https://example.com/file.tar.gz", url)
			assert.Equal(t, "/tmp/file.tar.gz", destination)
			assert.Equal(t, 3, opts.MaxRetries)
			return nil
		},
	}

	opts := download.DefaultDownloadOptions().WithMaxRetries(3)
	err := mock.Download(context.Background(), "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)

	require.NoError(t, err)
	assert.True(t, downloadCalled, "Download should be called")
}

// TestMockDownloader_VerifyChecksum verifies the mock downloader checksum verification.
func TestMockDownloader_VerifyChecksum(t *testing.T) {
	verifyCalled := false
	mock := &MockDownloader{
		VerifyChecksumFunc: func(ctx context.Context, file, expectedChecksum string) error {
			verifyCalled = true
			assert.Equal(t, "/tmp/file.tar.gz", file)
			assert.Equal(t, "sha256:abc123", expectedChecksum)
			return nil
		},
	}

	err := mock.VerifyChecksum(context.Background(), "/tmp/file.tar.gz", "sha256:abc123")

	require.NoError(t, err)
	assert.True(t, verifyCalled, "VerifyChecksum should be called")
}

// TestDownloader_ContextCancellation verifies that implementations respect context cancellation.
func TestDownloader_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mock := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}

	opts := download.DefaultDownloadOptions()
	err := mock.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled, "Should return context.Canceled error")
}

// TestDownloader_ContextTimeout verifies that implementations respect context deadlines.
func TestDownloader_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(10 * time.Millisecond)

	mock := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}

	opts := download.DefaultDownloadOptions()
	err := mock.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Should return context.DeadlineExceeded error")
}

// TestDownloadOptions_ZeroValues verifies behavior with zero values.
func TestDownloadOptions_ZeroValues(t *testing.T) {
	var opts download.DownloadOptions

	assert.Equal(t, 0, opts.MaxRetries, "MaxRetries should be zero")
	assert.Equal(t, time.Duration(0), opts.Timeout, "Timeout should be zero")
	assert.Empty(t, opts.Checksum, "Checksum should be empty")
	assert.Nil(t, opts.ProgressCallback, "ProgressCallback should be nil")
}

// TestDownloadOptions_ProgressCallbackNil verifies that nil progress callback is safe.
func TestDownloadOptions_ProgressCallbackNil(t *testing.T) {
	opts := download.DefaultDownloadOptions()

	// Should not panic when callback is nil
	assert.NotPanics(t, func() {
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(100, 1000)
		}
	})
}

// TestDownloadOptions_ProgressCallbackInvocation verifies progress callback is invoked correctly.
func TestDownloadOptions_ProgressCallbackInvocation(t *testing.T) {
	var downloaded, total int64
	callback := func(d, t int64) {
		downloaded = d
		total = t
	}

	opts := download.DefaultDownloadOptions().WithProgressCallback(callback)

	opts.ProgressCallback(500, 1000)

	assert.Equal(t, int64(500), downloaded, "Downloaded bytes should be 500")
	assert.Equal(t, int64(1000), total, "Total bytes should be 1000")
}

// TestDownloadOptions_MultipleProgressCallbacks verifies multiple progress updates.
func TestDownloadOptions_MultipleProgressCallbacks(t *testing.T) {
	var updates []int64
	callback := func(downloaded, total int64) {
		updates = append(updates, downloaded)
	}

	opts := download.DefaultDownloadOptions().WithProgressCallback(callback)

	// Simulate multiple progress updates
	opts.ProgressCallback(100, 1000)
	opts.ProgressCallback(500, 1000)
	opts.ProgressCallback(1000, 1000)

	require.Len(t, updates, 3, "Should have 3 progress updates")
	assert.Equal(t, int64(100), updates[0])
	assert.Equal(t, int64(500), updates[1])
	assert.Equal(t, int64(1000), updates[2])
}
