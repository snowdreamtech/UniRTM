// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigValue_Types(t *testing.T) {
	// 1. Integer
	intSchema := ConfigPropertySchema{Type: "integer"}
	assert.NoError(t, ValidateConfigValue("port", "8080", intSchema))
	assert.Error(t, ValidateConfigValue("port", "abc", intSchema))

	// 2. Boolean
	boolSchema := ConfigPropertySchema{Type: "boolean"}
	assert.NoError(t, ValidateConfigValue("enabled", "true", boolSchema))
	assert.NoError(t, ValidateConfigValue("enabled", "false", boolSchema))
	assert.NoError(t, ValidateConfigValue("enabled", "1", boolSchema))
	assert.Error(t, ValidateConfigValue("enabled", "yes_maybe", boolSchema))

	// 3. Float / Number
	numSchema := ConfigPropertySchema{Type: "number"}
	assert.NoError(t, ValidateConfigValue("factor", "3.14", numSchema))
	assert.NoError(t, ValidateConfigValue("factor", "42", numSchema))
	assert.Error(t, ValidateConfigValue("factor", "nan_invalid", numSchema))

	// 4. Array / List
	arrSchema := ConfigPropertySchema{Type: "array"}
	assert.NoError(t, ValidateConfigValue("tags", `["v1", "v2"]`, arrSchema))
	assert.Error(t, ValidateConfigValue("tags", `[unclosed`, arrSchema))
}

func TestValidateConfigValue_Enums(t *testing.T) {
	enumSchema := ConfigPropertySchema{
		Type: "string",
		Enum: []interface{}{"debug", "info", "warn", "error"},
	}

	assert.NoError(t, ValidateConfigValue("log_level", "info", enumSchema))
	assert.NoError(t, ValidateConfigValue("log_level", "debug", enumSchema))

	err := ValidateConfigValue("log_level", "trace", enumSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowed enums")
}

func TestFindPluginManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0755))

	manifestContent := `{
  "name": "docker",
  "version": "1.0.0",
  "description": "Docker container runtime support",
  "contributes": {
    "configuration": {
      "port": {
        "type": "integer",
        "description": "Port to bind container to",
        "default": 8080
      },
      "driver": {
        "type": "string",
        "enum": ["overlay2", "btrfs", "vfs"]
      }
    }
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "docker.manifest.json"), []byte(manifestContent), 0644))

	m, err := FindPluginManifest("docker", tmpDir)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "docker", m.Name)
	assert.Equal(t, "1.0.0", m.Version)

	portSchema, ok := m.Contributes.Configuration["port"]
	require.True(t, ok)
	assert.Equal(t, "integer", portSchema.Type)

	driverSchema, ok := m.Contributes.Configuration["driver"]
	require.True(t, ok)
	assert.Len(t, driverSchema.Enum, 3)
}
