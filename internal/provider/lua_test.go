// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestResolveLuaBinariesURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goarch  string
		want    string
		wantErr bool
	}{
		{
			name:    "Windows x64",
			version: "5.4.2",
			goarch:  "amd64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Win64_bin.zip/download",
			wantErr: false,
		},
		{
			name:    "Windows x86",
			version: "5.4.2",
			goarch:  "386",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Win32_bin.zip/download",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLuaBinariesURL(tt.version, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLuaBinariesURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveLuaBinariesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadLuaArchive(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		time.Sleep(20 * time.Millisecond) // Ensure concurrent overlap
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-lua-archive-content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("UNIRTM_DOWNLOADS_DIR", tmpDir)
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	ctx := context.Background()
	filename := fmt.Sprintf("test-lua-%d.tar.gz", time.Now().UnixNano())

	// 1. Run 5 concurrent downloads to test singleflight deduplication
	var wg sync.WaitGroup
	errCh := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, err := downloadLuaArchive(ctx, server.URL, filename)
			if err != nil {
				errCh <- err
				return
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "fake-lua-archive-content" {
				errCh <- fmt.Errorf("unexpected file content: %s", string(data))
				return
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent download error: %v", err)
	}

	if count := atomic.LoadInt32(&requestCount); count != 1 {
		t.Errorf("expected 1 HTTP request due to singleflight, got %d", count)
	}

	// 2. Second sequential call should directly hit disk cache without HTTP request
	path, err := downloadLuaArchive(ctx, server.URL, filename)
	if err != nil {
		t.Fatalf("disk cached download error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist at %s: %v", path, err)
	}
	if count := atomic.LoadInt32(&requestCount); count != 1 {
		t.Errorf("expected still 1 HTTP request due to disk cache hit, got %d", count)
	}
}
