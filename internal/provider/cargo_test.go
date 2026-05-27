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

func TestCargoProvider_Interface(t *testing.T) {
	var _ Provider = (*CargoProvider)(nil)
}

func TestCargoProvider_FindCargo(t *testing.T) {
	p := NewCargoProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	scriptPath := filepath.Join(binDir, "cargo")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho cargo"), 0755)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	cargo, err := p.findCargo()
	if err != nil {
		t.Fatalf("expected to find cargo, but got error: %v", err)
	}
	if cargo != scriptPath {
		t.Errorf("expected %s, got %s", scriptPath, cargo)
	}
}

func TestCargoProvider_Install(t *testing.T) {
	p := NewCargoProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	scriptPath := filepath.Join(binDir, "cargo")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho installing..."), 0755)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "cargo-binstall", installPath, "", "1.0.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}

func TestCargoProvider_findCargo(t *testing.T) {
	tmpData := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpData)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	
	p := NewCargoProvider()
	
	// Create fake rust installation
	rustDir := filepath.Join(tmpData, "installs", "rust", "1.70.0", "bin")
	os.MkdirAll(rustDir, 0755)
	cargoPath := filepath.Join(rustDir, "cargo")
	os.WriteFile(cargoPath, []byte("fake binary"), 0755)
	
	found, err := p.findCargo()
	require.NoError(t, err)
	require.Equal(t, cargoPath, found)
}

func TestCargoProvider_Install_CargoNotFound(t *testing.T) {
	t.Setenv("PATH", "")

	p := NewCargoProvider()
	installPath := filepath.Join(t.TempDir(), "cargo_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cargo is required")
}

func TestCargoProvider_ListExecutables(t *testing.T) {
	p := NewCargoProvider()
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
