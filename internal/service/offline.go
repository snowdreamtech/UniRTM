// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"golang.org/x/sync/singleflight"
)

type offlineDiskCache struct {
	Online   bool      `json:"online"`
	CachedAt time.Time `json:"cached_at"`
}

func getOfflineDiskCachePath() string {
	return filepath.Join(env.GetCacheDir(), "network", "online_status.json")
}

func readOfflineDiskCache(ttl time.Duration) (bool, bool) {
	p := getOfflineDiskCachePath()
	data, err := os.ReadFile(p)
	if err != nil {
		return false, false
	}
	var c offlineDiskCache
	if err := json.Unmarshal(data, &c); err != nil {
		return false, false
	}
	if time.Since(c.CachedAt) > ttl {
		return false, false
	}
	return c.Online, true
}

func writeOfflineDiskCache(online bool) {
	p := getOfflineDiskCachePath()
	c := offlineDiskCache{
		Online:   online,
		CachedAt: time.Now(),
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err == nil {
		_ = os.Rename(tmp, p)
	}
}

// ClearOfflineDiskCache clears the persistent network status cache. Mainly used for testing.
func ClearOfflineDiskCache() {
	_ = os.Remove(getOfflineDiskCachePath())
}

// OfflineManager detects network availability and provides offline operation support.
//
// Validates Requirements: 19.1, 19.2, 19.3, 19.4, 19.5, 19.6, 19.7
type OfflineManager struct {
	// probeURLs are the URLs used to check connectivity.
	probeURLs []string
	// timeout is the HTTP probe timeout.
	timeout time.Duration
	// cachedStatus is the last known network status.
	cachedStatus *bool
	// cachedAt is when the status was last checked.
	cachedAt time.Time
	// cacheDuration is how long to cache the status check.
	cacheDuration time.Duration

	mu     sync.RWMutex
	flight singleflight.Group
	client *http.Client
}

// NewOfflineManager creates a new OfflineManager.
func NewOfflineManager() *OfflineManager {
	timeout := 5 * time.Second
	return &OfflineManager{
		probeURLs: []string{
			"https://api.github.com",
			"https://aquaproj.github.io",
		},
		timeout:       timeout,
		cacheDuration: 30 * time.Second,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   3 * time.Second,
					KeepAlive: 0,
				}).DialContext,
			},
		},
	}
}

// IsOnline checks whether network connectivity is available.
//
// It caches the result for cacheDuration to avoid repeated probes.
//
// Validates Requirement: 19.1 (network availability detection)
func (om *OfflineManager) IsOnline(ctx context.Context) bool {
	// Return cached status if recent enough (read lock)
	om.mu.RLock()
	if om.cachedStatus != nil && time.Since(om.cachedAt) < om.cacheDuration {
		status := *om.cachedStatus
		om.mu.RUnlock()
		return status
	}
	om.mu.RUnlock()

	// Check cross-process persistent disk cache
	if status, ok := readOfflineDiskCache(om.cacheDuration); ok {
		om.mu.Lock()
		om.cachedStatus = &status
		om.cachedAt = time.Now()
		om.mu.Unlock()
		return status
	}

	res, _, _ := om.flight.Do("is_online", func() (interface{}, error) {
		// Double check after flight acquisition
		om.mu.RLock()
		if om.cachedStatus != nil && time.Since(om.cachedAt) < om.cacheDuration {
			status := *om.cachedStatus
			om.mu.RUnlock()
			return status, nil
		}
		om.mu.RUnlock()

		if status, ok := readOfflineDiskCache(om.cacheDuration); ok {
			om.mu.Lock()
			om.cachedStatus = &status
			om.cachedAt = time.Now()
			om.mu.Unlock()
			return status, nil
		}

		client := om.client
		if client == nil {
			client = &http.Client{
				Timeout: om.timeout,
				Transport: &http.Transport{
					DialContext: (&net.Dialer{
						Timeout:   3 * time.Second,
						KeepAlive: 0,
					}).DialContext,
				},
			}
		}

		online := false
		if len(om.probeURLs) == 1 {
			req, err := http.NewRequestWithContext(ctx, http.MethodHead, om.probeURLs[0], nil)
			if err == nil {
				if resp, err := client.Do(req); err == nil {
					resp.Body.Close()
					online = true
				}
			}
		} else if len(om.probeURLs) > 1 {
			probeCtx, cancelProbe := context.WithCancel(ctx)
			defer cancelProbe()

			var wg sync.WaitGroup
			var once sync.Once
			for _, probeURL := range om.probeURLs {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					req, err := http.NewRequestWithContext(probeCtx, http.MethodHead, url, nil)
					if err != nil {
						return
					}
					resp, err := client.Do(req)
					if err == nil {
						resp.Body.Close()
						once.Do(func() {
							online = true
							cancelProbe()
						})
					}
				}(probeURL)
			}
			wg.Wait()
		}

		om.mu.Lock()
		om.cachedStatus = &online
		om.cachedAt = time.Now()
		om.mu.Unlock()

		writeOfflineDiskCache(online)

		logger.Debug("Network status checked", map[string]interface{}{
			"online": online,
		})

		return online, nil
	})

	return res.(bool)
}

// RequireOnline checks network connectivity and returns an error if offline.
//
// Validates Requirement: 19.6 (clear feedback when network required but unavailable)
func (om *OfflineManager) RequireOnline(ctx context.Context, operation string) error {
	if !om.IsOnline(ctx) {
		return fmt.Errorf(
			"operation %q requires network access, but no connection is available.\n"+
				"  Tip: Install tools while online so they are cached for offline use.\n"+
				"  Tip: Use locally cached index with: unirtm search (if index is fresh)",
			operation,
		)
	}
	return nil
}

// CanOperateOffline reports whether the given operation can be performed offline.
//
// Validates Requirements: 19.2, 19.3, 19.4, 19.5
func (om *OfflineManager) CanOperateOffline(operation string) bool {
	offlineOps := map[string]bool{
		"list":       true,  // Req 19.2: list installed tools
		"activate":   true,  // Req 19.3: activate installed tools
		"deactivate": true,  // Req 19.3: deactivate tools
		"doctor":     true,  // Req 19.4: local diagnostics
		"config":     true,  // local config management
		"search":     true,  // Req 19.5: search cached index
		"cache":      true,  // local cache management
		"install":    false, // requires download
		"update":     false, // requires version check
		"index":      false, // requires index refresh
	}
	return offlineOps[operation]
}

// SkipIfOffline logs a message and returns true if the operation should be
// skipped because we are offline. Used for optional network operations.
//
// Validates Requirement: 19.7 (skip optional network operations when offline)
func (om *OfflineManager) SkipIfOffline(ctx context.Context, operationName string) bool {
	if !om.IsOnline(ctx) {
		logger.Info("Skipping optional network operation (offline)", map[string]interface{}{
			"operation": operationName,
		})
		return true
	}
	return false
}
