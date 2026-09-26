// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOfflineManager_IsOnline(t *testing.T) {
	ClearOfflineDiskCache()
	defer ClearOfflineDiskCache()
	om := NewOfflineManager()

	// Initially we don't know if online without actual network,
	// but we can test with a mocked local server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	om.probeURLs = []string{ts.URL}

	if !om.IsOnline(context.Background()) {
		t.Error("expected IsOnline to return true with working server")
	}

	// Should cache
	om.probeURLs = []string{"http://invalid.local.url.test"}
	if !om.IsOnline(context.Background()) {
		t.Error("expected IsOnline to use cached true value")
	}

	// Force cache expiration
	om.cachedAt = time.Now().Add(-1 * time.Hour)
	ClearOfflineDiskCache()
	if om.IsOnline(context.Background()) {
		t.Error("expected IsOnline to return false after cache expiration with bad URL")
	}
}

func TestOfflineManager_RequireOnline(t *testing.T) {
	om := NewOfflineManager()

	om.cachedStatus = new(bool)
	*om.cachedStatus = true
	om.cachedAt = time.Now()

	err := om.RequireOnline(context.Background(), "install")
	if err != nil {
		t.Errorf("expected no error when online, got %v", err)
	}

	*om.cachedStatus = false
	err = om.RequireOnline(context.Background(), "install")
	if err == nil {
		t.Error("expected error when offline")
	}
}

func TestOfflineManager_CanOperateOffline(t *testing.T) {
	om := NewOfflineManager()

	tests := []struct {
		op   string
		want bool
	}{
		{"list", true},
		{"activate", true},
		{"install", false},
		{"update", false},
	}

	for _, tt := range tests {
		if got := om.CanOperateOffline(tt.op); got != tt.want {
			t.Errorf("CanOperateOffline(%q) = %v, want %v", tt.op, got, tt.want)
		}
	}
}

func TestOfflineManager_SkipIfOffline(t *testing.T) {
	om := NewOfflineManager()

	om.cachedStatus = new(bool)
	*om.cachedStatus = true
	om.cachedAt = time.Now()

	if om.SkipIfOffline(context.Background(), "test") {
		t.Error("expected SkipIfOffline to return false when online")
	}

	*om.cachedStatus = false
	if !om.SkipIfOffline(context.Background(), "test") {
		t.Error("expected SkipIfOffline to return true when offline")
	}
}

func TestOfflineManager_Concurrent(t *testing.T) {
	ClearOfflineDiskCache()
	defer ClearOfflineDiskCache()
	om := NewOfflineManager()

	var reqCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	om.probeURLs = []string{ts.URL}

	var wg sync.WaitGroup
	results := make([]bool, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = om.IsOnline(context.Background())
		}(i)
	}
	wg.Wait()

	for i := 0; i < 10; i++ {
		if !results[i] {
			t.Errorf("worker %d got false, want true", i)
		}
	}

	if atomic.LoadInt32(&reqCount) != 1 {
		t.Errorf("expected 1 probe request due to singleflight, got %d", reqCount)
	}
}

func TestOfflineManager_DiskCache(t *testing.T) {
	ClearOfflineDiskCache()
	defer ClearOfflineDiskCache()

	// Write status to disk
	writeOfflineDiskCache(true)

	// Create a brand new OfflineManager without in-memory cache
	om := NewOfflineManager()
	om.probeURLs = []string{"http://unreachable.test.domain.invalid"}

	// Should hit disk cache and return true
	if !om.IsOnline(context.Background()) {
		t.Error("expected IsOnline to hit disk cache and return true")
	}
}
