// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHTTPDownloader_DownloadErrorsExt(t *testing.T) {
	dl := NewHTTPDownloader()

	// Empty URL
	err := dl.Download(context.Background(), "", "test", DownloadOptions{})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}

	// Max retries
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	dest := filepath.Join(t.TempDir(), "fail.txt")
	err = dl.Download(context.Background(), ts.URL, dest, DownloadOptions{MaxRetries: 1})
	if err == nil {
		t.Fatal("expected error after max retries")
	}
}

func TestHTTPDownloader_DownloadConcurrent_Errors(t *testing.T) {
	dl := NewHTTPDownloader()
	dir := t.TempDir()

	// 1. Server ignores range
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("full content"))
	}))
	defer ts1.Close()

	dest1 := filepath.Join(dir, "no_range.txt")
	err := dl.downloadConcurrent(context.Background(), ts1.URL, dest1, 10485760, DownloadOptions{})
	if err == nil || err.Error() != "server ignored Range header, concurrent download impossible" {
		t.Fatalf("expected server ignored range error, got %v", err)
	}

	// 2. Mock panic in concurrent thread
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", "10485760")
		w.WriteHeader(http.StatusPartialContent)
		// Return partial then let thread die
		w.Write([]byte("ok"))
	}))
	defer ts2.Close()

	dest2 := filepath.Join(dir, "panic.txt")

	// Temporarily limit retries to fail fast
	os.Setenv("JOBS", "1")
	defer os.Setenv("JOBS", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = dl.downloadConcurrent(ctx, ts2.URL, dest2, 10485760, DownloadOptions{})
	if err == nil {
		t.Fatal("expected error in concurrent due to short body or failure")
	}
}

func TestHTTPDownloader_VerifyGPGSignature_Skipped(t *testing.T) {
	dl := NewHTTPDownloader()
	dir := t.TempDir()

	// Mock 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	dest := filepath.Join(dir, "test.txt")

	opts := DownloadOptions{
		VerifyGPG: true,
		GPGResult: &GPGResult{},
	}
	_ = dl.Download(context.Background(), ts.URL, dest, opts)
}

func TestHTTPDownloader_VerifyGPGSignature_500(t *testing.T) {
	dl := NewHTTPDownloader()
	dir := t.TempDir()

	// Mock 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	dest := filepath.Join(dir, "test2.txt")
	os.WriteFile(dest, []byte("data"), 0644)

	// Mock valid keyring but fetch fails
	os.Setenv("UNIRTM_DATA_DIR", dir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	// An empty file will trigger "failed to parse keyring", but we want it to parse
	// Actually we just test what happens if keyring fails to parse, wait we already did that.
	// We want to hit the 500 error branch:
	_ = dl.verifyGPGSignature(context.Background(), ts.URL, dest)
}
