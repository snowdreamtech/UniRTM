package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDenoProvider_Interface(t *testing.T) {
	var p Provider = NewDenoProvider()
	require.Equal(t, "deno", p.Name())
}

func TestDenoProvider_GetBinPaths(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewDenoProvider()

	// Test without executables
	paths, err := p.GetBinPaths("deno", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{tmpDir}, paths)

	// Create a fake executable
	denoPath := filepath.Join(tmpDir, "bin", "deno")
	os.MkdirAll(filepath.Dir(denoPath), 0755)
	os.WriteFile(denoPath, []byte("fake"), 0755)

	paths, err = p.GetBinPaths("deno", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Dir(denoPath)}, paths)
}

func TestDenoProvider_Install(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewDenoProvider()
	installPath := filepath.Join(tmpDir, "install")
	err := p.Install(context.Background(), "deno", installPath, "artifact", "1.0.0")
	require.NoError(t, err)
	require.DirExists(t, installPath)
}

func TestDenoProvider_GenerateShims(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewDenoProvider()

	denoPath := filepath.Join(tmpDir, "deno")
	os.WriteFile(denoPath, []byte("fake"), 0755)

	shims, err := p.GenerateShims("deno", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, 1, len(shims))
	require.Equal(t, denoPath, shims["deno"])
}
