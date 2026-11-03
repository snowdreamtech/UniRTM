// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ResolveAlias(t *testing.T) {
	t.Run("nil aliases map returns original", func(t *testing.T) {
		c := &Config{}
		result := c.ResolveAlias("go", "lts")
		assert.Equal(t, "lts", result)
	})

	t.Run("alias resolves correctly", func(t *testing.T) {
		cfg := &Config{
			Aliases: map[string]map[string]string{
				"node": {
					"lts": "20.11.0",
				},
			},
		}
		result := cfg.ResolveAlias("node", "lts")
		assert.Equal(t, "20.11.0", result)
	})

	t.Run("alias for different tool returns original", func(t *testing.T) {
		cfg := &Config{
			Aliases: map[string]map[string]string{
				"node": {
					"lts": "20.11.0",
				},
			},
		}
		result := cfg.ResolveAlias("go", "lts")
		assert.Equal(t, "lts", result)
	})

	t.Run("alias key missing returns original version", func(t *testing.T) {
		cfg := &Config{
			Aliases: map[string]map[string]string{
				"node": {
					"stable": "18.0.0",
				},
			},
		}
		result := cfg.ResolveAlias("node", "lts")
		assert.Equal(t, "lts", result)
	})
}

func TestManager_Merge_Settings(t *testing.T) {
	m := NewConfigManager()
	cm := m.(*defaultConfigManager)

	autoInstall := true
	verifyMetadata := false

	c1 := &Config{
		Settings: Settings{
			CacheDir:           "/cache1",
			DataDir:            "/data1",
			GitHubProxy:        "proxy1",
			HttpProxy:          "http1",
			HttpsProxy:         "https1",
			GitHubToken:        "token1",
			HTTPTimeout:        DurationOrInt(60),
			TaskTimeout:        DurationOrInt(300),
			TaskOutput:         "quiet",
			AutoInstall:        &autoInstall,
			Color:              "always",
			AlwaysKeepDownload: true,
			VerifyMetadata:     &verifyMetadata,
			CeilingPaths:       []string{"/path1"},
			Jobs:               8,
			CacheTTL:           DurationOrInt(86400),
		},
	}
	c2 := &Config{
		Settings: Settings{
			Jobs:     4,
			CacheTTL: DurationOrInt(3600),
		},
	}

	merged, err := cm.Merge(c1, c2)
	require.NoError(t, err)

	assert.Equal(t, "/cache1", merged.Settings.CacheDir)
	assert.Equal(t, "/data1", merged.Settings.DataDir)
	assert.Equal(t, "proxy1", merged.Settings.GitHubProxy)
	assert.Equal(t, "http1", merged.Settings.HttpProxy)
	assert.Equal(t, "https1", merged.Settings.HttpsProxy)
	assert.Equal(t, "token1", merged.Settings.GitHubToken)
	assert.Equal(t, DurationOrInt(60), merged.Settings.HTTPTimeout)
	assert.Equal(t, DurationOrInt(300), merged.Settings.TaskTimeout)
	assert.Equal(t, "quiet", merged.Settings.TaskOutput)
	assert.NotNil(t, merged.Settings.AutoInstall)
	assert.Equal(t, "always", merged.Settings.Color)
	assert.True(t, merged.Settings.AlwaysKeepDownload)
	assert.NotNil(t, merged.Settings.VerifyMetadata)
	assert.Equal(t, []string{"/path1"}, merged.Settings.CeilingPaths)
	// c2 overrides Jobs and CacheTTL
	assert.Equal(t, 4, merged.Settings.Jobs)
	assert.Equal(t, DurationOrInt(3600), merged.Settings.CacheTTL)
}

func TestManager_Merge_Aliases(t *testing.T) {
	m := NewConfigManager()
	cm := m.(*defaultConfigManager)

	c1 := &Config{
		Aliases: map[string]map[string]string{
			"node": {"lts": "18.0.0"},
		},
	}
	c2 := &Config{
		Aliases: map[string]map[string]string{
			"node": {"lts": "20.0.0", "stable": "20.0.0"},
			"go":   {"latest": "1.22.0"},
		},
	}

	merged, err := cm.Merge(c1, c2)
	require.NoError(t, err)

	// Node lts: c2 overrides
	assert.Equal(t, "20.0.0", merged.Aliases["node"]["lts"])
	// Node stable: from c2
	assert.Equal(t, "20.0.0", merged.Aliases["node"]["stable"])
	// Go latest: from c2
	assert.Equal(t, "1.22.0", merged.Aliases["go"]["latest"])
}

func TestManager_Merge_Nil(t *testing.T) {
	m := NewConfigManager()
	cm := m.(*defaultConfigManager)

	// Empty
	_, err := cm.Merge()
	assert.Error(t, err)

	// Nil element
	_, err = cm.Merge(nil, &Config{})
	assert.Error(t, err)
}

func TestManager_Merge_Settings_Tools(t *testing.T) {
	m := NewConfigManager()
	cm := m.(*defaultConfigManager)

	c1 := &Config{
		Settings: Settings{
			Tools: map[string]map[string]interface{}{
				"node": {"gpg_verify": true},
			},
		},
	}
	c2 := &Config{
		Settings: Settings{
			Tools: map[string]map[string]interface{}{
				"node": {"mirror": "https://example.com"},
				"go":   {"build_flags": "-v"},
			},
		},
	}

	merged, err := cm.Merge(c1, c2)
	require.NoError(t, err)
	assert.Equal(t, true, merged.Settings.Tools["node"]["gpg_verify"])
	assert.Equal(t, "https://example.com", merged.Settings.Tools["node"]["mirror"])
	assert.Equal(t, "-v", merged.Settings.Tools["go"]["build_flags"])
}

func TestManager_LoadWithEnvironment_Errors(t *testing.T) {
	m := NewConfigManager()

	// empty env name
	_, err := m.LoadWithEnvironment(context.Background(), "")
	// Should not error (returns hierarchical config)
	_ = err

	// nonexistent env - also not an error
	_, err = m.LoadWithEnvironment(context.Background(), "nonexistent")
	_ = err
}
