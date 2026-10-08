// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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
	Strategy  RouteStrategy `json:"strategy"`
	ExpiresAt time.Time     `json:"expires_at"`
}

// DomainRouteCache provides thread-safe, TTL-based caching for domain routing decisions
// with optional on-disk persistence across CLI executions.
type DomainRouteCache struct {
	mu         sync.RWMutex
	routes     map[string]RouteEntry
	defaultTTL time.Duration
	filePath   string
}

// NewDomainRouteCache creates an in-memory DomainRouteCache.
func NewDomainRouteCache(ttl time.Duration) *DomainRouteCache {
	return NewPersistentDomainRouteCache(ttl, "")
}

// NewPersistentDomainRouteCache creates a DomainRouteCache backed by a persistent file.
func NewPersistentDomainRouteCache(ttl time.Duration, filePath string) *DomainRouteCache {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	c := &DomainRouteCache{
		routes:     make(map[string]RouteEntry),
		defaultTTL: ttl,
		filePath:   filePath,
	}
	if filePath != "" {
		c.loadFromDisk()
	}
	return c
}

var (
	defaultRouteCache     *DomainRouteCache
	defaultRouteCacheOnce sync.Once
)

// DefaultRouteCache returns the shared persistent singleton instance of DomainRouteCache.
func DefaultRouteCache() *DomainRouteCache {
	defaultRouteCacheOnce.Do(func() {
		cacheDir := env.GetCacheDir()
		var p string
		if cacheDir != "" {
			p = filepath.Join(cacheDir, "network", "routes.json")
		}
		defaultRouteCache = NewPersistentDomainRouteCache(30*time.Minute, p)
	})
	return defaultRouteCache
}

// ResetDefaultRouteCache resets the global singleton instance (primarily for testing).
func ResetDefaultRouteCache() {
	defaultRouteCacheOnce = sync.Once{}
	defaultRouteCache = nil
}

func (c *DomainRouteCache) loadFromDisk() {
	if c.filePath == "" {
		return
	}
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return
	}
	var loaded map[string]RouteEntry
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}

	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range loaded {
		if now.Before(v.ExpiresAt) {
			c.routes[k] = v
		}
	}
}

func (c *DomainRouteCache) saveToDiskLocked() {
	if c.filePath == "" {
		return
	}
	now := time.Now()
	active := make(map[string]RouteEntry)
	for k, v := range c.routes {
		if now.Before(v.ExpiresAt) {
			active[k] = v
		}
	}
	data, err := json.MarshalIndent(active, "", "  ")
	if err != nil {
		return
	}

	dir := filepath.Dir(c.filePath)
	_ = os.MkdirAll(dir, 0755)

	tmpFile := c.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err == nil {
		_ = os.Rename(tmpFile, c.filePath)
	}
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
		c.saveToDiskLocked()
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
	c.saveToDiskLocked()
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
	c.saveToDiskLocked()
	c.mu.Unlock()
}

// Clear flushes all cached routing entries.
func (c *DomainRouteCache) Clear() {
	c.mu.Lock()
	c.routes = make(map[string]RouteEntry)
	c.saveToDiskLocked()
	c.mu.Unlock()
}
