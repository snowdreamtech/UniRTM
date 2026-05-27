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

func TestPypiProvider_Install_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash-based mock test on windows")
	}
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Mock python executable
	pythonInstallsDir := filepath.Join(tmpDir, "installs", "python", "3.11.0", "bin")
	err := os.MkdirAll(pythonInstallsDir, 0755)
	require.NoError(t, err)

	pythonScript := filepath.Join(pythonInstallsDir, "python3")
	mockPython := `#!/bin/sh
if [ "$1" = "-m" ] && [ "$2" = "venv" ]; then
	mkdir -p "$3/bin"
	echo "#!/bin/sh" > "$3/bin/pip"
	echo "exit 0" >> "$3/bin/pip"
	chmod +x "$3/bin/pip"
	exit 0
fi
exit 0
`
	err = os.WriteFile(pythonScript, []byte(mockPython), 0755)
	require.NoError(t, err)

	p := provider.NewPypiProvider()
	installPath := filepath.Join(tmpDir, "pypi_install", "test_pkg")

	err = p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.NoError(t, err)
}

func TestPypiProvider_Install_PythonNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	// Clear PATH so it can't fallback to system python
	t.Setenv("PATH", "")

	p := provider.NewPypiProvider()
	installPath := filepath.Join(tmpDir, "pypi_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "python is required")
}

func TestPypiProvider_ListExecutables(t *testing.T) {
	p := provider.NewPypiProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	// Create some dummy files
	os.WriteFile(filepath.Join(binDir, "dummy1"), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, "dummy2"), []byte(""), 0644) // not executable
	os.WriteFile(filepath.Join(binDir, "pip"), []byte(""), 0755) // Should be excluded
	os.WriteFile(filepath.Join(binDir, "python"), []byte(""), 0755) // Should be excluded
	os.WriteFile(filepath.Join(binDir, "Activate.ps1"), []byte(""), 0755) // Should be excluded

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	require.NoError(t, err)
	assert.Len(t, exes, 4)
	assert.Contains(t, exes, filepath.Join(binDir, "dummy1"))
}
