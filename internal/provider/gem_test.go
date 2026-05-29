// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGemProvider_Interface(t *testing.T) {
	var _ Provider = (*GemProvider)(nil)
}

func TestGemProvider_FindGem(t *testing.T) {
	p := NewGemProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	scriptPath := filepath.Join(binDir, "gem")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho gem"), 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	gem, err := p.findGem()
	if err != nil {
		t.Fatalf("expected to find gem, but got error: %v", err)
	}
	if gem != scriptPath {
		t.Errorf("expected %s, got %s", scriptPath, gem)
	}
}

func TestGemProvider_Install(t *testing.T) {
	p := NewGemProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	scriptPath := filepath.Join(binDir, "gem")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho installing..."), 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "tool", installPath, "", "1.0.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}

func TestGemProvider_findGem(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)

	p := NewGemProvider()

	// Create fake gem installation
	gemDir := filepath.Join(tmpData, "installs", "ruby", "3.2.2", "bin")
	os.MkdirAll(gemDir, 0755)
	gemPath := filepath.Join(gemDir, "gem")
	os.WriteFile(gemPath, []byte("fake binary"), 0755)

	found, err := p.findGem()
	require.NoError(t, err)
	require.Equal(t, gemPath, found)
}

func TestGemProvider_Install_GemNotFound(t *testing.T) {
	t.Setenv("PATH", "")

	p := NewGemProvider()
	installPath := filepath.Join(t.TempDir(), "gem_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "gem is required")
}

func TestGemProvider_ListExecutables(t *testing.T) {
	p := NewGemProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(binDir, "dummy1"), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, "dummy2"), []byte(""), 0644)

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Len(t, exes, 1)
	require.Contains(t, exes, filepath.Join(binDir, "dummy1"))
}
