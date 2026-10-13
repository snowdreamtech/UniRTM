// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"regexp"
	"strings"
)

// ReservedCoreTables is the set of top-level table names permanently reserved for UniRTM core features.
// Third-party plugins and user extensions are strictly prohibited from creating new root-level tables.
var ReservedCoreTables = map[string]bool{
	"tools":        true,
	"tasks":        true,
	"format":       true,
	"lint":         true,
	"hooks":        true,
	"env":          true,
	"settings":     true,
	"environments": true,
	"aliases":      true,
	"plugins":      true,
}

// ReservedRootFields are scalar or list fields explicitly reserved at the root level of .unirtm.toml.
var ReservedRootFields = map[string]bool{
	"network_preset": true,
	"min_version":    true,
}

// pluginIDRegex validates plugin identifiers.
// Allows standard formats such as "docker", "vendor/name", "io.unirtm.name", or "github.com/org/plugin".
var pluginIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+(/[a-zA-Z0-9_\-\.]+)*$`)

// IsReservedRoot checks if a top-level key or table name is reserved by the UniRTM core.
func IsReservedRoot(rootName string) bool {
	return ReservedCoreTables[rootName] || ReservedRootFields[rootName]
}

// KnownSettingShorthands are common settings fields that can be specified without the 'settings.' prefix.
var KnownSettingShorthands = map[string]bool{
	"browser":   true,
	"cache_ttl": true,
	"data_dir":  true,
	"editor":    true,
	"shell":     true,
	"color":     true,
}

// ValidateKeyNamespace validates whether a configuration key path adheres to UniRTM's
// namespace rules.
func ValidateKeyNamespace(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("configuration key cannot be empty")
	}

	// Single key (no dot) validation
	if !strings.Contains(key, ".") {
		if key == "plugins" {
			return fmt.Errorf("plugin configuration requires a plugin ID (e.g. 'plugins.\"<plugin-id>\".config.<key>')")
		}
		if IsReservedRoot(key) || KnownSettingShorthands[key] {
			return nil
		}
		return fmt.Errorf("root key %q is not a reserved core table or valid plugin namespace; plugins must use 'plugins.\"<id>\".config.<key>'", key)
	}

	parts := strings.Split(key, ".")
	root := strings.Trim(parts[0], `"'`)

	if root == "plugins" {
		if len(parts) < 2 {
			return fmt.Errorf("plugin configuration requires a plugin ID (e.g. 'plugins.\"<plugin-id>\".config.<key>')")
		}
		// Extract plugin ID (might be quoted, e.g. plugins."vendor/name".config)
		rawPluginID := parts[1]
		pluginID := strings.Trim(rawPluginID, `"'`)
		if pluginID == "" {
			return fmt.Errorf("plugin ID cannot be empty in key %q", key)
		}
		if !pluginIDRegex.MatchString(pluginID) {
			return fmt.Errorf("invalid plugin ID %q: must contain only alphanumeric characters, dots, hyphens, and slashes", pluginID)
		}
		return nil
	}

	if IsReservedRoot(root) {
		return nil
	}

	return fmt.Errorf("root namespace %q is not reserved for core; third-party plugins must use 'plugins.\"<id>\".config.<key>'", root)
}
