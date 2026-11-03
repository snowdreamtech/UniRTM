// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileLock_BasicAcquireRelease(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, ".unirtm.toml")

	lock, err := AcquireFileLock(targetFile, 1*time.Second)
	require.NoError(t, err)
	require.NotNil(t, lock)

	// Second attempt with very short timeout should fail with timeout
	_, err2 := AcquireFileLock(targetFile, 50*time.Millisecond)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "timeout")

	// Release first lock
	require.NoError(t, lock.Release())

	// Now third attempt should succeed immediately
	lock3, err3 := AcquireFileLock(targetFile, 1*time.Second)
	require.NoError(t, err3)
	require.NotNil(t, lock3)
	require.NoError(t, lock3.Release())
}

func TestFileLock_ConcurrentWriters_SC004(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, ".unirtm.toml")

	initialContent := `# UniRTM Base Config
[settings]
cache_ttl = 3600
`
	require.NoError(t, os.WriteFile(targetFile, []byte(initialContent), 0644))

	const numWorkers = 5
	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		workerID := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock, err := AcquireFileLock(targetFile, 5*time.Second)
			if err != nil {
				errCh <- fmt.Errorf("worker %d lock fail: %w", workerID, err)
				return
			}
			defer lock.Release()

			// Read current content
			data, err := os.ReadFile(targetFile)
			if err != nil {
				errCh <- fmt.Errorf("worker %d read fail: %w", workerID, err)
				return
			}

			// Add a distinct plugin config key with vendor prefix
			keySegs := []string{"plugins", fmt.Sprintf("vendor/worker-%d", workerID), "config", "port"}
			updated, err := UpsertTOMLSetting(string(data), keySegs, 8000+workerID)
			if err != nil {
				errCh <- fmt.Errorf("worker %d upsert fail: %w", workerID, err)
				return
			}

			// Simulate small IO latency
			time.Sleep(10 * time.Millisecond)

			if err := os.WriteFile(targetFile, []byte(updated), 0644); err != nil {
				errCh <- fmt.Errorf("worker %d write fail: %w", workerID, err)
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}

	// Read final file
	finalData, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	content := string(finalData)

	// Ensure header comment was preserved
	assert.Contains(t, content, "# UniRTM Base Config")

	// Ensure all 5 worker plugin configurations are present without corruption
	for i := 0; i < numWorkers; i++ {
		assert.Contains(t, content, fmt.Sprintf(`[plugins."vendor/worker-%d".config]`, i))
		assert.Contains(t, content, fmt.Sprintf("port = %d", 8000+i))
	}
}

func TestUpsertTOMLSetting_PreservesCommentsAndWhitespace(t *testing.T) {
	inputTOML := `# Header comment of unirtm.toml
# Managed by team infrastructure

[settings]
# Browser choice
browser = "chrome" # inline comment
cache_ttl = 3600

# Docker plugin configuration section
[plugins."snowdreamtech/docker".config]
# Port number for reverse proxy
port = 8080
registry = "harbor.internal"
`

	// 1. Update existing key: port = 9000
	updated, err := UpsertTOMLSetting(inputTOML, []string{"plugins", "snowdreamtech/docker", "config", "port"}, 9000)
	require.NoError(t, err)

	assert.Contains(t, updated, "# Header comment of unirtm.toml")
	assert.Contains(t, updated, "# Managed by team infrastructure")
	assert.Contains(t, updated, "# Browser choice")
	assert.Contains(t, updated, `browser = "chrome" # inline comment`)
	assert.Contains(t, updated, "# Docker plugin configuration section")
	assert.Contains(t, updated, "# Port number for reverse proxy")
	assert.Contains(t, updated, "port = 9000")
	assert.NotContains(t, updated, "port = 8080")
	assert.Contains(t, updated, `registry = "harbor.internal"`)

	// 2. Add new key into existing table: auto_prune = true
	updated2, err := UpsertTOMLSetting(updated, []string{"plugins", "snowdreamtech/docker", "config", "auto_prune"}, true)
	require.NoError(t, err)

	assert.Contains(t, updated2, "auto_prune = true")
	assert.Contains(t, updated2, "# Docker plugin configuration section")
	assert.Contains(t, updated2, "port = 9000")

	// 3. Add brand new plugin table: [plugins."community/deploy".config]
	updated3, err := UpsertTOMLSetting(updated2, []string{"plugins", "community/deploy", "config", "cluster"}, "staging-k8s")
	require.NoError(t, err)

	assert.Contains(t, updated3, `[plugins."community/deploy".config]`)
	assert.Contains(t, updated3, `cluster = "staging-k8s"`)
	// Verify earlier comments are still present
	assert.Contains(t, updated3, "# Header comment of unirtm.toml")
	assert.Contains(t, updated3, "# Port number for reverse proxy")
}

func TestUpsertTOMLSetting_RootSection(t *testing.T) {
	input := `# Root key test
cache_ttl = 100
[tools]
node = "20"
`
	updated, err := UpsertTOMLSetting(input, []string{"cache_ttl"}, 200)
	require.NoError(t, err)
	assert.Contains(t, updated, "cache_ttl = 200")
	assert.Contains(t, updated, "[tools]\nnode = \"20\"")
	assert.Contains(t, updated, "# Root key test")
}
