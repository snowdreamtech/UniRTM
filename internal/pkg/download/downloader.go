package download

import (
	"context"
	"time"
)

// Downloader defines the interface for downloading files from remote sources.
// Implementations must support retry logic, timeout configuration, checksum verification,
// and progress reporting.
//
// The interface is designed to be generic enough to support multiple implementations
// (HTTP, custom protocols, etc.) and integrates with the error handling system.
//
// Example usage:
//
//	opts := DownloadOptions{
//	    Checksum:   "sha256:abc123...",
//	    MaxRetries: 5,
//	    Timeout:    60 * time.Second,
//	    ProgressCallback: func(downloaded, total int64) {
//	        fmt.Printf("Progress: %d/%d bytes\n", downloaded, total)
//	    },
//	}
//	err := downloader.Download(ctx, "https://example.com/file.tar.gz", "/tmp/file.tar.gz", opts)
type Downloader interface {
	// Download downloads a file from the specified URL to the destination path.
	// The operation respects the context for cancellation and deadlines.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - url: The source URL to download from
	//   - destination: The local file path where the downloaded file will be saved
	//   - opts: Download options including retry, timeout, and progress callback
	//
	// Returns:
	//   - error: nil on success, or an error describing the failure
	//
	// The implementation MUST:
	//   - Respect context cancellation and deadlines
	//   - Implement retry logic with exponential backoff as specified in opts
	//   - Call the progress callback (if provided) during download
	//   - Verify the checksum after download if specified in opts
	//   - Clean up partial downloads on failure
	//   - Return descriptive errors with context (URL, attempt count, failure reason)
	Download(ctx context.Context, url string, destination string, opts DownloadOptions) error

	// VerifyChecksum verifies that the file at the given path matches the expected checksum.
	// The checksum format should be "algorithm:hash" (e.g., "sha256:abc123...").
	//
	// Parameters:
	//   - ctx: Context for cancellation
	//   - file: Path to the file to verify
	//   - expectedChecksum: Expected checksum in "algorithm:hash" format
	//
	// Returns:
	//   - error: nil if checksum matches, ErrChecksumMismatch if it doesn't match,
	//            or another error if verification fails for other reasons
	//
	// The implementation MUST:
	//   - Support SHA-256 checksums (required by Requirement 4.6)
	//   - Return ErrChecksumMismatch from internal/pkg/errors when checksums don't match
	//   - Delete the file if checksum verification fails
	//   - Support the format "sha256:hash" or just "hash" (assuming SHA-256)
	VerifyChecksum(ctx context.Context, file string, expectedChecksum string) error
}

// DownloadOptions configures the behavior of a download operation.
// All fields are optional and have sensible defaults when not specified.
type DownloadOptions struct {
	// Checksum is the expected SHA-256 checksum of the downloaded file.
	// Format: "sha256:hash" or just "hash" (SHA-256 is assumed).
	// If specified, the downloader will verify the checksum after download
	// and fail if it doesn't match.
	// If empty, checksum verification is skipped.
	Checksum string

	// MaxRetries is the maximum number of retry attempts for failed downloads.
	// The implementation should use exponential backoff: 1s → 2s → 4s → 8s → 16s.
	// Default: 5 attempts (as specified in Requirement 4.3)
	// Set to 0 to disable retries (fail immediately on first error).
	MaxRetries int

	// Timeout is the total operation timeout for the download.
	// This includes all retry attempts.
	// Default: implementation-specific (recommended: 5 minutes for large files)
	// The context passed to Download() can also enforce timeouts.
	Timeout time.Duration

	// ProgressCallback is called periodically during download to report progress.
	// Parameters:
	//   - bytesDownloaded: Number of bytes downloaded so far
	//   - totalBytes: Total size of the file (may be 0 if unknown)
	//
	// The callback is optional. If nil, progress is not reported.
	// The callback should be fast and non-blocking.
	// The callback may be called from a different goroutine.
	ProgressCallback func(bytesDownloaded, totalBytes int64)
}

// DefaultDownloadOptions returns DownloadOptions with sensible defaults.
// This is a convenience function for common use cases.
//
// Defaults:
//   - MaxRetries: 5 (as per Requirement 4.3)
//   - Timeout: 5 minutes
//   - No checksum verification
//   - No progress callback
func DefaultDownloadOptions() DownloadOptions {
	return DownloadOptions{
		MaxRetries: 5,
		Timeout:    5 * time.Minute,
	}
}

// WithChecksum returns a copy of the options with the checksum set.
// This is a convenience method for fluent configuration.
func (o DownloadOptions) WithChecksum(checksum string) DownloadOptions {
	o.Checksum = checksum
	return o
}

// WithMaxRetries returns a copy of the options with MaxRetries set.
// This is a convenience method for fluent configuration.
func (o DownloadOptions) WithMaxRetries(retries int) DownloadOptions {
	o.MaxRetries = retries
	return o
}

// WithTimeout returns a copy of the options with Timeout set.
// This is a convenience method for fluent configuration.
func (o DownloadOptions) WithTimeout(timeout time.Duration) DownloadOptions {
	o.Timeout = timeout
	return o
}

// WithProgressCallback returns a copy of the options with ProgressCallback set.
// This is a convenience method for fluent configuration.
func (o DownloadOptions) WithProgressCallback(callback func(int64, int64)) DownloadOptions {
	o.ProgressCallback = callback
	return o
}
