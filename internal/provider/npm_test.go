package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNpmProvider_Install_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash-based mock test on windows")
	}
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Mock npm executable
	nodeInstallsDir := filepath.Join(tmpDir, "installs", "node", "18.0.0", "bin")
	err := os.MkdirAll(nodeInstallsDir, 0755)
	require.NoError(t, err)

	npmScript := filepath.Join(nodeInstallsDir, "npm")
	mockNpm := `#!/bin/sh
# Mock npm install
exit 0
`
	err = os.WriteFile(npmScript, []byte(mockNpm), 0755)
	require.NoError(t, err)

	p := provider.NewNpmProvider()
	installPath := filepath.Join(tmpDir, "npm_install", "test_pkg")

	err = p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.NoError(t, err)
}

func TestNpmProvider_Install_NpmNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	// Clear PATH
	t.Setenv("PATH", "")

	p := provider.NewNpmProvider()
	installPath := filepath.Join(tmpDir, "npm_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "npm is required")
}

func TestNpmProvider_ListExecutables(t *testing.T) {
	p := provider.NewNpmProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(binDir, "dummy1"), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, "dummy2"), []byte(""), 0644) // not executable

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	require.NoError(t, err)
	assert.Len(t, exes, 1)
	assert.Contains(t, exes, filepath.Join(binDir, "dummy1"))
}
