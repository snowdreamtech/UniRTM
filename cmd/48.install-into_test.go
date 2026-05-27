package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestRunInstallInto(t *testing.T) {
	// Set up isolated environment
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	// Ensure DB parent directory is initialized
	os.MkdirAll(filepath.Dir(env.GetDatabasePath()), 0755)

	installPath := filepath.Join(tmpDir, "custom_install")

	// Test successful install-into
	err := runInstallInto(installIntoCmd, []string{"mytool@1.0.0", installPath})
	assert.NoError(t, err)

	// Ensure the directory was created
	info, err := os.Stat(installPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Test without specific version (should default to "latest")
	installPath2 := filepath.Join(tmpDir, "custom_install2")
	err = runInstallInto(installIntoCmd, []string{"another_tool", installPath2})
	assert.NoError(t, err)

	// Test with specific backend
	installIntoBackend = "native"
	installPath3 := filepath.Join(tmpDir, "custom_install3")
	err = runInstallInto(installIntoCmd, []string{"third_tool", installPath3})
	assert.NoError(t, err)
}

