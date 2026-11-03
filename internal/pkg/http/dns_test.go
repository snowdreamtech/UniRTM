// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDNSCache_LookupIP_IPAddress(t *testing.T) {
	cache := NewDNSCache(time.Minute)
	ctx := context.Background()

	ips, err := cache.LookupIP(ctx, "127.0.0.1")
	assert.NoError(t, err)
	assert.Len(t, ips, 1)
	assert.Equal(t, "127.0.0.1", ips[0].String())

	ipsV6, err := cache.LookupIP(ctx, "::1")
	assert.NoError(t, err)
	assert.Len(t, ipsV6, 1)
	assert.Equal(t, "::1", ipsV6[0].String())
}

func TestDNSCache_LookupIP_CachingAndTTL(t *testing.T) {
	cache := NewDNSCache(50 * time.Millisecond)
	ctx := context.Background()

	ips1, err := cache.LookupIP(ctx, "localhost")
	assert.NoError(t, err)
	assert.NotEmpty(t, ips1)

	// Second lookup should hit cache
	ips2, err := cache.LookupIP(ctx, "localhost")
	assert.NoError(t, err)
	assert.Equal(t, ips1, ips2)

	// Wait for TTL expiration
	time.Sleep(70 * time.Millisecond)
	ips3, err := cache.LookupIP(ctx, "localhost")
	assert.NoError(t, err)
	assert.NotEmpty(t, ips3)
}

func TestDNSCache_EvictAndClear(t *testing.T) {
	cache := NewDNSCache(time.Minute)
	ctx := context.Background()

	_, err := cache.LookupIP(ctx, "localhost")
	assert.NoError(t, err)

	cache.mu.RLock()
	_, found := cache.cache["localhost"]
	cache.mu.RUnlock()
	assert.True(t, found)

	cache.Evict("localhost")
	cache.mu.RLock()
	_, found = cache.cache["localhost"]
	cache.mu.RUnlock()
	assert.False(t, found)

	_, _ = cache.LookupIP(ctx, "localhost")
	cache.Clear()
	cache.mu.RLock()
	assert.Empty(t, cache.cache)
	cache.mu.RUnlock()
}

func TestDefaultTransport_DialContext_WithDNSCache(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	trans := DefaultTransport()
	assert.NotNil(t, trans.DialContext)

	client := &http.Client{
		Transport: trans,
	}

	resp, err := client.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
