// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginManager_LoadAll_NotExist(t *testing.T) {
	pm := NewPluginManager("/non/existent/dir/for/plugins", nil, nil)

	err := pm.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for non-existent directory, got %v", err)
	}

	if pm.PluginsDir() != "/non/existent/dir/for/plugins" {
		t.Errorf("expected plugins dir to be set")
	}

	if len(pm.ListLoaded()) != 0 {
		t.Errorf("expected 0 plugins loaded")
	}
}

func TestPluginManager_LoadAll_SkipsNonPlugins(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some non-plugin files and directories
	os.WriteFile(filepath.Join(tmpDir, "not-a-plugin.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "unirtm-plugin-"), []byte("test"), 0644) // Should match prefix but will fail to load, which is fine since loadPlugin logs and skips. Wait, it doesn't fail the whole LoadAll.
	os.Mkdir(filepath.Join(tmpDir, "unirtm-plugin-dir"), 0755)                  // directory, should be skipped

	pm := NewPluginManager(tmpDir, nil, nil)

	err := pm.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(pm.ListLoaded()) != 0 {
		t.Errorf("expected 0 plugins successfully loaded")
	}
}

func TestPluginManager_Cleanup(t *testing.T) {
	pm := NewPluginManager("test", nil, nil)
	// Add a fake client? We can't easily do this since hplugin.Client fields are private,
	// but we can call Cleanup to ensure it doesn't panic on empty list.
	pm.Cleanup()
	if pm.clients != nil {
		t.Error("expected clients to be nil after cleanup")
	}
}

func TestPluginManager_LoadPlugin_Fails(t *testing.T) {
	// Attempting to load a file that is not a valid plugin should return an error
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "unirtm-plugin-dummy")
	os.WriteFile(pluginPath, []byte("invalid binary"), 0755)

	pm := NewPluginManager(tmpDir, nil, nil)
	err := pm.loadPlugin(context.Background(), pluginPath)
	if err == nil {
		t.Fatal("expected error loading invalid plugin binary, got nil")
	}
}
