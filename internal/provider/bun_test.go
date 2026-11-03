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

func TestBunProvider_Interface(t *testing.T) {
	var p Provider = NewBunProvider()
	require.Equal(t, "bun", p.Name())
}

func TestBunProvider_GetBinPaths(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewBunProvider()

	// Test without executables
	paths, err := p.GetBinPaths("bun", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{tmpDir}, paths)

	// Create a fake executable
	bunPath := filepath.Join(tmpDir, "bin", "bun")
	os.MkdirAll(filepath.Dir(bunPath), 0755)
	os.WriteFile(bunPath, []byte("fake"), 0755)

	paths, err = p.GetBinPaths("bun", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Dir(bunPath)}, paths)
}

func TestBunProvider_Install(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewBunProvider()
	installPath := filepath.Join(tmpDir, "install")
	err := p.Install(context.Background(), "bun", installPath, "artifact", "1.0.0")
	require.NoError(t, err)
	require.DirExists(t, installPath)
}

func TestBunProvider_GenerateShims(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewBunProvider()

	bunPath := filepath.Join(tmpDir, "bun")
	os.WriteFile(bunPath, []byte("fake"), 0755)

	shims, err := p.GenerateShims("bun", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, 1, len(shims))
	require.Equal(t, bunPath, shims["bun"])
}

func TestBunProvider_ListExecutables(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewBunProvider()

	// Error path
	_, err := p.ListExecutables("bun", "/fake/nonexistent/path/that/fails/walk", "1.0.0")
	require.Error(t, err)

	// Success path
	bunPath := filepath.Join(tmpDir, "bun")
	os.WriteFile(bunPath, []byte("fake"), 0755)

	bunExePath := filepath.Join(tmpDir, "bun.exe")
	os.WriteFile(bunExePath, []byte("fake"), 0755)

	exes, err := p.ListExecutables("bun", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Len(t, exes, 2)
}

func TestBunProvider_DetectVersion(t *testing.T) {
	p := NewBunProvider()
	version, err := p.DetectVersion(context.Background(), "bun", "/fake/path/v1.0.0")
	require.NoError(t, err)
	require.Equal(t, "v1.0.0", version)
}
