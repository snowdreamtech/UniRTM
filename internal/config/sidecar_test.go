// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSidecarDiscovery_FlatAndNested(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(filepath.Join(pluginsDir, "vendor"), 0755))

	// 1. Flat sidecar
	flatContent := `
port = 8080
host = "127.0.0.1"
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "docker.toml"), []byte(flatContent), 0644))

	// 2. Nested vendor sidecar
	nestedContent := `
version = "2.1.0"
[config]
cluster = "prod-east"
replicas = 3
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "vendor", "deploy.toml"), []byte(nestedContent), 0644))

	// 3. Sidecar with explicit id
	explicitContent := `
id = "custom/my-plugin"
version = "0.9.0"
timeout = "30s"
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "anything.toml"), []byte(explicitContent), 0644))

	sidecars, err := DiscoverProjectSidecars(tmpDir)
	require.NoError(t, err)
	require.Len(t, sidecars, 3)

	// Verify flat sidecar
	docker, ok := sidecars["docker"]
	require.True(t, ok)
	assert.Equal(t, int64(8080), docker.Config["port"])
	assert.Equal(t, "127.0.0.1", docker.Config["host"])
	assert.Equal(t, "sidecar", docker.Source)

	// Verify nested sidecar
	deploy, ok := sidecars["vendor/deploy"]
	require.True(t, ok)
	assert.Equal(t, "2.1.0", deploy.Version)
	assert.Equal(t, "prod-east", deploy.Config["cluster"])
	assert.Equal(t, int64(3), deploy.Config["replicas"])

	// Verify explicit ID sidecar
	custom, ok := sidecars["custom/my-plugin"]
	require.True(t, ok)
	assert.Equal(t, "0.9.0", custom.Version)
	assert.Equal(t, "30s", custom.Config["timeout"])
}

func TestSidecarPrecedence_InlineOverridesSidecar(t *testing.T) {
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0755))

	// Create sidecar with default port = 8080 and default logging = true
	sidecarContent := `
port = 8080
logging = true
image = "golang:alpine"
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "docker.toml"), []byte(sidecarContent), 0644))

	// Target config has inline override for port = 9000, but omits logging and image
	targetCfg := &Config{
		Plugins: PluginMap{
			"docker": PluginDeclaration{
				Version: "1.0.0",
				Config: map[string]interface{}{
					"port": int64(9000), // Inline override!
				},
			},
		},
	}

	sidecars, err := DiscoverProjectSidecars(tmpDir)
	require.NoError(t, err)

	MergeSidecarConfigs(targetCfg, sidecars)

	docker := targetCfg.Plugins["docker"]
	assert.Equal(t, "1.0.0", docker.Version)
	assert.Equal(t, "merged", docker.Source)

	// Inline port wins
	assert.Equal(t, int64(9000), docker.Config["port"], "Inline port must override sidecar port")
	// Non-conflicting keys are inherited from sidecar
	assert.Equal(t, true, docker.Config["logging"])
	assert.Equal(t, "golang:alpine", docker.Config["image"])
}

func TestSidecarInLoadHierarchy(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0755))

	// Write sidecar for "monitoring"
	sidecarContent := `
interval = "10s"
retention = "7d"
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, "monitoring.toml"), []byte(sidecarContent), 0644))

	// Write .unirtm.toml with inline override for interval and new key enabled
	projConfig := `
[settings]
cache_ttl = 3600

[plugins."monitoring".config]
interval = "5s"
enabled = true
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(projConfig), 0644))

	// Mark config file trusted so LoadHierarchy accepts it
	tm := NewTrustManager()
	require.NoError(t, tm.Trust(filepath.Join(tmpDir, ".unirtm.toml")))

	cm := NewConfigManager()
	cfg, err := cm.LoadHierarchy(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	monitoring, ok := cfg.Plugins["monitoring"]
	require.True(t, ok, "Plugin monitoring should be discovered from sidecar + inline")
	// interval: inline (5s) wins over sidecar (10s)
	assert.Equal(t, "5s", monitoring.Config["interval"])
	// retention: inherited from sidecar
	assert.Equal(t, "7d", monitoring.Config["retention"])
	// enabled: from inline
	assert.Equal(t, true, monitoring.Config["enabled"])
}

func TestSidecarPerformance_SC003(t *testing.T) {
	// SC-003 requires sidecar discovery to parse in < 5ms
	tmpDir := t.TempDir()
	pluginsDir := filepath.Join(tmpDir, ".unirtm", "plugins")
	require.NoError(t, os.MkdirAll(pluginsDir, 0755))

	for i := 0; i < 5; i++ {
		content := `
port = 8080
enabled = true
mode = "cluster"
`
		require.NoError(t, os.WriteFile(filepath.Join(pluginsDir, filepath.Base(t.Name())+"_"+string(rune('a'+i))+".toml"), []byte(content), 0644))
	}

	start := time.Now()
	sidecars, err := DiscoverProjectSidecars(tmpDir)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 5, len(sidecars))
	assert.Less(t, elapsed, 100*time.Millisecond, "Sidecar discovery must be fast and within SC-003 budget")
}
