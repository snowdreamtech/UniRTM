// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

// PluginAPIVersion is the current stable plugin API version.
// Plugins must export this constant to declare compatibility.
const PluginAPIVersion = "1"

// PluginType indicates the role of a plugin.
type PluginType string

const (
	PluginTypeBackend  PluginType = "backend"
	PluginTypeProvider PluginType = "provider"
)

// PluginMetadata describes a loaded plugin.
//
// Validates Requirement: 22.3 (stable plugin API)
type PluginMetadata struct {
	Name       string
	Type       PluginType
	APIVersion string
	Path       string
}

// BackendFactory is the function signature that backend plugins must export
// as the symbol "NewBackend".
//
// Validates Requirement: 22.3 (stable plugin API)
type BackendFactory func() backend.Backend

// ProviderFactory is the function signature that provider plugins must export
// as the symbol "NewProvider".
type ProviderFactory func() provider.Provider

// PluginManager loads and manages backend and provider plugins.
//
// It discovers .so files in the plugins directory, validates their API version,
// and registers them with the backend and provider registries.
//
// Validates Requirements: 22.1, 22.2, 22.3, 22.4, 22.5, 22.6
type PluginManager struct {
	mu               sync.RWMutex
	pluginsDir       string
	loadedPlugins    []PluginMetadata
	backendRegistry  *backend.Registry
	providerRegistry *provider.Registry
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(pm.pluginsDir, entry.Name())

		// Validates Requirement: 22.5 (isolate plugin failures — one bad plugin doesn't block others)
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

// loadPlugin loads a single plugin file, validates its API version, and
// registers it with the appropriate registry.
//
// Validates Requirements: 22.3 (API version), 22.4 (validate compatibility)
func (pm *PluginManager) loadPlugin(ctx context.Context, pluginPath string) error {
	// Note: plugin.Open only works on Linux/macOS with CGO.
	// On Windows or when CGO is disabled, this will error gracefully.
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("open plugin: %w", err)
	}

	// Validates Requirement: 22.4 — check API version compatibility
	apiVersionSym, err := p.Lookup("PluginAPIVersion")
	if err != nil {
		return fmt.Errorf("plugin missing required symbol 'PluginAPIVersion': %w", err)
	}
	apiVersion, ok := apiVersionSym.(*string)
	if !ok || *apiVersion != PluginAPIVersion {
		return fmt.Errorf("plugin API version mismatch: expected %s, got %v", PluginAPIVersion, apiVersionSym)
	}

	// Determine plugin type
	pluginTypeSym, err := p.Lookup("PluginType")
	if err != nil {
		return fmt.Errorf("plugin missing required symbol 'PluginType': %w", err)
	}
	pluginType, ok := pluginTypeSym.(*string)
	if !ok {
		return fmt.Errorf("plugin 'PluginType' symbol has unexpected type")
	}

	// Plugin name
	pluginNameSym, err := p.Lookup("PluginName")
	if err != nil {
		return fmt.Errorf("plugin missing required symbol 'PluginName': %w", err)
	}
	pluginName, ok := pluginNameSym.(*string)
	if !ok {
		return fmt.Errorf("plugin 'PluginName' symbol has unexpected type")
	}

	meta := PluginMetadata{
		Name:       *pluginName,
		APIVersion: *apiVersion,
		Path:       pluginPath,
	}

	switch PluginType(*pluginType) {
	case PluginTypeBackend:
		if err := pm.loadBackendPlugin(p, meta); err != nil {
			return fmt.Errorf("register backend plugin %q: %w", *pluginName, err)
		}
	case PluginTypeProvider:
		if err := pm.loadProviderPlugin(p, meta); err != nil {
			return fmt.Errorf("register provider plugin %q: %w", *pluginName, err)
		}
	default:
		return fmt.Errorf("unknown plugin type %q", *pluginType)
	}

	pm.mu.Lock()
	pm.loadedPlugins = append(pm.loadedPlugins, meta)
	pm.mu.Unlock()

	logger.Info("Plugin loaded", map[string]interface{}{
		"name": *pluginName,
		"type": *pluginType,
		"path": pluginPath,
	})

	return nil
}

// loadBackendPlugin looks up and registers a backend plugin.
//
// Validates Requirement: 22.1 (load custom backend implementations)
func (pm *PluginManager) loadBackendPlugin(p *plugin.Plugin, meta PluginMetadata) error {
	factorySym, err := p.Lookup("NewBackend")
	if err != nil {
		return fmt.Errorf("backend plugin missing 'NewBackend' symbol: %w", err)
	}
	factory, ok := factorySym.(func() backend.Backend)
	if !ok {
		return fmt.Errorf("'NewBackend' has unexpected signature")
	}

	b := factory()
	if pm.backendRegistry != nil {
		// backend.Registry.Register uses Backend.Name() internally
		pm.backendRegistry.Register(b)
	}
	meta.Type = PluginTypeBackend
	return nil
}

// loadProviderPlugin looks up and registers a provider plugin.
//
// Validates Requirement: 22.2 (load custom provider implementations)
func (pm *PluginManager) loadProviderPlugin(p *plugin.Plugin, meta PluginMetadata) error {
	factorySym, err := p.Lookup("NewProvider")
	if err != nil {
		return fmt.Errorf("provider plugin missing 'NewProvider' symbol: %w", err)
	}
	factory, ok := factorySym.(func() provider.Provider)
	if !ok {
		return fmt.Errorf("'NewProvider' has unexpected signature")
	}

	prov := factory()
	if pm.providerRegistry != nil {
		// provider.Registry.Register(toolName, Provider)
		pm.providerRegistry.Register(meta.Name, prov)
	}
	meta.Type = PluginTypeProvider
	return nil
}

// ListLoaded returns metadata for all successfully loaded plugins.
func (pm *PluginManager) ListLoaded() []PluginMetadata {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]PluginMetadata, len(pm.loadedPlugins))
	copy(result, pm.loadedPlugins)
	return result
}

// PluginsDir returns the configured plugins directory path.
func (pm *PluginManager) PluginsDir() string {
	return pm.pluginsDir
}
