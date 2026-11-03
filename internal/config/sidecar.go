// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// DiscoverSidecars scans the specified directory for plugin sidecar configuration files (*.toml).
// Usually pluginsDir is filepath.Join(projectRoot, ".unirtm", "plugins").
// Supports flat files (.unirtm/plugins/docker.toml) and nested vendor structures (.unirtm/plugins/vendor/docker.toml).
func DiscoverSidecars(pluginsDir string) (PluginMap, error) {
	result := make(PluginMap)

	info, err := OsStat(pluginsDir)
	if err != nil || !info.IsDir() {
		// Directory does not exist or is not a directory; silently return empty map
		return result, nil
	}

	err = filepath.WalkDir(pluginsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".toml") {
			return nil
		}

		decl, pluginID, parseErr := parseSidecarFile(pluginsDir, path)
		if parseErr != nil {
			return fmt.Errorf("failed to parse sidecar %s: %w", path, parseErr)
		}
		if pluginID != "" {
			result[pluginID] = decl
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DiscoverProjectSidecars scans `.unirtm/plugins` under the project directory.
func DiscoverProjectSidecars(projectDir string) (PluginMap, error) {
	pluginsDir := filepath.Join(projectDir, ".unirtm", "plugins")
	return DiscoverSidecars(pluginsDir)
}

// DiscoverGlobalSidecars scans `~/.config/unirtm/plugins` for global sidecar configurations.
func DiscoverGlobalSidecars() (PluginMap, error) {
	globalDir := filepath.Join(env.GetConfigDir(), "plugins")
	return DiscoverSidecars(globalDir)
}

// parseSidecarFile reads and unmarshals a sidecar file into a PluginDeclaration.
func parseSidecarFile(baseDir, filePath string) (PluginDeclaration, string, error) {
	data, err := OsReadFile(filePath)
	if err != nil {
		return PluginDeclaration{}, "", err
	}

	var raw map[string]interface{}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return PluginDeclaration{}, "", err
	}

	decl := PluginDeclaration{
		Config:      make(map[string]interface{}),
		Source:      "sidecar",
		SidecarPath: filePath,
	}

	// Determine plugin ID
	var pluginID string
	if idVal, ok := raw["id"].(string); ok && idVal != "" {
		pluginID = idVal
		delete(raw, "id")
	} else if idVal, ok := raw["plugin"].(string); ok && idVal != "" {
		pluginID = idVal
		delete(raw, "plugin")
	} else if idVal, ok := raw["plugin_id"].(string); ok && idVal != "" {
		pluginID = idVal
		delete(raw, "plugin_id")
	} else {
		// Infer from relative path under baseDir
		rel, err := filepath.Rel(baseDir, filePath)
		if err == nil {
			relNoExt := strings.TrimSuffix(rel, ".toml")
			pluginID = filepath.ToSlash(relNoExt)
		} else {
			pluginID = strings.TrimSuffix(filepath.Base(filePath), ".toml")
		}
	}

	// Check if version is declared in sidecar
	if vVal, ok := raw["version"].(string); ok {
		decl.Version = vVal
		delete(raw, "version")
	}

	// Check if config is nested inside a [config] section
	if cfgSub, ok := raw["config"].(map[string]interface{}); ok {
		decl.Config = cfgSub
	} else {
		// All remaining keys are considered plugin config
		decl.Config = raw
	}

	return decl, pluginID, nil
}

// MergeSidecarConfigs merges sidecar configurations into target Config.
// Inline configuration in target Config always takes precedence over sidecar configuration (Inline overrides Sidecar).
func MergeSidecarConfigs(target *Config, sidecars PluginMap) {
	if target == nil || len(sidecars) == 0 {
		return
	}

	if target.Plugins == nil {
		target.Plugins = make(PluginMap)
	}

	for pluginID, sidecarDecl := range sidecars {
		inlineDecl, exists := target.Plugins[pluginID]
		if !exists {
			// No inline declaration; take sidecar as-is
			target.Plugins[pluginID] = sidecarDecl
		} else {
			// Inline declaration exists -> Inline takes precedence over sidecar
			mergedCfg := make(map[string]interface{})

			// 1. Populate defaults from sidecar
			for k, v := range sidecarDecl.Config {
				mergedCfg[k] = v
			}

			// 2. Override with inline values (inline wins conflicts)
			for k, v := range inlineDecl.Config {
				mergedCfg[k] = v
			}

			version := inlineDecl.Version
			if version == "" {
				version = sidecarDecl.Version
			}

			target.Plugins[pluginID] = PluginDeclaration{
				Version:     version,
				Config:      mergedCfg,
				Source:      "merged",
				SidecarPath: sidecarDecl.SidecarPath,
			}
		}
	}
}
