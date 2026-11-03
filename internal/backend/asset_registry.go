// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// AssetRegistryEntry describes known asset naming conventions for a specific tool.
type AssetRegistryEntry struct {
	// Patterns maps platform keys (e.g. "macos-amd64") to the exact asset filename.
	Patterns map[string]string `toml:"patterns"`

	// Description explains why the non-standard mapping is needed.
	Description string `toml:"description"`
}

// localRegistryFile is the structure of ~/.config/unirtm/asset-registry.toml.
//
// Example file:
//
//	[tools."editorconfig-checker/editorconfig-checker"]
//	description = "Uses darwin-all universal binary for macOS"
//
//	[tools."editorconfig-checker/editorconfig-checker".patterns]
//	"macos-amd64" = "editorconfig-checker-darwin-all.tar.gz"
//	"macos-arm64" = "editorconfig-checker-darwin-all.tar.gz"
type localRegistryFile struct {
	Tools map[string]AssetRegistryEntry `toml:"tools"`
}

// DefaultLocalRegistryPath returns the canonical path for the user-level
// asset registry file: $UNIRTM_CONFIG_DIR/asset-registry.toml.
func DefaultLocalRegistryPath() string {
	return filepath.Join(env.GetConfigDir(), "asset-registry.toml")
}

// LoadLocalAssetRegistry reads a local TOML asset registry file and returns its
// contents.  Returns (nil, nil) if the file does not exist (not an error).
func LoadLocalAssetRegistry(path string) (map[string]AssetRegistryEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var f localRegistryFile
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Tools, nil
}

// MergeAssetRegistries returns a new registry that merges local entries over
// the built-in ones.  localRegistry keys take precedence over builtin keys.
func MergeAssetRegistries(local, builtin map[string]AssetRegistryEntry) map[string]AssetRegistryEntry {
	merged := make(map[string]AssetRegistryEntry, len(builtin)+len(local))
	for k, v := range builtin {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v // local overrides built-in
	}
	return merged
}

// ─── Built-in registry ────────────────────────────────────────────────────────

// AssetRegistry is the built-in community-maintained registry of known asset
// naming quirks.  The map key is the tool identifier as it appears in
// .unirtm.toml (e.g. "editorconfig-checker/editorconfig-checker").
//
// Entries here are consulted automatically as a fallback when:
//   - The heuristic FindBestAsset returns nil (no matching asset)
//   - AND no user-level asset_patterns override is configured for the tool
//
// User-level asset_patterns and local asset-registry.toml entries always
// take precedence over this built-in registry.
var AssetRegistry = map[string]AssetRegistryEntry{
	// editorconfig-checker publishes a single darwin-all archive for both macOS arches.
	"editorconfig-checker/editorconfig-checker": {
		Description: "Uses darwin-all universal binary for macOS",
		Patterns: map[string]string{
			"macos-amd64": "editorconfig-checker-darwin-all.tar.gz",
			"macos-arm64": "editorconfig-checker-darwin-all.tar.gz",
		},
	},
	// cosign publishes raw binaries named cosign-<os>-<arch> with no archive.
	"sigstore/cosign": {
		Description: "Publishes raw binaries named cosign-<os>-<arch> with no archive",
		Patterns: map[string]string{
			"linux-amd64":   "cosign-linux-amd64",
			"linux-arm64":   "cosign-linux-arm64",
			"macos-amd64":   "cosign-darwin-amd64",
			"macos-arm64":   "cosign-darwin-arm64",
			"windows-amd64": "cosign-windows-amd64.exe",
		},
	},
}

// ─── Effective registry (built-in merged with local file) ─────────────────────

var (
	effectiveOnce     sync.Once
	effectiveRegistry map[string]AssetRegistryEntry
)

// EffectiveAssetRegistry returns the merged asset registry (local TOML over
// built-in).  The local file is loaded once from DefaultLocalRegistryPath() and
// cached for the lifetime of the process.
//
// Call InvalidateEffectiveRegistry() in tests to reset the cache.
func EffectiveAssetRegistry() map[string]AssetRegistryEntry {
	effectiveOnce.Do(func() {
		local, _ := LoadLocalAssetRegistry(DefaultLocalRegistryPath())
		effectiveRegistry = MergeAssetRegistries(local, AssetRegistry)
	})
	return effectiveRegistry
}

// InvalidateEffectiveRegistry resets the cached effective registry so that the
// next call to EffectiveAssetRegistry() reloads from disk.  Intended for tests.
func InvalidateEffectiveRegistry() {
	effectiveOnce = sync.Once{}
	effectiveRegistry = nil
}

// ─── Lookup helper ────────────────────────────────────────────────────────────

// LookupRegistryPatterns returns the effective (local+built-in merged) registry
// entry for toolKey, or (AssetRegistryEntry{}, false) if not registered.
//
// Matching is attempted with and without common backend prefixes.
func LookupRegistryPatterns(toolKey string) (AssetRegistryEntry, bool) {
	reg := EffectiveAssetRegistry()
	if e, ok := reg[toolKey]; ok {
		return e, true
	}
	for _, prefix := range []string{"github:", "asdf:", "aqua:", "native:"} {
		if len(toolKey) > len(prefix) && toolKey[:len(prefix)] == prefix {
			stripped := toolKey[len(prefix):]
			if e, ok := reg[stripped]; ok {
				return e, true
			}
		}
	}
	return AssetRegistryEntry{}, false
}
