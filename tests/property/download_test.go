package property

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	pkgerrors "github.com/snowdreamtech/unirtm/internal/pkg/errors"
)

// TestProperty13_DownloadRetryBehavior verifies that downloads retry with exponential backoff
// Property 13: Download Retry Behavior
// Validates: Requirements 4.3
func TestProperty13_DownloadRetryBehavior(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("downloads retry with exponential backoff on transient failures", prop.ForAll(
		func(maxRetries int) bool {
			if maxRetries < 0 || maxRetries > 10 {
				return true // Skip invalid values
			}

			// Track attempts and timing
			attempts := 0
			attemptTimes := []time.Time{}

			// Create test server that fails first N attempts
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				attemptTimes = append(attemptTimes, time.Now())

				if attempts <= maxRetries {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			}))
			defer server.Close()

			// Create temporary destination
			tmpDir, err := os.MkdirTemp("", "download-test-*")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)
			destination := filepath.Join(tmpDir, "test.txt")

			// Download with retries
			downloader := download.NewHTTPDownloader()
			opts := download.DefaultDownloadOptions().WithMaxRetries(maxRetries + 1)
			err = downloader.Download(context.Background(), server.URL, destination, opts)

			// Should succeed after retries
			if err != nil {
				return false
			}

			// Verify correct number of attempts
			if attempts != maxRetries+2 { // Initial + retries + final success
				return false
			}

			// Verify exponential backoff (allow some timing variance)
			if len(attemptTimes) >= 3 {
				delay1 := attemptTimes[1].Sub(attemptTimes[0])
				delay2 := attemptTimes[2].Sub(attemptTimes[1])
				// Second delay should be roughly 2x first delay (with tolerance)
				if delay2 < delay1 || delay2 > delay1*3 {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 3), // Test with 0-3 retries for speed
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty14_ChecksumVerification verifies checksum verification correctness
// Property 14: Checksum Verification
// Validates: Requirements 4.6
func TestProperty14_ChecksumVerification(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("checksum verification detects mismatches", prop.ForAll(
		func(content string) bool {
			if len(content) == 0 {
				return true // Skip empty content
			}

			// Compute correct checksum
			hasher := sha256.New()
			hasher.Write([]byte(content))
			correctChecksum := hex.EncodeToString(hasher.Sum(nil))

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(content))
			}))
			defer server.Close()

			// Create temporary destination
			tmpDir, err := os.MkdirTemp("", "download-test-*")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)

			// Test 1: Correct checksum should succeed
			destination1 := filepath.Join(tmpDir, "test1.txt")
			downloader := download.NewHTTPDownloader()
			opts1 := download.DefaultDownloadOptions().WithChecksum("sha256:" + correctChecksum)
			err = downloader.Download(context.Background(), server.URL, destination1, opts1)
			if err != nil {
				return false
			}

			// Verify file exists
			if _, err := os.Stat(destination1); os.IsNotExist(err) {
				return false
			}

			// Test 2: Wrong checksum should fail
			destination2 := filepath.Join(tmpDir, "test2.txt")
			wrongChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			opts2 := download.DefaultDownloadOptions().WithChecksum(wrongChecksum)
			err = downloader.Download(context.Background(), server.URL, destination2, opts2)

			// Should return checksum mismatch error
			if !errors.Is(err, pkgerrors.ErrChecksumMismatch) {
				return false
			}

			// File should be deleted after mismatch
			if _, err := os.Stat(destination2); !os.IsNotExist(err) {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty15_DownloadErrorReporting verifies error reporting completeness
// Property 15: Download Error Reporting
// Validates: Requirements 4.7
func TestProperty15_DownloadErrorReporting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("download errors include URL and context", prop.ForAll(
		func(statusCode int) bool {
			if statusCode < 400 || statusCode > 599 {
				return true // Skip non-error status codes
			}

			// Create test server that returns error
			testURL := ""
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			}))
			defer server.Close()
			testURL = server.URL

			// Create temporary destination
			tmpDir, err := os.MkdirTemp("", "download-test-*")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)
			destination := filepath.Join(tmpDir, "test.txt")

			// Attempt download
			downloader := download.NewHTTPDownloader()
			opts := download.DefaultDownloadOptions().WithMaxRetries(0)
			err = downloader.Download(context.Background(), testURL, destination, opts)

			// Should return error
			if err == nil {
				return false
			}

			// Error should be categorized as external
			if !pkgerrors.IsExternalError(err) {
				return false
			}

			// Error message should contain status code
			errMsg := err.Error()
			if !contains(errMsg, fmt.Sprintf("%d", statusCode)) {
				return false
			}

			return true
		},
		gen.OneConstOf(400, 403, 404, 500, 502, 503),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
