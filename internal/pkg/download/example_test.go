// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download_test

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
)

// ExampleDefaultDownloadOptions demonstrates creating download options with defaults.
func ExampleDefaultDownloadOptions() {
	opts := download.DefaultDownloadOptions()

	fmt.Printf("MaxRetries: %d\n", opts.MaxRetries)
	fmt.Printf("Timeout: %v\n", opts.Timeout)
	fmt.Printf("Has Checksum: %v\n", opts.Checksum != "")
	fmt.Printf("Has Callback: %v\n", opts.ProgressCallback != nil)

	// Output:
	// MaxRetries: 5
	// Timeout: 15m0s
	// Has Checksum: false
	// Has Callback: false
}

// ExampleDownloadOptions_WithChecksum demonstrates setting a checksum.
func ExampleDownloadOptions_WithChecksum() {
	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123def456789")

	fmt.Printf("Checksum: %s\n", opts.Checksum)

	// Output:
	// Checksum: sha256:abc123def456789
}

// ExampleDownloadOptions_WithMaxRetries demonstrates configuring retry behavior.
func ExampleDownloadOptions_WithMaxRetries() {
	opts := download.DefaultDownloadOptions().
		WithMaxRetries(3)

	fmt.Printf("MaxRetries: %d\n", opts.MaxRetries)

	// Output:
	// MaxRetries: 3
}

// ExampleDownloadOptions_fluent demonstrates fluent configuration chaining.
func ExampleDownloadOptions_fluent() {
	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123").
		WithMaxRetries(3).
		WithProgressCallback(func(downloaded, total int64) {
			// Progress callback implementation
		})

	fmt.Printf("Checksum: %s\n", opts.Checksum)
	fmt.Printf("MaxRetries: %d\n", opts.MaxRetries)
	fmt.Printf("Has Callback: %v\n", opts.ProgressCallback != nil)

	// Output:
	// Checksum: sha256:abc123
	// MaxRetries: 3
	// Has Callback: true
}

// ExampleDownloader_basic demonstrates basic download usage.
func ExampleDownloader_basic() {
	// Create a mock downloader for demonstration
	downloader := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			fmt.Printf("Downloading %s to %s\n", url, destination)
			return nil
		},
	}

	// Configure download options
	opts := download.DefaultDownloadOptions()

	// Perform download
	err := downloader.Download(
		context.Background(),
		"https://example.com/tool-v1.0.0.tar.gz",
		"/tmp/tool-v1.0.0.tar.gz",
		opts,
	)

	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Println("Download completed successfully")

	// Output:
	// Downloading https://example.com/tool-v1.0.0.tar.gz to /tmp/tool-v1.0.0.tar.gz
	// Download completed successfully
}

// ExampleDownloader_withChecksum demonstrates download with checksum verification.
func ExampleDownloader_withChecksum() {
	downloader := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			fmt.Printf("Downloading with checksum: %s\n", opts.Checksum)
			return nil
		},
	}

	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123def456789")

	err := downloader.Download(
		context.Background(),
		"https://example.com/tool.tar.gz",
		"/tmp/tool.tar.gz",
		opts,
	)

	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Println("Download and verification completed")

	// Output:
	// Downloading with checksum: sha256:abc123def456789
	// Download and verification completed
}

// ExampleDownloader_withProgress demonstrates download with progress reporting.
func ExampleDownloader_withProgress() {
	downloader := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			// Simulate progress updates
			if opts.ProgressCallback != nil {
				opts.ProgressCallback(0, 1000)
				opts.ProgressCallback(500, 1000)
				opts.ProgressCallback(1000, 1000)
			}
			return nil
		},
	}

	opts := download.DefaultDownloadOptions().
		WithProgressCallback(func(downloaded, total int64) {
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("Progress: %.0f%%\n", percent)
			}
		})

	err := downloader.Download(
		context.Background(),
		"https://example.com/large-file.tar.gz",
		"/tmp/large-file.tar.gz",
		opts,
	)

	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	// Output:
	// Progress: 0%
	// Progress: 50%
	// Progress: 100%
}

// ExampleDownloader_VerifyChecksum demonstrates standalone checksum verification.
func ExampleDownloader_VerifyChecksum() {
	downloader := &MockDownloader{
		VerifyChecksumFunc: func(ctx context.Context, file, expectedChecksum string) error {
			fmt.Printf("Verifying %s with checksum %s\n", file, expectedChecksum)
			return nil
		},
	}

	err := downloader.VerifyChecksum(
		context.Background(),
		"/tmp/tool-v1.0.0.tar.gz",
		"sha256:abc123def456789",
	)

	if err != nil {
		log.Fatalf("Checksum verification failed: %v", err)
	}

	fmt.Println("Checksum verification passed")

	// Output:
	// Verifying /tmp/tool-v1.0.0.tar.gz with checksum sha256:abc123def456789
	// Checksum verification passed
}

// ExampleDownloader_contextCancellation demonstrates context cancellation handling.
func ExampleDownloader_contextCancellation() {
	downloader := &MockDownloader{
		DownloadFunc: func(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := download.DefaultDownloadOptions()
	err := downloader.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)

	if err != nil {
		fmt.Printf("Download cancelled: %v\n", err)
	}

	// Output:
	// Download cancelled: context canceled
}
