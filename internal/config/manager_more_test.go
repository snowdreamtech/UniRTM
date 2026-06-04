// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigManager_ApplyEnvironment_Error(t *testing.T) {
	cm := NewConfigManager()
	
	// Nil config
	_, err := cm.ApplyEnvironment(nil, "dev")
	require.Error(t, err)

	// Empty env
	cfg := &Config{}
	_, err = cm.ApplyEnvironment(cfg, "")
	require.Error(t, err)

	// Env not found
	_, err = cm.ApplyEnvironment(cfg, "prod")
	require.Error(t, err)
}

func TestConfigManager_LoadHierarchy_LoadError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir, _ = filepath.EvalSymlinks(tmpDir)
	invalidToml := filepath.Join(tmpDir, "unirtm.toml")
	err := os.WriteFile(invalidToml, []byte("invalid = = toml"), 0644)
	require.NoError(t, err)

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cm := NewConfigManager()
	// Trust it first!
	err = cm.(*defaultConfigManager).trustManager.Trust(invalidToml)
	require.NoError(t, err)
	
	ctx := context.Background()
	_, err = cm.LoadHierarchy(ctx)
	// It should fail because the unirtm.toml is invalid
	require.Error(t, err)
}

func TestConfigManager_tryLoad(t *testing.T) {
	cm := NewConfigManager().(*defaultConfigManager)
	
	// Test file that does not exist
	cfg, err := cm.tryLoad(context.Background(), "/does/not/exist.toml", false, nil)
	require.NoError(t, err)
	require.Nil(t, cfg)
}

func TestConfigManager_Merge_Errors(t *testing.T) {
	cm := NewConfigManager()
	
	_, err := cm.Merge()
	require.Error(t, err)

	_, err = cm.Merge(&Config{}, nil)
	require.Error(t, err)
}

func TestConfigManager_Validate_Error(t *testing.T) {
	cm := NewConfigManager()
	err := cm.Validate(context.Background(), nil)
	require.Error(t, err)
}

func TestConfigManager_LoadWithEnvironment_Error(t *testing.T) {
	cm := NewConfigManager()
	_, err := cm.LoadWithEnvironment(context.Background(), "missing_env")
	// Since LoadHierarchy doesn't fail, but ApplyEnvironment fails
	require.Error(t, err)
}

func TestConfigManager_ApplyEnvironment_MergeSettings(t *testing.T) {
	cm := NewConfigManager()
	base := &Config{
		Environments: map[string]EnvironmentConfig{
			"dev": {
				Settings: Settings{
					CacheDir: "/tmp/cache",
					DataDir: "/tmp/data",
					CacheTTL: 3600,
					Jobs: 4,
					GitHubProxy: "proxy",
					HttpProxy: "http",
					HttpsProxy: "https",
					GitHubToken: "token",
					HTTPTimeout: 10,
					TaskTimeout: 10,
					TaskOutput: "interleaved",
					AutoInstall: func() *bool { b := true; return &b }(),
					Color: "always",
					AlwaysKeepDownload: true,
					VerifyMetadata: func() *bool { b := true; return &b }(),
				},
				Tools: map[string]ToolConfig{"node": {Version: "18"}},
				Env: map[string]interface{}{"FOO": "bar"},
				Tasks: map[string]Task{"build": {Run: StringArray{"make"}}},
			},
		},
		Aliases: map[string]map[string]string{
			"node": {"lts": "20"},
		},
	}
	
	envCfg, err := cm.ApplyEnvironment(base, "dev")
	require.NoError(t, err)
	require.Equal(t, "/tmp/cache", envCfg.Settings.CacheDir)
	require.Equal(t, "/tmp/data", envCfg.Settings.DataDir)
	require.Equal(t, "token", envCfg.Settings.GitHubToken)
}

func TestResolveAlias(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]map[string]string{
			"node": {"lts": "20.x"},
		},
	}
	require.Equal(t, "20.x", cfg.ResolveAlias("node", "lts"))
	require.Equal(t, "18.x", cfg.ResolveAlias("node", "18.x"))
	require.Equal(t, "1.x", cfg.ResolveAlias("go", "1.x"))

	cfgNil := &Config{}
	require.Equal(t, "1.x", cfgNil.ResolveAlias("node", "1.x"))
}

func TestRenderTemplate_Error(t *testing.T) {
	// Test error parsing template
	res := renderTemplate("{{ foo", nil)
	require.Equal(t, "{{ foo", res)

	// Test bridgeJinja2
	res2 := bridgeJinja2("foo is defined")
	require.Equal(t, "foo", res2)
}

func TestConfigManager_tryLoad_Branches(t *testing.T) {
	tmpDir, _ := filepath.EvalSymlinks(t.TempDir())
	path := filepath.Join(tmpDir, "unirtm.toml")
	err := os.WriteFile(path, []byte("min_version = '1.0.0'"), 0644)
	require.NoError(t, err)

	cm := NewConfigManager().(*defaultConfigManager)

	// 1. enforceTrust = false
	cfg, err := cm.tryLoad(context.Background(), path, false, nil)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 2. globally trusted
	settings := &Settings{
		TrustedConfigPaths: []string{path},
	}
	cfg, err = cm.tryLoad(context.Background(), path, true, settings)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 3. Untrusted (not globally trusted, not in trustManager)
	cfg, err = cm.tryLoad(context.Background(), path, true, &Settings{})
	require.NoError(t, err)
	require.Nil(t, cfg) // returns nil, nil for untrusted

	// 4. Trusted
	err = cm.trustManager.Trust(path)
	require.NoError(t, err)
	cfg, err = cm.tryLoad(context.Background(), path, true, &Settings{})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 5. Modified
	err = os.WriteFile(path, []byte("[env]\nSECRET='123'\n[tasks.build]\nrun=['make']"), 0644)
	require.NoError(t, err)
	cfg, err = cm.tryLoad(context.Background(), path, true, &Settings{})
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Empty(t, cfg.Env) // Should be stripped
	require.Empty(t, cfg.Tasks) // Should be stripped
}
