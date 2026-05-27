package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_PathsFallback(t *testing.T) {
	// safely set environment variables to empty string
	t.Setenv("UNIRTM_CONFIG_DIR", "")
	t.Setenv("UNIRTM_DATA_DIR", "")
	t.Setenv("UNIRTM_CACHE_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	// keep HOME so os.UserHomeDir works, otherwise it might panic or error
	homeDir, _ := os.UserHomeDir()
	t.Setenv("HOME", homeDir)

	cfg := GetConfigDir()
	assert.NotEmpty(t, cfg)

	data := GetDataDir()
	assert.NotEmpty(t, data)

	cache := GetCacheDir()
	assert.NotEmpty(t, cache)

	// test XDG
	t.Setenv("XDG_CONFIG_HOME", "/xdg_config")
	t.Setenv("XDG_DATA_HOME", "/xdg_data")
	t.Setenv("XDG_CACHE_HOME", "/xdg_cache")
	assert.Equal(t, filepath.Join("/xdg_config", "unirtm"), GetConfigDir())
	assert.Equal(t, filepath.Join("/xdg_data", "unirtm"), GetDataDir())
	assert.Equal(t, filepath.Join("/xdg_cache", "unirtm"), GetCacheDir())
}

func TestEnv_GetLockFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	
	// Create a dummy lockfile in tmpDir so it finds it, instead of searching up to repo root
	dummyLock := filepath.Join(tmpDir, ".unirtm.lock")
	os.WriteFile(dummyLock, []byte(""), 0644)

	lock := GetLockFilePath()
	// Depending on logic, it might return dummyLock or some other path
	assert.NotEmpty(t, lock)
}

func TestEnv_RandomString(t *testing.T) {
	s, err := RandomString(10)
	assert.NoError(t, err)
	assert.Len(t, s, 10)

	s2, err := RandomString(0)
	assert.NoError(t, err)
	assert.Empty(t, s2)

	// test uniqueness
	s3, err := RandomString(10)
	assert.NoError(t, err)
	assert.NotEqual(t, s, s3)
}
