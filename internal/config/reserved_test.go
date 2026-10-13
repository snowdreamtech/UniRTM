// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReservedCoreTables(t *testing.T) {
	// Verify that all fundamental tables are permanently reserved
	expectedTables := []string{
		"tools", "tasks", "format", "lint", "hooks", "env", "settings", "environments", "aliases", "plugins",
	}

	for _, table := range expectedTables {
		assert.True(t, ReservedCoreTables[table], "Table %q must be in ReservedCoreTables", table)
		assert.True(t, IsReservedRoot(table), "IsReservedRoot(%q) must be true", table)
	}

	// Verify that arbitrary third-party plugin names are NOT reserved core tables
	assert.False(t, IsReservedRoot("docker"))
	assert.False(t, IsReservedRoot("my_plugin"))
	assert.False(t, IsReservedRoot("custom_deploy"))
}

func TestValidateKeyNamespace_CoreKeys(t *testing.T) {
	validCoreKeys := []string{
		"tools.go",
		"tools.node.version",
		"settings.cache_ttl",
		"settings.browser",
		"env.NODE_ENV",
		"tasks.dev",
		"format.indent",
		"format.overrides",
		"lint.tools",
		"hooks.pre-commit",
		"environments.production.env.PORT",
		"network_preset",
		"browser",
	}

	for _, key := range validCoreKeys {
		err := ValidateKeyNamespace(key)
		assert.NoError(t, err, "Key %q should be valid", key)
	}
}

func TestValidateKeyNamespace_PluginKeys(t *testing.T) {
	validPluginKeys := []string{
		`plugins.docker.config.port`,
		`plugins."snowdreamtech/docker".config.compose_file`,
		`plugins."community/deploy".config.cluster`,
		`plugins.io.unirtm.rust.config.toolchain`,
		`plugins.my-plugin.version`,
	}

	for _, key := range validPluginKeys {
		err := ValidateKeyNamespace(key)
		assert.NoError(t, err, "Plugin key %q should be valid", key)
	}
}

func TestValidateKeyNamespace_IllegalKeys(t *testing.T) {
	illegalKeys := []struct {
		key         string
		errContains string
	}{
		{"", "cannot be empty"},
		{"docker.port", "not reserved for core"},
		{"myplugin.setting", "not reserved for core"},
		{"random_root", "not a reserved core table"},
		{"plugins.", "plugin ID cannot be empty"},
		{"plugins", "requires a plugin ID"},
		{`plugins."invalid id with spaces".config`, "invalid plugin ID"},
	}

	for _, tc := range illegalKeys {
		err := ValidateKeyNamespace(tc.key)
		if assert.Error(t, err, "Key %q must fail validation", tc.key) {
			assert.Contains(t, err.Error(), tc.errContains, "Error for %q should contain %q", tc.key, tc.errContains)
		}
	}
}

func TestPluginDeclarationParsing(t *testing.T) {
	// 1. String version declaration
	pd1 := parsePluginDeclaration("1.2.0")
	assert.Equal(t, "1.2.0", pd1.Version)
	assert.Equal(t, "inline", pd1.Source)
	assert.Empty(t, pd1.Config)

	// 2. Map with version and config table
	rawMap := map[string]interface{}{
		"version": "2.0.0",
		"config": map[string]interface{}{
			"port":    8080,
			"cluster": "prod",
		},
	}
	pd2 := parsePluginDeclaration(rawMap)
	assert.Equal(t, "2.0.0", pd2.Version)
	assert.Equal(t, 8080, pd2.Config["port"])
	assert.Equal(t, "prod", pd2.Config["cluster"])

	// 3. PostLoad integration
	cfg := &Config{
		PluginsRaw: map[string]interface{}{
			"simple": "0.1.0",
			"complex": map[string]interface{}{
				"version": "1.0.0",
				"config": map[string]interface{}{
					"enabled": true,
				},
			},
		},
	}
	cfg.PostLoad()
	assert.NotNil(t, cfg.Plugins)
	assert.Equal(t, "0.1.0", cfg.Plugins["simple"].Version)
	assert.Equal(t, "1.0.0", cfg.Plugins["complex"].Version)
	assert.Equal(t, true, cfg.Plugins["complex"].Config["enabled"])
}

func TestSplitConfigKey(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"settings.cache_ttl", []string{"settings", "cache_ttl"}},
		{"plugins.docker.config.port", []string{"plugins", "docker", "config", "port"}},
		{`plugins."snowdreamtech/docker".config.compose_file`, []string{"plugins", "snowdreamtech/docker", "config", "compose_file"}},
		{`plugins."io.unirtm.rust".config.toolchain`, []string{"plugins", "io.unirtm.rust", "config", "toolchain"}},
	}
	for _, tc := range cases {
		actual := SplitConfigKey(tc.input)
		assert.Equal(t, tc.expected, actual, "Input %q should split properly", tc.input)
	}
}
