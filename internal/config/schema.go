// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// PluginManifest defines the extension metadata and contributed configuration schema.
type PluginManifest struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version" yaml:"version"`
	Description string            `json:"description" yaml:"description"`
	Contributes PluginContributes `json:"contributes" yaml:"contributes"`
}

// PluginContributes holds contributed configurations and extensions.
type PluginContributes struct {
	Configuration map[string]ConfigPropertySchema `json:"configuration" yaml:"configuration"`
}

// ConfigPropertySchema declares types, defaults, descriptions, and enums for a property.
type ConfigPropertySchema struct {
	Type        string        `json:"type" yaml:"type"` // "string", "integer", "number", "boolean", "array", "object"
	Default     interface{}   `json:"default,omitempty" yaml:"default,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
}

// ValidateConfigValue validates a string input value against a property's ConfigPropertySchema.
func ValidateConfigValue(prop string, val string, schema ConfigPropertySchema) error {
	// 1. Type validation
	switch strings.ToLower(schema.Type) {
	case "string":
		// string values are always valid strings
	case "integer", "int":
		if _, err := strconv.ParseInt(val, 10, 64); err != nil {
			return fmt.Errorf("property %q must be an integer, got %q", prop, val)
		}
	case "number", "float":
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return fmt.Errorf("property %q must be a number, got %q", prop, val)
		}
	case "boolean", "bool":
		low := strings.ToLower(val)
		if low != "true" && low != "false" && low != "1" && low != "0" {
			return fmt.Errorf("property %q must be a boolean (true/false), got %q", prop, val)
		}
	case "array", "list":
		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			var arr []interface{}
			if err := json.Unmarshal([]byte(val), &arr); err != nil {
				return fmt.Errorf("property %q must be a valid list/array, got %q", prop, val)
			}
		}
	}

	// 2. Enum validation
	if len(schema.Enum) > 0 {
		match := false
		for _, allowed := range schema.Enum {
			if fmt.Sprintf("%v", allowed) == val {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("property %q value %q is not in allowed enums: %v", prop, val, schema.Enum)
		}
	}

	return nil
}

// FindPluginManifest looks for manifest files in project and global plugins dirs.
func FindPluginManifest(pluginID string, projectDir string) (*PluginManifest, error) {
	slug := strings.ReplaceAll(pluginID, "/", "-")
	candidates := []string{
		filepath.Join(projectDir, ".unirtm", "plugins", pluginID+".manifest.json"),
		filepath.Join(projectDir, ".unirtm", "plugins", slug+".manifest.json"),
		filepath.Join(projectDir, ".unirtm", "plugins", pluginID, "manifest.json"),
		filepath.Join(env.GetConfigDir(), "plugins", pluginID+".manifest.json"),
		filepath.Join(env.GetConfigDir(), "plugins", slug+".manifest.json"),
	}

	for _, c := range candidates {
		data, err := os.ReadFile(c)
		if err == nil {
			var m PluginManifest
			if err := json.Unmarshal(data, &m); err == nil {
				return &m, nil
			}
		}
	}
	return nil, nil
}
