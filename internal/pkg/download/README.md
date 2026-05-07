# Download Module

## Overview

The download module provides a generic interface for downloading files from remote sources with support for retry logic, timeout configuration, checksum verification, and progress reporting.

## Architecture

The module follows the **Dependency Inversion Principle** — the core business logic depends on the `Downloader` interface abstraction, not on concrete implementations. This allows for:

- **Pluggable implementations**: HTTP, S3, custom protocols
- **Easy testing**: Mock implementations for unit tests
- **Flexibility**: Different download strategies for different backends

## Components

### Downloader Interface

The `Downloader` interface defines two methods:

1. **`Download(ctx, url, destination, opts)`**: Downloads a file from a URL to a local path
2. **`VerifyChecksum(ctx, file, expectedChecksum)`**: Verifies the SHA-256 checksum of a file

### DownloadOptions Struct

Configures download behavior:

- **`Checksum`**: Expected SHA-256 checksum (format: "sha256:hash" or just "hash")
- **`MaxRetries`**: Maximum retry attempts (default: 5)
- **`Timeout`**: Total operation timeout (default: 5 minutes)
- **`ProgressCallback`**: Optional callback for progress reporting

## Usage

### Basic Download

```go
package main

import (
    "context"
    "fmt"
    "github.com/snowdreamtech/unirtm/internal/pkg/download"
)

func main() {
    // Create a downloader implementation (e.g., HTTPDownloader)
    downloader := NewHTTPDownloader()

    // Use default options
    opts := download.DefaultDownloadOptions()

    // Download a file
    err := downloader.Download(
        context.Background(),
        "https://example.com/tool-v1.0.0.tar.gz",
        "/tmp/tool-v1.0.0.tar.gz",
        opts,
    )
    if err != nil {
        fmt.Printf("Download failed: %v\n", err)
        return
    }

    fmt.Println("Download successful!")
}
```

### Download with Checksum Verification

```go
opts := download.DefaultDownloadOptions().
    WithChecksum("sha256:abc123def456...").
    WithMaxRetries(3)

err := downloader.Download(ctx, url, destination, opts)
if err != nil {
    // Handle error (checksum mismatch, network failure, etc.)
}
```

### Download with Progress Reporting

```go
opts := download.DefaultDownloadOptions().
    WithProgressCallback(func(downloaded, total int64) {
        if total > 0 {
            percent := float64(downloaded) / float64(total) * 100
            fmt.Printf("\rProgress: %.2f%% (%d/%d bytes)", percent, downloaded, total)
        } else {
            fmt.Printf("\rDownloaded: %d bytes", downloaded)
        }
    })

err := downloader.Download(ctx, url, destination, opts)
```

### Standalone Checksum Verification

```go
err := downloader.VerifyChecksum(
    ctx,
    "/tmp/tool-v1.0.0.tar.gz",
    "sha256:abc123def456...",
)
if errors.Is(err, pkgerrors.ErrChecksumMismatch) {
    fmt.Println("Checksum verification failed!")
}
```

## Implementation Requirements

Any implementation of the `Downloader` interface MUST:

1. **Respect context cancellation and deadlines**
2. **Implement retry logic with exponential backoff** (1s → 2s → 4s → 8s → 16s)
3. **Support connection timeouts** (recommended: 10 seconds)
4. **Support read timeouts** (recommended: 60 seconds)
5. **Support proxy configuration** via `HTTP_PROXY` and `HTTPS_PROXY` environment variables
6. **Verify SHA-256 checksums** after download completion
7. **Clean up partial downloads** on failure
8. **Return descriptive errors** with context (URL, attempt count, failure reason)
9. **Call progress callback** (if provided) during download
10. **Delete files** that fail checksum verification

## Error Handling

The download module integrates with the project's error handling system (`internal/pkg/errors`):

- **User Errors** (CategoryUser): Invalid URL, invalid checksum format
- **System Errors** (CategorySystem): Disk full, permission denied
- **External Errors** (CategoryExternal): Network failures, HTTP errors, timeout

### Common Errors

- `ErrChecksumMismatch`: Checksum verification failed
- `ErrNetworkFailure`: Network operation failed after retries
- Context errors: `context.Canceled`, `context.DeadlineExceeded`

## Testing

### Unit Tests

Test the interface contract with mock implementations:

```go
type MockDownloader struct {
    DownloadFunc        func(context.Context, string, string, DownloadOptions) error
    VerifyChecksumFunc  func(context.Context, string, string) error
}

func (m *MockDownloader) Download(ctx context.Context, url, dest string, opts DownloadOptions) error {
    return m.DownloadFunc(ctx, url, dest, opts)
}

func (m *MockDownloader) VerifyChecksum(ctx context.Context, file, checksum string) error {
    return m.VerifyChecksumFunc(ctx, file, checksum)
}
```

### Integration Tests

Test concrete implementations with real network operations:

- Test successful downloads
- Test retry behavior with transient failures
- Test checksum verification (both success and failure)
- Test timeout behavior
- Test context cancellation
- Test progress callback invocation

## Design Decisions

### Why an Interface?

The `Downloader` interface allows for:

1. **Multiple implementations**: HTTP, S3, FTP, custom protocols
2. **Easy mocking**: Unit tests don't require network access
3. **Flexibility**: Different backends can use different download strategies
4. **Testability**: Property-based tests can verify interface contracts

### Why Exponential Backoff?

Exponential backoff (1s → 2s → 4s → 8s → 16s) is a best practice for retry logic:

- Reduces load on failing servers
- Gives transient issues time to resolve
- Prevents thundering herd problems
- Specified in Requirement 4.3

### Why SHA-256?

SHA-256 is the industry standard for file integrity verification:

- Cryptographically secure
- Widely supported
- Fast enough for large files
- Specified in Requirement 4.6

## Future Enhancements

Potential future improvements:

1. **Resume support**: Resume interrupted downloads using HTTP Range requests
2. **Parallel chunk downloads**: Download large files in parallel chunks
3. **Compression**: Transparent decompression of gzip/bzip2 files
4. **Signature verification**: GPG signature verification for tools that provide signatures
5. **Mirror support**: Automatic failover to mirror URLs
6. **Bandwidth limiting**: Rate limiting for downloads

## Related Components

- **Backend System** (`internal/backend/`): Uses the downloader to fetch tool artifacts
- **Error Handling** (`internal/pkg/errors/`): Provides error classification
- **Logger** (`internal/pkg/logger/`): Logs download operations and errors
- **Cache Manager** (future): Caches downloaded artifacts

## References

- Design Document: `.kiro/specs/unirtm/design.md` (Download System section)
- Requirements: `.kiro/specs/unirtm/requirements.md` (Requirement 4: Generic Download Interface)
- Error Handling: `internal/pkg/errors/README.md`
