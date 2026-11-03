// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download_test

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/pkg/download"
)

// ExampleNewHTTPDownloader demonstrates creating a new HTTP downloader.
func ExampleNewHTTPDownloader() {
	downloader := download.NewHTTPDownloader()
	fmt.Printf("Created HTTP downloader: %T\n", downloader)
	// Output: Created HTTP downloader: *download.HTTPDownloader
}

// ExampleHTTPDownloader_Download demonstrates basic HTTP download.
func ExampleHTTPDownloader_Download() {
	_ = download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions()

	// In a real scenario, you would use an actual URL and destination
	// err := downloader.Download(context.Background(), "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)
	// if err != nil {
	//     log.Fatal(err)
	// }

	fmt.Printf("Downloader ready with options: MaxRetries=%d, Timeout=%v\n", opts.MaxRetries, opts.Timeout)
	// Output: Downloader ready with options: MaxRetries=5, Timeout=15m0s
}

// ExampleHTTPDownloader_Download_withChecksum demonstrates download with checksum verification.
func ExampleHTTPDownloader_Download_withChecksum() {
	_ = download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123def456...")

	// In a real scenario:
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     if errors.Is(err, errors.ErrChecksumMismatch) {
	//         log.Fatal("Checksum verification failed")
	//     }
	//     log.Fatal(err)
	// }

	fmt.Printf("Download with checksum: %s\n", opts.Checksum)
	// Output: Download with checksum: sha256:abc123def456...
}

// ExampleHTTPDownloader_Download_withProgress demonstrates download with progress reporting.
func ExampleHTTPDownloader_Download_withProgress() {
	_ = download.NewHTTPDownloader()

	progressCallback := func(downloaded, total int64) {
		if total > 0 {
			percent := float64(downloaded) / float64(total) * 100
			fmt.Printf("Progress: %.1f%%\n", percent)
		}
	}

	opts := download.DefaultDownloadOptions().
		WithProgressCallback(progressCallback)

	// In a real scenario:
	// err := downloader.Download(ctx, url, destination, opts)

	fmt.Printf("Progress callback configured: %v\n", opts.ProgressCallback != nil)
	// Output: Progress callback configured: true
}

// ExampleHTTPDownloader_Download_withRetry demonstrates download with custom retry configuration.
func ExampleHTTPDownloader_Download_withRetry() {
	_ = download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().
		WithMaxRetries(3) // Retry up to 3 times (4 total attempts)

	// In a real scenario:
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     log.Printf("Download failed after %d attempts: %v", opts.MaxRetries+1, err)
	// }

	fmt.Printf("Max retries: %d (total attempts: %d)\n", opts.MaxRetries, opts.MaxRetries+1)
	// Output: Max retries: 3 (total attempts: 4)
}

// ExampleHTTPDownloader_Download_withTimeout demonstrates download with custom timeout.
func ExampleHTTPDownloader_Download_withTimeout() {
	_ = download.NewHTTPDownloader()
	opts := download.DefaultDownloadOptions().
		WithTimeout(30 * 1000000000) // 30 seconds in nanoseconds

	// In a real scenario:
	// ctx := context.Background()
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     if errors.Is(err, context.DeadlineExceeded) {
	//         log.Fatal("Download timed out")
	//     }
	//     log.Fatal(err)
	// }

	fmt.Printf("Timeout: %v\n", opts.Timeout)
	// Output: Timeout: 30s
}

// ExampleHTTPDownloader_Download_complete demonstrates a complete download with all options.
func ExampleHTTPDownloader_Download_complete() {
	_ = download.NewHTTPDownloader()

	// Configure all options
	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:abc123...").
		WithMaxRetries(3).
		WithTimeout(60 * 1000000000). // 60 seconds
		WithProgressCallback(func(downloaded, total int64) {
			// Progress reporting
		})

	// In a real scenario:
	// ctx := context.Background()
	// err := downloader.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// log.Println("Download completed successfully")

	fmt.Printf("Configured: checksum=%v, retries=%d, timeout=%v, progress=%v\n",
		opts.Checksum != "", opts.MaxRetries, opts.Timeout, opts.ProgressCallback != nil)
	// Output: Configured: checksum=true, retries=3, timeout=1m0s, progress=true
}

// ExampleHTTPDownloader_VerifyChecksum demonstrates standalone checksum verification.
func ExampleHTTPDownloader_VerifyChecksum() {
	_ = download.NewHTTPDownloader()

	// In a real scenario:
	// err := downloader.VerifyChecksum(context.Background(), "/tmp/file.tar.gz", "sha256:abc123...")
	// if err != nil {
	//     if errors.Is(err, errors.ErrChecksumMismatch) {
	//         log.Fatal("Checksum verification failed - file may be corrupted")
	//     }
	//     log.Fatal(err)
	// }
	// log.Println("Checksum verification passed")

	fmt.Println("Checksum verification ready")
	// Output: Checksum verification ready
}

// ExampleHTTPDownloader_Download_contextCancellation demonstrates context cancellation.
func ExampleHTTPDownloader_Download_contextCancellation() {
	_ = download.NewHTTPDownloader()
	_ = download.DefaultDownloadOptions()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately for demonstration
	cancel()

	// In a real scenario:
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     if errors.Is(err, context.Canceled) {
	//         log.Println("Download was cancelled")
	//     }
	// }

	fmt.Printf("Context cancellation configured: %v\n", ctx.Err() != nil)
	// Output: Context cancellation configured: true
}

// ExampleHTTPDownloader_Download_proxySupport demonstrates proxy configuration.
func ExampleHTTPDownloader_Download_proxySupport() {
	// The HTTP downloader automatically respects HTTP_PROXY and HTTPS_PROXY
	// environment variables. No additional configuration is needed.

	// In a real scenario, set environment variables:
	// os.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
	// os.Setenv("HTTPS_PROXY", "https://proxy.example.com:8443")

	downloader := download.NewHTTPDownloader()
	_ = download.DefaultDownloadOptions()

	// The downloader will automatically use the proxy
	// err := downloader.Download(ctx, url, destination, opts)

	fmt.Printf("Downloader created with proxy support: %T\n", downloader)
	// Output: Downloader created with proxy support: *download.HTTPDownloader
}

// ExampleHTTPDownloader_Download_errorHandling demonstrates comprehensive error handling.
func ExampleHTTPDownloader_Download_errorHandling() {
	downloader := download.NewHTTPDownloader()
	_ = download.DefaultDownloadOptions()

	// In a real scenario:
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     switch {
	//     case errors.IsUserError(err):
	//         log.Printf("User error (invalid input): %v", err)
	//     case errors.IsSystemError(err):
	//         log.Printf("System error (disk/permissions): %v", err)
	//     case errors.IsExternalError(err):
	//         log.Printf("External error (network/server): %v", err)
	//     default:
	//         log.Printf("Unknown error: %v", err)
	//     }
	//     os.Exit(errors.ExitCode(err))
	// }

	fmt.Printf("Error handling configured for downloader: %T\n", downloader)
	// Output: Error handling configured for downloader: *download.HTTPDownloader
}

// ExampleHTTPDownloader_Download_realWorld demonstrates a real-world usage pattern.
func ExampleHTTPDownloader_Download_realWorld() {
	// This example shows a complete real-world download workflow
	_ = download.NewHTTPDownloader()

	// Configure options
	opts := download.DefaultDownloadOptions().
		WithChecksum("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855").
		WithMaxRetries(5).
		WithProgressCallback(func(downloaded, total int64) {
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				log.Printf("Download progress: %.1f%% (%d/%d bytes)", percent, downloaded, total)
			}
		})

	// In a real scenario:
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()
	//
	// url := "https://example.com/releases/v1.0.0/tool.tar.gz"
	// destination := "/tmp/tool.tar.gz"
	//
	// log.Printf("Downloading %s to %s...", url, destination)
	// err := downloader.Download(ctx, url, destination, opts)
	// if err != nil {
	//     log.Fatalf("Download failed: %v", err)
	// }
	// log.Println("Download completed successfully")

	fmt.Printf("Real-world download configured: retries=%d, checksum=%v\n",
		opts.MaxRetries, opts.Checksum != "")
	// Output: Real-world download configured: retries=5, checksum=true
}
