// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPDownloader_CheckRedirect(t *testing.T) {
	d := NewHTTPDownloader()

	// Too many redirects
	via := make([]*http.Request, 10)
	err := d.client.CheckRedirect(&http.Request{}, via)
	if err == nil || err.Error() != "too many redirects" {
		t.Errorf("expected too many redirects error")
	}

	// Normal redirect without proxy
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://github.com/foo", nil)
	err = d.client.CheckRedirect(req, []*http.Request{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test redirect through HTTPDownloader with mock server
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "https://github.com/foo", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectServer.Close()

	opts := DefaultDownloadOptions()
	opts.MaxRetries = 1
	opts.GitHubProxy = "https://proxy.example.com/"
	dest := filepath.Join(t.TempDir(), "proxy.txt")
	// This will fail because https://proxy.example.com/https://github.com/foo doesn't exist, but it will trigger the redirect logic
	_ = d.Download(context.Background(), redirectServer.URL+"/redirect", dest, opts)
}

func TestHTTPDownloader_JitterBackoff(t *testing.T) {
	// Test backoff calculation
	delay1 := jitterBackoff(1 * time.Second)
	if delay1 < time.Second || delay1 > 2*time.Second {
		t.Errorf("expected delay1 to be between 1s and 2s, got %v", delay1)
	}

	delay3 := jitterBackoff(4 * time.Second)
	if delay3 < 4*time.Second || delay3 > 8*time.Second {
		t.Errorf("expected delay3 to be between 4s and 8s, got %v", delay3)
	}
}

func TestHTTPDownloader_DownloadConcurrent(t *testing.T) {
	// Create a large file test server to trigger concurrent download
	content := strings.Repeat("A", 1024*1024*6) // 6MB, more than concurrentChunkSize (5MB)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer ts.Close()

	d := NewHTTPDownloader()
	dir := t.TempDir()
	dest := filepath.Join(dir, "large_file.txt")

	opts := DefaultDownloadOptions()
	err := d.Download(context.Background(), ts.URL, dest, opts)
	if err != nil {
		t.Fatalf("unexpected error downloading large file: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("unexpected error reading downloaded file: %v", err)
	}

	if len(data) != len(content) {
		t.Errorf("expected %d bytes, got %d", len(content), len(data))
	}
	if string(data) != content {
		t.Errorf("content mismatch")
	}
}

func TestHTTPDownloader_VerifyGPGSignature(t *testing.T) {
	d := NewHTTPDownloader()
	ctx := context.Background()
	dir := t.TempDir()
	dest := filepath.Join(dir, "test.txt")
	os.WriteFile(dest, []byte("test data"), 0644)

	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// 1. Keyring not found
	// override UNIRTM_DATA_DIR to a fake directory
	os.Setenv("UNIRTM_DATA_DIR", filepath.Join(dir, "nonexistent"))
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	err := d.verifyGPGSignature(ctx, ts.URL, dest)
	if err == nil || !strings.Contains(err.Error(), "keyring not found") {
		t.Errorf("expected keyring not found error, got %v", err)
	}

	// 2. Keyring exists but invalid format
	os.Setenv("UNIRTM_DATA_DIR", dir)
	keyringPath := filepath.Join(dir, "keyring.gpg")
	os.WriteFile(keyringPath, []byte("invalid keyring data"), 0644)

	err = d.verifyGPGSignature(ctx, ts.URL, dest)
	if err == nil || !strings.Contains(err.Error(), "failed to parse keyring") {
		t.Errorf("expected parse keyring error, got %v", err)
	}
}
