// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"sync"
	"time"
)

// RouteStrategy defines the routing strategy chosen for an apex domain.
type RouteStrategy int

const (
	// RouteDirect indicates direct connection without any proxy.
	RouteDirect RouteStrategy = iota
	// RouteProxy indicates routing through the user-configured HTTP/HTTPS proxy.
	RouteProxy
)

// String returns the string representation of the route strategy.
func (s RouteStrategy) String() string {
	switch s {
	case RouteDirect:
		return "DIRECT"
	case RouteProxy:
		return "PROXY"
	default:
		return "UNKNOWN"
	}
}

// RouteEntry stores the strategy and expiration timestamp for a domain.
type RouteEntry struct {
	Strategy  RouteStrategy
	ExpiresAt time.Time
}

// DomainRouteCache provides thread-safe, TTL-based caching for domain routing decisions.
type DomainRouteCache struct {
	mu         sync.RWMutex
	routes     map[string]RouteEntry
	defaultTTL time.Duration
}

// NewDomainRouteCache creates a new DomainRouteCache with the specified default TTL.
func NewDomainRouteCache(ttl time.Duration) *DomainRouteCache {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &DomainRouteCache{
		routes:     make(map[string]RouteEntry),
		defaultTTL: ttl,
	}
}

var (
	defaultRouteCache     *DomainRouteCache
	defaultRouteCacheOnce sync.Once
)

// DefaultRouteCache returns the shared global singleton instance of DomainRouteCache.
func DefaultRouteCache() *DomainRouteCache {
	defaultRouteCacheOnce.Do(func() {
		defaultRouteCache = NewDomainRouteCache(30 * time.Minute)
	})
	return defaultRouteCache
}

// ResetDefaultRouteCache resets the global singleton instance (primarily for testing).
func ResetDefaultRouteCache() {
	defaultRouteCacheOnce = sync.Once{}
	defaultRouteCache = nil
}

// Get returns the cached routing strategy for the host's apex domain if valid.
func (c *DomainRouteCache) Get(host string) (RouteStrategy, bool) {
	apex := ExtractApexDomain(host)
	if apex == "" {
		return RouteDirect, false
	}

	c.mu.RLock()
	entry, exists := c.routes[apex]
	c.mu.RUnlock()

	if !exists {
		return RouteDirect, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.routes, apex)
		c.mu.Unlock()
		return RouteDirect, false
	}

	return entry.Strategy, true
}

// Set stores the routing strategy for the host's apex domain with the default TTL.
func (c *DomainRouteCache) Set(host string, strategy RouteStrategy) {
	c.SetWithTTL(host, strategy, c.defaultTTL)
}

// SetWithTTL stores the routing strategy for the host's apex domain with a custom TTL.
func (c *DomainRouteCache) SetWithTTL(host string, strategy RouteStrategy, ttl time.Duration) {
	apex := ExtractApexDomain(host)
	if apex == "" {
		return
	}

	c.mu.Lock()
	c.routes[apex] = RouteEntry{
		Strategy:  strategy,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Invalidate removes the routing decision for the host's apex domain.
// This is typically called when an active connection encounters network failures or RSTs.
func (c *DomainRouteCache) Invalidate(host string) {
	apex := ExtractApexDomain(host)
	if apex == "" {
		return
	}

	c.mu.Lock()
	delete(c.routes, apex)
	c.mu.Unlock()
}

// Clear flushes all cached routing entries.
func (c *DomainRouteCache) Clear() {
	c.mu.Lock()
	c.routes = make(map[string]RouteEntry)
	c.mu.Unlock()
}
