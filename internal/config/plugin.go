// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

// PluginDeclaration represents an enabled plugin, its version, and its namespaced configuration.
type PluginDeclaration struct {
	// Version specifies the semantic version requirement (e.g. "1.0.0", "latest").
	Version string `toml:"version,omitempty" yaml:"version,omitempty" mapstructure:"version,omitempty" json:"version,omitempty"`

	// Config holds arbitrary namespaced settings owned exclusively by this plugin.
	Config map[string]interface{} `toml:"config,omitempty" yaml:"config,omitempty" mapstructure:"config,omitempty" json:"config,omitempty"`

	// Source indicates where the config originated: "inline", "sidecar", or "merged".
	Source string `toml:"-" yaml:"-" mapstructure:"-" json:"source,omitempty"`

	// SidecarPath tracks the path to the external file if loaded via sidecar.
	SidecarPath string `toml:"-" yaml:"-" mapstructure:"-" json:"sidecar_path,omitempty"`
}

// PluginMap maps a plugin identifier (e.g. "snowdreamtech/docker") to its declaration.
type PluginMap map[string]PluginDeclaration

// parsePluginDeclaration converts a generic interface{} (from unmarshaled TOML) into a PluginDeclaration.
func parsePluginDeclaration(v interface{}) PluginDeclaration {
	pd := PluginDeclaration{
		Config: make(map[string]interface{}),
		Source: "inline",
	}

	switch val := v.(type) {
	case string:
		pd.Version = val
	case map[string]interface{}:
		for k, item := range val {
			if k == "version" {
				if verStr, ok := item.(string); ok {
					pd.Version = verStr
				}
			} else if k == "config" {
				if cfgMap, ok := item.(map[string]interface{}); ok {
					for ck, cv := range cfgMap {
						pd.Config[ck] = cv
					}
				}
			} else {
				// Allow attributes inside plugin table to populate config
				pd.Config[k] = item
			}
		}
	}
	return pd
}
