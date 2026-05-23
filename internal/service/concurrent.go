// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ConcurrentInstallResult represents the result of a single concurrent installation.
type ConcurrentInstallResult struct {
	Tool    string
	Version string
	Success bool
	Error   string
}

// ConcurrentManager manages parallel tool installations with controlled concurrency.
//
// It respects dependency order, limits concurrent operations to avoid resource
// exhaustion, serializes database writes, and reports progress per-operation.
//
// Validates Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6, 18.7
type ConcurrentManager struct {
	// installManager handles the actual installation of each tool.
	installManager *InstallationManager
	// maxConcurrency is the maximum number of parallel installations.
	maxConcurrency int
	// progressFn is an optional callback for progress reporting.
	progressFn func(tool, version, status string)
}

// ConcurrentManagerConfig holds configuration for ConcurrentManager.
type ConcurrentManagerConfig struct {
	// MaxConcurrency is the maximum parallel operations (0 = CPU count).
	MaxConcurrency int
	// ProgressFn is an optional progress callback.
	ProgressFn func(tool, version, status string)
}

// NewConcurrentManager creates a new ConcurrentManager.
//
// Validates Requirement: 18.2 (configurable concurrency limit)
func NewConcurrentManager(im *InstallationManager, config ConcurrentManagerConfig) *ConcurrentManager {
	maxConcurrency := config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU()
	}

	return &ConcurrentManager{
		installManager: im,
		maxConcurrency: maxConcurrency,
		progressFn:     config.ProgressFn,
	}
}

// ToolInstallRequest specifies a tool and version to install.
type ToolInstallRequest struct {
	ToolKey   string // The full key as specified in the configuration
	Tool      string // The stripped tool name
	Version   string
	Backend   string
	// DependsOn lists tool names that must be installed before this one.
	DependsOn []string
}

// InstallAll installs multiple tools concurrently, respecting dependency order.
//
// It:
//  1. Topologically sorts by DependsOn to determine installation order (Req 18.6)
//  2. Uses errgroup for parallel execution within each dependency layer (Req 18.1)
//  3. Limits concurrency to maxConcurrency (Req 18.2)
//  4. Reports progress via progressFn (Req 18.5)
//  5. Cancels dependents if a dependency fails (Req 18.7)
//
// Validates Requirements: 18.1, 18.2, 18.4, 18.5, 18.6, 18.7
func (cm *ConcurrentManager) InstallAll(ctx context.Context, requests []ToolInstallRequest) ([]ConcurrentInstallResult, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	// Suppress individual interactive download/resolve progress UI during concurrent execution
	ctx = context.WithValue(ctx, ContextKeyQuietProgress, true)

	// Topological sort to determine installation layers
	layers, err := topoSort(requests)
	if err != nil {
		return nil, fmt.Errorf("resolve installation order: %w", err)
	}

	var (
		resultsMu sync.Mutex
		results   []ConcurrentInstallResult
		failedSet = make(map[string]bool)
	)

	// Semaphore to limit concurrency (Req 18.2)
	sem := make(chan struct{}, cm.maxConcurrency)

	// Process each layer sequentially; within a layer, run in parallel (Req 18.1)
	for _, layer := range layers {
		g, gctx := errgroup.WithContext(ctx)

		for i := range layer {
			req := layer[i] // capture loop variable

			// Skip if a dependency failed (Req 18.7)
			dependencyFailed := false
			for _, dep := range req.DependsOn {
				if failedSet[dep] {
					dependencyFailed = true
					break
				}
			}
			if dependencyFailed {
				resultsMu.Lock()
				results = append(results, ConcurrentInstallResult{
					Tool:    req.Tool,
					Version: req.Version,
					Success: false,
					Error:   "skipped: a dependency failed",
				})
				failedSet[req.Tool] = true
				resultsMu.Unlock()
				continue
			}

			// Skip if already installed
			isInstalled, _ := cm.installManager.IsInstalled(ctx, req.Tool, req.Version, req.Backend)
			if isInstalled {
				resultsMu.Lock()
				results = append(results, ConcurrentInstallResult{
					Tool:    req.Tool,
					Version: req.Version,
					Success: true,
					Error:   ErrAlreadyInstalled.Error(),
				})
				resultsMu.Unlock()
				cm.reportProgress(req.Tool, req.Version, "failed: "+ErrAlreadyInstalled.Error())
				continue
			}

			g.Go(func() error {
				// Acquire semaphore slot
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-gctx.Done():
					return gctx.Err()
				}

				cm.reportProgress(req.Tool, req.Version, "starting")

				// Validates Req 18.3: serialized database writes handled by InstallationManager internally
				installErr := cm.installManager.Install(gctx, req.ToolKey, req.Tool, req.Version, req.Backend)

				var result ConcurrentInstallResult
				result.Tool = req.Tool
				result.Version = req.Version

				isAlreadyInstalled := installErr != nil && (installErr == ErrAlreadyInstalled || strings.Contains(installErr.Error(), "already installed"))

				if installErr != nil && !isAlreadyInstalled {
					result.Success = false
					result.Error = installErr.Error()
					cm.reportProgress(req.Tool, req.Version, "failed: "+installErr.Error())
				} else {
					result.Success = true
					if isAlreadyInstalled {
						result.Error = installErr.Error()
						cm.reportProgress(req.Tool, req.Version, "failed: "+installErr.Error())
					} else {
						cm.reportProgress(req.Tool, req.Version, "done")
					}
				}

				resultsMu.Lock()
				results = append(results, result)
				if !result.Success {
					failedSet[req.Tool] = true
				}
				resultsMu.Unlock()

				// Validates Req 18.4: other tools continue even if one fails
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			// Context cancelled (e.g., by signal)
			return results, fmt.Errorf("installation interrupted: %w", err)
		}
	}

	return results, nil
}

// reportProgress calls the progress callback if set.
func (cm *ConcurrentManager) reportProgress(tool, version, status string) {
	if cm.progressFn != nil {
		cm.progressFn(tool, version, status)
	}
}

// topoSort performs a topological sort of install requests by DependsOn.
// Returns a slice of layers; each layer can be installed in parallel.
func topoSort(requests []ToolInstallRequest) ([][]ToolInstallRequest, error) {
	byName := make(map[string]ToolInstallRequest)
	for _, r := range requests {
		byName[r.Tool] = r
	}

	inDegree := make(map[string]int)
	for _, r := range requests {
		if _, ok := inDegree[r.Tool]; !ok {
			inDegree[r.Tool] = 0
		}
		for _, dep := range r.DependsOn {
			if _, ok := byName[dep]; ok {
				inDegree[r.Tool]++
			}
		}
	}

	dependents := make(map[string][]string)
	for _, r := range requests {
		for _, dep := range r.DependsOn {
			if _, ok := byName[dep]; ok {
				dependents[dep] = append(dependents[dep], r.Tool)
			}
		}
	}

	var layers [][]ToolInstallRequest
	remaining := make(map[string]bool)
	for name := range byName {
		remaining[name] = true
	}

	for len(remaining) > 0 {
		var layer []ToolInstallRequest
		for name := range remaining {
			if inDegree[name] == 0 {
				layer = append(layer, byName[name])
			}
		}

		if len(layer) == 0 {
			var cycle []string
			for name := range remaining {
				cycle = append(cycle, name)
			}
			return nil, fmt.Errorf("circular dependency detected involving: %v", cycle)
		}

		for _, r := range layer {
			delete(remaining, r.Tool)
			for _, dep := range dependents[r.Tool] {
				inDegree[dep]--
			}
		}

		layers = append(layers, layer)
	}

	return layers, nil
}
