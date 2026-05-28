// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVfoxProvider_Interface(t *testing.T) {
	var p Provider = NewVfoxProvider()
	require.Equal(t, "vfox", p.Name())
}

func TestVfoxProvider_GetBinPaths(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewVfoxProvider()

	// Test without executables
	paths, err := p.GetBinPaths("vfox", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(tmpDir, "bin")}, paths)
}

func TestVfoxProvider_GenerateShims(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewVfoxProvider()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	vfoxPath := filepath.Join(binDir, "vfox")
	os.WriteFile(vfoxPath, []byte("fake"), 0755)

	os.Chmod(vfoxPath, 0755)

	shims, err := p.GenerateShims("vfox", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, 1, len(shims))
	require.Equal(t, vfoxPath, shims["vfox"])
}
