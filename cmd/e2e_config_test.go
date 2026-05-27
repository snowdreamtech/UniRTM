package cmd

import (
	"path/filepath"
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
)

func TestE2E_Config(t *testing.T) {
	h := NewE2EHarness(t)
	// Initialize an empty config
	_, _, err := h.Run("config", "ls")
	assert.NoError(t, err)

	// Set a setting
	_, _, err = h.Run("config", "set", "settings.verbose", "true")
	assert.NoError(t, err)

	// Get a setting
	stdout, _, err := h.Run("config", "get", "settings.verbose")
	assert.NoError(t, err)
	_ = stdout
	
	// Trust
	dummyConfig := filepath.Join(h.TmpDir, ".unirtm.toml")
	os.WriteFile(dummyConfig, []byte(""), 0644)
	
	_, _, err = h.Run("trust", dummyConfig)
	assert.NoError(t, err)
	
	_, _, err = h.Run("untrust", dummyConfig)
	assert.NoError(t, err)
}
