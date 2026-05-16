// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/errors"
)

// ErrGPGSkipped is returned when a signature file is not found (404) and verification is skipped.
var ErrGPGSkipped = errors.NewUserError("GPG signature not found, skipped", nil)

type contextKey string

const githubProxyKey contextKey = "github_proxy"

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
	h := &HTTPDownloader{}
	h.client = &http.Client{
		Timeout: 30 * time.Minute, // Increased total timeout for large files
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				// Check UNIRTM_ or MISE_ prefixed env vars first via env.Get
				if req.URL.Scheme == "http" {
					if v := env.Get("HTTP_PROXY"); v != "" {
						return url.Parse(v)
					}
				} else if req.URL.Scheme == "https" {
					if v := env.Get("HTTPS_PROXY"); v != "" {
						return url.Parse(v)
					}
				}
				if v := env.Get("ALL_PROXY"); v != "" {
					return url.Parse(v)
				}

				// Fallback to standard system environment
				return http.ProxyFromEnvironment(req)
			},
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
		},
	}

	// Support disabling HTTP/2 via environment variable for compatibility with some proxies/mirrors (like Aliyun)
	// that might send malformed HTTP/2 frames or have ALPN issues.
	if env.Get("HTTP2") == "0" {
		// Setting TLSNextProto to an empty, non-nil map disables HTTP/2 support in http.Transport
		h.client.Transport.(*http.Transport).TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	}

	h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}

		// Get proxy from context
		proxy, ok := req.Context().Value(githubProxyKey).(string)
		if ok && proxy != "" {
			nextURL := req.URL.String()
			if (strings.Contains(nextURL, "github.com") || strings.Contains(nextURL, "githubusercontent.com")) && !strings.HasPrefix(nextURL, proxy) {
				// Ensure proxy ends with /
				p := proxy
				if !strings.HasSuffix(p, "/") {
					p += "/"
				}
				// Apply proxy to the redirect target
				newURL, err := url.Parse(p + nextURL)
				if err == nil {
					req.URL = newURL
				}
			}
		}
		return nil
	}
	return h
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
	// Inject proxy into context for CheckRedirect
	if opts.GitHubProxy != "" {
		ctx = context.WithValue(ctx, githubProxyKey, opts.GitHubProxy)
	}

	// Track if we need to force HTTP/1.1 due to protocol errors
	forceHTTP11 := env.Get("HTTP2") == "0"

	// Apply GitHub proxy if configured and URL is from GitHub
	if opts.GitHubProxy != "" && (strings.Contains(url, "github.com") || strings.Contains(url, "githubusercontent.com")) {
		// Ensure proxy ends with /
		proxy := opts.GitHubProxy
		if !strings.HasSuffix(proxy, "/") {
			proxy += "/"
		}
		// Avoid double proxying
		if !strings.HasPrefix(url, proxy) {
			url = proxy + url
		}
	}
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

		// If forced to HTTP/1.1, ensure client is configured properly for this attempt
		var originalClient *http.Client
		if forceHTTP11 {
			if transport, ok := h.client.Transport.(*http.Transport); ok {
				newTransport := transport.Clone()
				// Physically disable HTTP/2
				newTransport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
				
				originalClient = h.client
				h.client = &http.Client{
					Timeout:       originalClient.Timeout,
					CheckRedirect: originalClient.CheckRedirect,
					Transport:     newTransport,
				}
			}
		}

		// Attempt download
		err := h.downloadOnce(ctx, url, destination, opts)

		// Restore client immediately if it was swapped
		if originalClient != nil {
			// Close idle connections of the temporary transport to be clean
			if tempTransport, ok := h.client.Transport.(*http.Transport); ok {
				tempTransport.CloseIdleConnections()
			}
			h.client = originalClient
		}
		if err == nil {
			// Success - verify checksum if specified
			if opts.Checksum != "" {
				if err := h.VerifyChecksum(ctx, destination, opts.Checksum); err != nil {
					// Checksum verification failed - clean up and return error
					_ = os.Remove(destination)
					return err
				}
			}

			// Verify GPG signature if specified
			if opts.VerifyGPG {
				err := h.verifyGPGSignature(ctx, url, destination)
				if err != nil {
					if err == ErrGPGSkipped {
						if opts.GPGResult != nil {
							opts.GPGResult.Status = "Skipped"
						}
					} else {
						if opts.GPGResult != nil {
							opts.GPGResult.Status = "Failed"
						}
						_ = os.Remove(destination)
						return err
					}
				} else {
					if opts.GPGResult != nil {
						opts.GPGResult.Status = "Success"
					}
				}
			}
			return nil
		}

		lastErr = err

		// Detect protocol errors and trigger fallback to HTTP/1.1
		if strings.Contains(err.Error(), "malformed HTTP response") {
			if !forceHTTP11 {
				forceHTTP11 = true
				// We don't increment attempt counter yet, just retry immediately with HTTP/1.1
				// if this is the first time we see this error.
				attempt--
				continue
			}
		}

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
	// 1. Pre-flight check: Get content length and range support using HEAD
	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err == nil {
		headResp, err := h.client.Do(headReq)
		if err == nil {
			defer headResp.Body.Close()
			if headResp.StatusCode == http.StatusOK {
				totalBytes := headResp.ContentLength
				acceptRanges := headResp.Header.Get("Accept-Ranges") == "bytes"

				// 2. Decide if we use concurrent download
				// Criteria: Size > 1MB, Server supports Ranges
				if acceptRanges && totalBytes > 1*1024*1024 {
					err := h.downloadConcurrent(ctx, url, destination, totalBytes, opts)
					if err == nil {
						return nil
					}
					// Fallback on failure
				}
			}
		}
	}

	// 3. Standard download logic...
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.NewUserError(fmt.Sprintf("create request for %q", url), err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return errors.NewExternalError(fmt.Sprintf("HTTP request to %q", url), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewExternalError(fmt.Sprintf("HTTP %d from %q", resp.StatusCode, url), nil)
	}
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return errors.NewSystemError(fmt.Sprintf("create directory %q", filepath.Dir(destination)), err)
	}

	// Create destination file with restricted permissions (0600) to prevent poisoning
	file, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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
	n, err := io.Copy(writer, resp.Body)
	if err != nil {
		// Clean up partial download
		_ = os.Remove(destination)
		return errors.NewExternalError(fmt.Sprintf("download from %q", url), err)
	}

	// Verify that we downloaded the expected amount of data
	if totalBytes > 0 && n != totalBytes {
		_ = os.Remove(destination)
		return errors.NewExternalError(fmt.Sprintf("download from %q", url), fmt.Errorf("size mismatch: expected %d bytes, got %d", totalBytes, n))
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

// verifyGPGSignature downloads a detached signature (.sig or .asc) and verifies it against the local keyring.
func (h *HTTPDownloader) verifyGPGSignature(ctx context.Context, targetURL, destination string) error {
	keyringPath := filepath.Join(env.GetDataDir(), "keyring.gpg")
	keyringFile, err := os.Open(keyringPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewUserError(fmt.Sprintf("keyring not found at %s. Please add trusted keys first", keyringPath), nil)
		}
		return errors.NewSystemError("failed to open keyring", err)
	}
	defer keyringFile.Close()

	keyring, err := openpgp.ReadKeyRing(keyringFile)
	if err != nil {
		return errors.NewSystemError("failed to parse keyring", err)
	}

	// Try .sig first, then .asc
	sigURL := targetURL + ".sig"
	sigDest := destination + ".sig"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sigURL, nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		// Fallback to .asc
		sigURL = targetURL + ".asc"
		sigDest = destination + ".asc"
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, sigURL, nil)
		resp, err = h.client.Do(req)
		if err != nil {
			return errors.NewExternalError("failed to fetch signature", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			// If neither .sig nor .asc exist, just skip verification (assume unsigned)
			if resp.StatusCode == http.StatusNotFound {
				return ErrGPGSkipped
			}
			return errors.NewExternalError("signature not found at .sig or .asc", nil)
		}
	}

	sigFile, err := os.Create(sigDest)
	if err != nil {
		resp.Body.Close()
		return err
	}
	defer os.Remove(sigDest)

	_, err = io.Copy(sigFile, resp.Body)
	resp.Body.Close()
	sigFile.Close() // Close before reading for verification

	if err != nil {
		return errors.NewSystemError("failed to save signature file", err)
	}

	// Perform verification
	targetFile, err := os.Open(destination)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	sigFileRead, err := os.Open(sigDest)
	if err != nil {
		return err
	}
	defer sigFileRead.Close()

	_, err = openpgp.CheckDetachedSignature(keyring, targetFile, sigFileRead, nil)
	if err != nil {
		return errors.Wrap(errors.ErrChecksumMismatch, "GPG signature verification failed: %v", err)
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
	if u.Scheme == "http" {
		fmt.Printf("⚠️  WARNING: Using insecure HTTP for download: %s. This is vulnerable to man-in-the-middle attacks.\n", rawURL)
	} else if u.Scheme != "https" {
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

// downloadConcurrent performs a multi-threaded download using Range requests.
func (h *HTTPDownloader) downloadConcurrent(ctx context.Context, url string, destination string, totalSize int64, opts DownloadOptions) error {
	// 1. Determine optimal number of threads
	numThreads := 4
	if totalSize > 200*1024*1024 {
		numThreads = 16 // Huge files (> 200MB)
	} else if totalSize > 50*1024*1024 {
		numThreads = 12 // Large files (50-200MB)
	} else if totalSize > 5*1024*1024 {
		numThreads = 8 // Medium files (5-50MB)
	}

	// 2. Allow explicit override via JOBS environment variable
	if jobs := env.Get("JOBS"); jobs != "" {
		var j int
		if _, err := fmt.Sscanf(jobs, "%d", &j); err == nil && j > 0 {
			numThreads = j
		}
	}

	// Clamp threads to avoid overwhelming resources
	if numThreads > 32 {
		numThreads = 32 // Hard cap at 32 threads
	}

	// Pre-allocate file
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Truncate(totalSize); err != nil {
		return err
	}

	chunkSize := totalSize / int64(numThreads)
	var wg sync.WaitGroup
	var downloadErr error
	var errOnce sync.Once
	
	downloadedBytes := int64(0)
	
	for i := 0; i < numThreads; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numThreads-1 {
			end = totalSize - 1
		}

		wg.Add(1)
		go func(start, end int64, threadID int) {
			defer wg.Done()
			
			backoff := 1 * time.Second
			for attempt := 0; attempt < 3; attempt++ {
				if ctx.Err() != nil {
					return
				}
				
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					continue
				}
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
				
				resp, err := h.client.Do(req)
				if err != nil {
					time.Sleep(backoff)
					backoff *= 2
					continue
				}
				
				if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					time.Sleep(backoff)
					backoff *= 2
					continue
				}
				
				// Use a local buffer to read and then WriteAt
				buf := make([]byte, 32*1024)
				currentOffset := start
				for {
					n, readErr := resp.Body.Read(buf)
					if n > 0 {
						_, writeErr := file.WriteAt(buf[:n], currentOffset)
						if writeErr != nil {
							errOnce.Do(func() { downloadErr = writeErr })
							resp.Body.Close()
							return
						}
						currentOffset += int64(n)
						atomic.AddInt64(&downloadedBytes, int64(n))
						if opts.ProgressCallback != nil {
							opts.ProgressCallback(atomic.LoadInt64(&downloadedBytes), totalSize)
						}
					}
					if readErr != nil {
						resp.Body.Close()
						if readErr == io.EOF {
							return // Success for this chunk
						}
						break // Retry
					}
				}
				time.Sleep(backoff)
				backoff *= 2
			}
			errOnce.Do(func() { downloadErr = fmt.Errorf("thread %d failed after retries", threadID) })
		}(start, end, i)
	}

	wg.Wait()
	return downloadErr
}
