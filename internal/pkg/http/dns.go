// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

// DNSCache provides a thread-safe in-memory DNS lookup cache with TTL and singleflight deduplication.
type DNSCache struct {
	mu       sync.RWMutex
	cache    map[string]dnsCacheEntry
	flight   singleflight.Group
	ttl      time.Duration
	resolver *net.Resolver
}

var (
	defaultDNSCache     *DNSCache
	defaultDNSCacheOnce sync.Once
)

// DefaultDNSCache returns the process-wide DNS cache with a default 2-minute TTL.
func DefaultDNSCache() *DNSCache {
	defaultDNSCacheOnce.Do(func() {
		defaultDNSCache = NewDNSCache(2 * time.Minute)
	})
	return defaultDNSCache
}

// NewDNSCache creates a new in-memory DNS cache with the specified TTL.
func NewDNSCache(ttl time.Duration) *DNSCache {
	return &DNSCache{
		cache:    make(map[string]dnsCacheEntry),
		ttl:      ttl,
		resolver: net.DefaultResolver,
	}
}

// LookupIP looks up IP addresses for host, using the cache if available and not expired.
func (c *DNSCache) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	// If it's already an IP address, return immediately
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	c.mu.RLock()
	entry, found := c.cache[host]
	c.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		return entry.ips, nil
	}

	// Singleflight resolution
	res, err, _ := c.flight.Do(host, func() (interface{}, error) {
		c.mu.RLock()
		entry, found := c.cache[host]
		c.mu.RUnlock()
		if found && time.Now().Before(entry.expiresAt) {
			return entry.ips, nil
		}

		ips, err := c.resolver.LookupIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.cache[host] = dnsCacheEntry{
			ips:       ips,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.mu.Unlock()

		return ips, nil
	})

	if err != nil {
		return nil, err
	}
	return res.([]net.IP), nil
}

// Evict removes host from the cache.
func (c *DNSCache) Evict(host string) {
	c.mu.Lock()
	delete(c.cache, host)
	c.mu.Unlock()
}

// Clear clears all entries in the cache.
func (c *DNSCache) Clear() {
	c.mu.Lock()
	c.cache = make(map[string]dnsCacheEntry)
	c.mu.Unlock()
}
