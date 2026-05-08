// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides the plugin loading and management capabilities.
// It relies on HashiCorp go-plugin to dynamically load standalone binaries
// that implement the Backend or Provider interfaces via RPC.
package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	uplugin "github.com/snowdreamtech/unirtm/internal/plugin"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

// PluginAPIVersion defines the current plugin API version for compatibility checking.
//
// Validates Requirement: 22.3 (API stability and versioning)
const PluginAPIVersion = "v1"

// PluginMetadata holds information about a loaded plugin.
type PluginMetadata struct {
	Name       string
	Type       string // "backend" or "provider"
	APIVersion string
	Path       string
}

// PluginManager handles discovery, loading, and registration of Go plugins.
//
// Validates Requirements: 22.1, 22.2
type PluginManager struct {
	mu               sync.RWMutex
	pluginsDir       string
	loadedPlugins    []PluginMetadata
	backendRegistry  *backend.Registry
	providerRegistry *provider.Registry
	clients          []*hplugin.Client
}

// NewPluginManager creates a new PluginManager.
//
// Validates Requirements: 22.1, 22.2
func NewPluginManager(pluginsDir string, br *backend.Registry, pr *provider.Registry) *PluginManager {
	return &PluginManager{
		pluginsDir:       pluginsDir,
		backendRegistry:  br,
		providerRegistry: pr,
	}
}

// LoadAll discovers and loads all plugins from the plugins directory.
//
// Validates Requirements: 22.1, 22.4 (validate compatibility), 22.5 (isolate failures)
func (pm *PluginManager) LoadAll(ctx context.Context) error {
	entries, err := os.ReadDir(pm.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("Plugins directory does not exist — skipping plugin loading", map[string]interface{}{
				"dir": pm.pluginsDir,
			})
			return nil
		}
		return fmt.Errorf("read plugins directory: %w", err)
	}

	for _, entry := range entries {
		// Validates Requirement: 22.5 (isolate plugin failures — one bad plugin doesn't block others)
		if entry.IsDir() || (!strings.HasPrefix(entry.Name(), "unirtm-plugin-") && !strings.HasSuffix(entry.Name(), ".exe")) {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "unirtm-plugin-") {
			continue
		}

		pluginPath := filepath.Join(pm.pluginsDir, entry.Name())

		if err := pm.loadPlugin(ctx, pluginPath); err != nil {
			logger.Error("Failed to load plugin (skipping)", map[string]interface{}{
				"path":  pluginPath,
				"error": err.Error(),
			})
		}
	}

	logger.Info("Plugin loading complete", map[string]interface{}{
		"loaded": len(pm.loadedPlugins),
	})

	return nil
}

// loadPlugin loads a single plugin binary, starts the RPC server, and
// registers it with the appropriate registry.
//
// Validates Requirements: 22.3 (API version), 22.4 (validate compatibility)
func (pm *PluginManager) loadPlugin(ctx context.Context, pluginPath string) error {
	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig:  uplugin.HandshakeConfig,
		Plugins:          uplugin.PluginMap,
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolNetRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to start plugin client: %w", err)
	}

	// Try to dispense a backend plugin first
	raw, err := rpcClient.Dispense("backend")
	if err == nil {
		b, ok := raw.(backend.Backend)
		if ok {
			meta := PluginMetadata{
				Name:       b.Name(),
				Type:       "backend",
				APIVersion: PluginAPIVersion,
				Path:       pluginPath,
			}
			pm.mu.Lock()
			pm.clients = append(pm.clients, client)
			pm.loadedPlugins = append(pm.loadedPlugins, meta)
			if pm.backendRegistry != nil {
				pm.backendRegistry.Register(b)
				logger.Debug("Registered plugin backend", map[string]interface{}{
					"name": meta.Name,
				})
			}
			pm.mu.Unlock()
			return nil
		}
	}

	// Try provider plugin
	raw, err = rpcClient.Dispense("provider")
	if err == nil {
		p, ok := raw.(provider.Provider)
		if ok {
			meta := PluginMetadata{
				Name:       p.Name(),
				Type:       "provider",
				APIVersion: PluginAPIVersion,
				Path:       pluginPath,
			}
			pm.mu.Lock()
			pm.clients = append(pm.clients, client)
			pm.loadedPlugins = append(pm.loadedPlugins, meta)
			if pm.providerRegistry != nil {
				pm.providerRegistry.Register(p)
				logger.Debug("Registered plugin provider", map[string]interface{}{
					"name": meta.Name,
				})
			}
			pm.mu.Unlock()
			return nil
		}
	}

	client.Kill()
	return fmt.Errorf("plugin does not dispense 'backend' or 'provider'")
}

// ListLoaded returns metadata for all loaded plugins.
func (pm *PluginManager) ListLoaded() []PluginMetadata {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent data races
	result := make([]PluginMetadata, len(pm.loadedPlugins))
	copy(result, pm.loadedPlugins)
	return result
}

// PluginsDir returns the directory where plugins are loaded from.
func (pm *PluginManager) PluginsDir() string {
	return pm.pluginsDir
}

// Cleanup kills all the running plugin processes.
func (pm *PluginManager) Cleanup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, client := range pm.clients {
		client.Kill()
	}
	pm.clients = nil
}
