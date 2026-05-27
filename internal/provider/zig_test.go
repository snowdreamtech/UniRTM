package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZigProvider_Interface(t *testing.T) {
	var p Provider = NewZigProvider()
	require.Equal(t, "zig", p.Name())
}

func TestZigProvider_GetBinPaths(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewZigProvider()
	
	// Test without executables
	paths, err := p.GetBinPaths("zig", tmpDir, "0.11.0")
	require.NoError(t, err)
	require.Equal(t, []string{tmpDir}, paths)
	
	// Create a fake executable
	zigPath := filepath.Join(tmpDir, "bin", "zig")
	os.MkdirAll(filepath.Dir(zigPath), 0755)
	os.WriteFile(zigPath, []byte("fake"), 0755)
	
	paths, err = p.GetBinPaths("zig", tmpDir, "0.11.0")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Dir(zigPath)}, paths)
}

func TestZigProvider_Install(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewZigProvider()
	installPath := filepath.Join(tmpDir, "install")
	err := p.Install(context.Background(), "zig", installPath, "artifact", "0.11.0")
	require.NoError(t, err)
	require.DirExists(t, installPath)
}

func TestZigProvider_GenerateShims(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewZigProvider()
	
	zigPath := filepath.Join(tmpDir, "zig")
	os.WriteFile(zigPath, []byte("fake"), 0755)
	
	shims, err := p.GenerateShims("zig", tmpDir, "0.11.0")
	require.NoError(t, err)
	require.Equal(t, 1, len(shims))
	require.Equal(t, zigPath, shims["zig"])
}
