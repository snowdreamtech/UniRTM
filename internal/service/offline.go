// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

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
}

// NewOfflineManager creates a new OfflineManager.
func NewOfflineManager() *OfflineManager {
	return &OfflineManager{
		probeURLs: []string{
			"https://api.github.com",
			"https://aquaproj.github.io",
		},
		timeout:       5 * time.Second,
		cacheDuration: 30 * time.Second,
	}
}

// IsOnline checks whether network connectivity is available.
//
// It caches the result for cacheDuration to avoid repeated probes.
//
// Validates Requirement: 19.1 (network availability detection)
func (om *OfflineManager) IsOnline(ctx context.Context) bool {
	// Return cached status if recent enough
	if om.cachedStatus != nil && time.Since(om.cachedAt) < om.cacheDuration {
		return *om.cachedStatus
	}

	client := &http.Client{
		Timeout: om.timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 0,
			}).DialContext,
		},
	}

	online := false
	for _, url := range om.probeURLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			online = true
			break
		}
	}

	om.cachedStatus = &online
	om.cachedAt = time.Now()

	logger.Debug("Network status checked", map[string]interface{}{
		"online": online,
	})

	return online
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
