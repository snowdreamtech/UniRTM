package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUbiProvider_Interface(t *testing.T) {
	var p Provider = NewUbiProvider()
	require.Equal(t, "ubi", p.Name())
}

func TestUbiProvider_GetBinPaths(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewUbiProvider()
	
	// Test without executables
	paths, err := p.GetBinPaths("ubi", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(tmpDir, "bin")}, paths)
}

func TestUbiProvider_GenerateShims(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewUbiProvider()
	
	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	ubiPath := filepath.Join(binDir, "ubi")
	os.WriteFile(ubiPath, []byte("fake"), 0755)
	
	// Note: ListExecutables looks for executables
	os.Chmod(ubiPath, 0755)
	
	shims, err := p.GenerateShims("ubi", tmpDir, "1.0.0")
	require.NoError(t, err)
	require.Equal(t, 1, len(shims))
	require.Equal(t, ubiPath, shims["ubi"])
}
