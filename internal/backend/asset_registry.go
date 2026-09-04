// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

// AssetRegistryEntry describes known asset naming conventions for a specific tool.
type AssetRegistryEntry struct {
	// Patterns maps platform keys (e.g. "macos-amd64") to the exact asset filename.
	Patterns map[string]string

	// Description explains why the non-standard mapping is needed.
	Description string
}

// AssetRegistry is the built-in community-maintained registry of known asset
// naming quirks. The map key is the tool identifier as it appears in
// .unirtm.toml (e.g. "editorconfig-checker/editorconfig-checker").
//
// Entries here are consulted automatically as a fallback when:
//   - The heuristic FindBestAsset returns nil (no matching asset)
//   - AND no user-level asset_patterns override is configured for the tool
//
// User-level asset_patterns always take precedence over registry entries.
var AssetRegistry = map[string]AssetRegistryEntry{
	// editorconfig-checker publishes a single darwin-all archive that works
	// on both Intel (amd64) and Apple Silicon (arm64) macOS.
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

	// osv-scanner uses osv-scanner_<version>_<os>_<arch> naming.
	"google/osv-scanner": {
		Description: "Uses osv-scanner_<version>_<os>_<arch> naming convention",
		Patterns:    map[string]string{},
	},
}

// LookupRegistryPatterns returns the AssetRegistry entry for toolKey, or
// (AssetRegistryEntry{}, false) if the tool is not registered.
func LookupRegistryPatterns(toolKey string) (AssetRegistryEntry, bool) {
	if e, ok := AssetRegistry[toolKey]; ok {
		return e, true
	}
	for _, prefix := range []string{"github:", "asdf:", "aqua:", "native:"} {
		if len(toolKey) > len(prefix) && toolKey[:len(prefix)] == prefix {
			stripped := toolKey[len(prefix):]
			if e, ok := AssetRegistry[stripped]; ok {
				return e, true
			}
		}
	}
	return AssetRegistryEntry{}, false
}
