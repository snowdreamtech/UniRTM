// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package updater

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCache_ReadWrite(t *testing.T) {
	// Setup temporary data directory
	tempDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tempDir)

	// Initially cache should be empty
	cache, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, "", cache.LatestVersion)
	assert.True(t, cache.LastChecked.IsZero())
	assert.True(t, cache.LastPrompted.IsZero())

	// Write cache
	now := time.Now()
	expected := &UpdateCache{
		LatestVersion: "1.0.0",
		LastChecked:   now,
		LastPrompted:  now.Add(-24 * time.Hour),
	}
	err = writeCache(expected)
	require.NoError(t, err)

	// Read cache back
	actual, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, expected.LatestVersion, actual.LatestVersion)

	// time.Time comparison over JSON serialization may lose some precision
	assert.Equal(t, expected.LastChecked.Unix(), actual.LastChecked.Unix())
	assert.Equal(t, expected.LastPrompted.Unix(), actual.LastPrompted.Unix())
}

func TestUpdateCache_Corrupt(t *testing.T) {
	// Setup temporary data directory
	tempDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tempDir)

	// Write corrupt JSON
	cachePath := getCachePath()
	err := os.MkdirAll(filepath.Dir(cachePath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(cachePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	// Reading corrupt cache should return empty struct instead of error
	cache, err := readCache()
	require.NoError(t, err)
	assert.Equal(t, "", cache.LatestVersion)
}

func TestFetchLatestRelease(t *testing.T) {
	// Only run this test if explicitly requested, as it makes network calls
	if os.Getenv("TEST_NETWORK") == "" {
		t.Skip("Skipping network test. Set TEST_NETWORK=1 to enable.")
	}

	version, err := fetchLatestRelease()
	require.NoError(t, err)
	assert.NotEmpty(t, version)
}

func TestPromptIfAvailable_Blacklist(t *testing.T) {
	// Tests that blacklisted commands do not output the prompt.
	// Hard to test os.Stderr side-effects directly without intercepting,
	// but we can ensure it doesn't panic and returns early.
	PromptIfAvailable("0.5.0", "env")
	PromptIfAvailable("0.5.0", "version")
	PromptIfAvailable("0.5.0", "completion")
}

func TestFetchLatestRelease_Mock(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name": "v1.2.3"}`))
	}))
	defer server.Close()

	oldURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = oldURL }()

	var wg sync.WaitGroup
	errs := make([]error, 5)
	versions := make([]string, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v, err := fetchLatestRelease()
			versions[idx] = v
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i := 0; i < 5; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, "1.2.3", versions[i])
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

