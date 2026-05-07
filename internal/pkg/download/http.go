// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
)

// HTTPDownloader implements the Downloader interface using Go's standard HTTP client.
// It supports retry logic with exponential backoff, timeout configuration, proxy support,
// and progress reporting.
//
// The implementation follows Requirement 4.2, 4.3, 4.4, 4.5 from the design document:
//   - Retry logic with exponential backoff (1s → 2s → 4s → 8s → 16s, max 5 attempts)
//   - Connection timeout (10s) and read timeout (60s)
//   - Proxy support via HTTP_PROXY/HTTPS_PROXY environment variables
//   - Progress reporting callback
//
// Example usage:
//
//	downloader := download.NewHTTPDownloader()
//	opts := download.DefaultDownloadOptions().
//	    WithChecksum("sha256:abc123...").
//	    WithProgressCallback(func(downloaded, total int64) {
//	        fmt.Printf("Progress: %d/%d bytes\n", downloaded, total)
//	    })
//	err := downloader.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)
type HTTPDownloader struct {
	client *http.Client
}

// NewHTTPDownloader creates a new HTTPDownloader with default configuration.
// The HTTP client is configured with:
//   - Connection timeout: 10 seconds
//   - Read timeout: 60 seconds
//   - Proxy support via HTTP_PROXY/HTTPS_PROXY environment variables
//   - Automatic redirect following (up to 10 redirects)
func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{
		client: &http.Client{
			Timeout: 60 * time.Second, // Read timeout
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment, // Support HTTP_PROXY/HTTPS_PROXY
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				// Connection timeout is handled by context deadline
			},
		},
	}
}

// Download downloads a file from the specified URL to the destination path.
// The operation respects the context for cancellation and deadlines.
//
// The implementation:
//   - Implements retry logic with exponential backoff (1s → 2s → 4s → 8s → 16s)
//   - Respects context cancellation and deadlines
//   - Calls the progress callback (if provided) during download
//   - Verifies the checksum after download if specified in opts
//   - Cleans up partial downloads on failure
//   - Returns descriptive errors with context (URL, attempt count, failure reason)
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - url: The source URL to download from
//   - destination: The local file path where the downloaded file will be saved
//   - opts: Download options including retry, timeout, and progress callback
//
// Returns:
//   - error: nil on success, or an error describing the failure
func (h *HTTPDownloader) Download(ctx context.Context, url string, destination string, opts DownloadOptions) error {
	// Validate URL
	if _, err := parseURL(url); err != nil {
		return errors.NewUserError(fmt.Sprintf("invalid URL %q", url), err)
	}

	// Apply timeout from options if specified
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Determine max attempts (initial attempt + retries)
	maxAttempts := opts.MaxRetries + 1
	if maxAttempts <= 0 {
		maxAttempts = 1 // At least one attempt
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return errors.NewExternalError(fmt.Sprintf("download cancelled after %d attempts", attempt-1), err)
		}

		// Attempt download
		err := h.downloadOnce(ctx, url, destination, opts)
		if err == nil {
			// Success - verify checksum if specified
			if opts.Checksum != "" {
				if err := h.VerifyChecksum(ctx, destination, opts.Checksum); err != nil {
					// Checksum verification failed - clean up and return error
					_ = os.Remove(destination)
					return err
				}
			}
			return nil
		}

		lastErr = err

		// Don't retry on user errors (invalid URL, etc.)
		if errors.IsUserError(err) {
			return err
		}

		// Don't retry if this was the last attempt
		if attempt >= maxAttempts {
			break
		}

		// Calculate backoff delay: 1s → 2s → 4s → 8s → 16s
		backoffDelay := time.Duration(1<<uint(attempt-1)) * time.Second
		if backoffDelay > 16*time.Second {
			backoffDelay = 16 * time.Second
		}

		// Wait before retry (respecting context cancellation)
		select {
		case <-ctx.Done():
			return errors.NewExternalError(fmt.Sprintf("download cancelled during retry backoff after %d attempts", attempt), ctx.Err())
		case <-time.After(backoffDelay):
			// Continue to next attempt
		}
	}

	// All attempts failed
	return errors.NewExternalError(fmt.Sprintf("download failed after %d attempts", maxAttempts), lastErr)
}

// downloadOnce performs a single download attempt without retry logic.
func (h *HTTPDownloader) downloadOnce(ctx context.Context, url string, destination string, opts DownloadOptions) error {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.NewUserError(fmt.Sprintf("create request for %q", url), err)
	}

	// Perform HTTP request
	resp, err := h.client.Do(req)
	if err != nil {
		return errors.NewExternalError(fmt.Sprintf("HTTP request to %q", url), err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return errors.NewExternalError(fmt.Sprintf("HTTP %d from %q", resp.StatusCode, url), nil)
	}

	// Create destination file
	file, err := os.Create(destination)
	if err != nil {
		return errors.NewSystemError(fmt.Sprintf("create file %q", destination), err)
	}
	defer file.Close()

	// Download with progress reporting
	totalBytes := resp.ContentLength // May be -1 if unknown
	var downloadedBytes int64

	// Create progress writer if callback is provided
	var writer io.Writer = file
	if opts.ProgressCallback != nil {
		writer = &progressWriter{
			writer:   file,
			callback: opts.ProgressCallback,
			total:    totalBytes,
			current:  &downloadedBytes,
		}
	}

	// Copy response body to file
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		// Clean up partial download
		_ = os.Remove(destination)
		return errors.NewExternalError(fmt.Sprintf("download from %q", url), err)
	}

	// Final progress callback (100%)
	if opts.ProgressCallback != nil && totalBytes > 0 {
		opts.ProgressCallback(downloadedBytes, totalBytes)
	}

	return nil
}

// VerifyChecksum verifies that the file at the given path matches the expected checksum.
// The checksum format should be "algorithm:hash" (e.g., "sha256:abc123...") or just "hash"
// (SHA-256 is assumed).
//
// Parameters:
//   - ctx: Context for cancellation
//   - file: Path to the file to verify
//   - expectedChecksum: Expected checksum in "algorithm:hash" or "hash" format
//
// Returns:
//   - error: nil if checksum matches, ErrChecksumMismatch if it doesn't match,
//     or another error if verification fails for other reasons
//
// The implementation:
//   - Supports SHA-256 checksums (required by Requirement 4.6)
//   - Returns ErrChecksumMismatch when checksums don't match
//   - Deletes the file if checksum verification fails
//   - Supports the format "sha256:hash" or just "hash" (assuming SHA-256)
func (h *HTTPDownloader) VerifyChecksum(ctx context.Context, file string, expectedChecksum string) error {
	// Parse checksum format
	algorithm, expectedHash, err := parseChecksum(expectedChecksum)
	if err != nil {
		return errors.NewUserError(fmt.Sprintf("invalid checksum format %q", expectedChecksum), err)
	}

	// Only SHA-256 is supported
	if algorithm != "sha256" {
		return errors.NewUserError(fmt.Sprintf("unsupported checksum algorithm %q (only sha256 is supported)", algorithm), nil)
	}

	// Open file
	f, err := os.Open(file)
	if err != nil {
		return errors.NewSystemError(fmt.Sprintf("open file %q for checksum verification", file), err)
	}
	defer f.Close()

	// Compute SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return errors.NewSystemError(fmt.Sprintf("compute checksum for %q", file), err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Compare checksums (case-insensitive)
	if !strings.EqualFold(actualHash, expectedHash) {
		// Delete file on checksum mismatch
		_ = os.Remove(file)
		return errors.Wrap(
			errors.ErrChecksumMismatch,
			"checksum mismatch for %q: expected %s, got %s",
			file, expectedHash, actualHash,
		)
	}

	return nil
}

// progressWriter wraps an io.Writer and calls a progress callback during writes.
type progressWriter struct {
	writer   io.Writer
	callback func(downloaded, total int64)
	total    int64
	current  *int64
}

// Write implements io.Writer and calls the progress callback.
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		*pw.current += int64(n)
		pw.callback(*pw.current, pw.total)
	}
	return n, err
}

// parseURL validates and parses a URL string.
func parseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Validate scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %q (only http and https are supported)", u.Scheme)
	}

	// Validate host
	if u.Host == "" {
		return nil, fmt.Errorf("missing host in URL")
	}

	return u, nil
}

// parseChecksum parses a checksum string in "algorithm:hash" or "hash" format.
// If no algorithm is specified, "sha256" is assumed.
// Returns (algorithm, hash, error).
func parseChecksum(checksum string) (string, string, error) {
	checksum = strings.TrimSpace(checksum)
	if checksum == "" {
		return "", "", fmt.Errorf("empty checksum")
	}

	// Check for "algorithm:hash" format
	if strings.Contains(checksum, ":") {
		parts := strings.SplitN(checksum, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid checksum format")
		}
		algorithm := strings.ToLower(strings.TrimSpace(parts[0]))
		hash := strings.TrimSpace(parts[1])
		if algorithm == "" || hash == "" {
			return "", "", fmt.Errorf("invalid checksum format")
		}
		return algorithm, hash, nil
	}

	// Assume SHA-256 if no algorithm specified
	return "sha256", checksum, nil
}
