// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDomainRouteCache_ApexDomainSharing(t *testing.T) {
	cache := NewDomainRouteCache(10 * time.Minute)

	// Set on one subdomain
	cache.Set("objects.githubusercontent.com", RouteProxy)

	// Another subdomain under same apex domain should hit the cache
	strat, ok := cache.Get("raw.githubusercontent.com")
	if !ok || strat != RouteProxy {
		t.Errorf("expected RouteProxy for raw.githubusercontent.com, got %v (ok=%v)", strat, ok)
	}

	// Different apex domain should not hit
	_, ok = cache.Get("api.github.com")
	if ok {
		t.Errorf("expected cache miss for api.github.com")
	}

	// Set github.com
	cache.Set("github.com", RouteDirect)
	strat, ok = cache.Get("api.github.com")
	if !ok || strat != RouteDirect {
		t.Errorf("expected RouteDirect for api.github.com, got %v (ok=%v)", strat, ok)
	}
}

func TestDomainRouteCache_Expiration(t *testing.T) {
	cache := NewDomainRouteCache(50 * time.Millisecond)

	cache.Set("example.com", RouteDirect)

	strat, ok := cache.Get("example.com")
	if !ok || strat != RouteDirect {
		t.Fatalf("expected immediate hit, got %v (ok=%v)", strat, ok)
	}

	time.Sleep(70 * time.Millisecond)

	_, ok = cache.Get("example.com")
	if ok {
		t.Errorf("expected cache entry to be expired")
	}
}

func TestDomainRouteCache_InvalidateAndClear(t *testing.T) {
	cache := NewDomainRouteCache(10 * time.Minute)

	cache.Set("mirrors.aliyun.com", RouteDirect)
	strat, ok := cache.Get("aliyun.com")
	if !ok || strat != RouteDirect {
		t.Fatalf("expected hit for aliyun.com")
	}

	// Invalidate via subdomain
	cache.Invalidate("oss.aliyun.com")
	_, ok = cache.Get("mirrors.aliyun.com")
	if ok {
		t.Errorf("expected aliyun.com to be invalidated")
	}

	cache.Set("foo.com", RouteDirect)
	cache.Set("bar.com", RouteProxy)
	cache.Clear()

	if _, ok := cache.Get("foo.com"); ok {
		t.Errorf("expected foo.com to be cleared")
	}
	if _, ok := cache.Get("bar.com"); ok {
		t.Errorf("expected bar.com to be cleared")
	}
}

func TestDomainRouteCache_DiskPersistence(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "network", "routes.json")

	// Process 1: populate cache and persist to disk
	cache1 := NewPersistentDomainRouteCache(10*time.Minute, tempFile)
	cache1.Set("objects.githubusercontent.com", RouteProxy)
	cache1.Set("mirrors.aliyun.com", RouteDirect)

	// Process 2: new instance reading from the same file
	cache2 := NewPersistentDomainRouteCache(10*time.Minute, tempFile)

	strat, ok := cache2.Get("raw.githubusercontent.com")
	if !ok || strat != RouteProxy {
		t.Errorf("expected RouteProxy for raw.githubusercontent.com from disk, got %v (ok=%v)", strat, ok)
	}

	strat, ok = cache2.Get("aliyun.com")
	if !ok || strat != RouteDirect {
		t.Errorf("expected RouteDirect for aliyun.com from disk, got %v (ok=%v)", strat, ok)
	}

	// Invalidate in Process 2 and verify Process 3 sees it removed
	cache2.Invalidate("aliyun.com")

	cache3 := NewPersistentDomainRouteCache(10*time.Minute, tempFile)
	if _, ok := cache3.Get("aliyun.com"); ok {
		t.Errorf("expected aliyun.com to remain invalidated on disk")
	}
}
