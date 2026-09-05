// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
	"github.com/stretchr/testify/assert"
)

func TestLockCommandStructure(t *testing.T) {
	assert.Equal(t, "lock [tool...]", lockCmd.Use)
}

func TestRunLockCheck(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	// Change working directory to the empty tmpDir so that config.Load()
	// (which calls LoadFromDir(".")) cannot find the project's .unirtm.toml.
	// Without this, the test picks up the real project config and fails the
	// P1 completeness gate because node@20 cannot be resolved in CI.
	t.Chdir(tmpDir)

	// Reset global flags to prevent leakage from other tests.
	lockCheck = false
	lockAllowIncomplete = false

	// Since there is no config file, it should just return no error.
	err := lockCmd.RunE(lockCmd, []string{})
	assert.NoError(t, err)
}

func TestCheckConfigLockSync(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	// Lockfile contains python@3.14.6
	lf := lockfile.New(lfPath)
	lf.UpsertEntry("python", &lockfile.ToolLockEntry{
		Version:   "3.14.6",
		Backend:   "native",
		Platforms: make(map[string]*lockfile.PlatformEntry),
	})

	// Config requests python@3.14.7
	cfg := &config.Config{
		Tools: map[string]config.ToolConfig{
			"python": {Version: "3.14.7"},
		},
	}

	err := checkConfigLockSync(cfg, lf, lfPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "desynchronized with project config")
	assert.Contains(t, err.Error(), "python@3.14.7")

	// Config requests python@3.14.6 (matching)
	cfgMatch := &config.Config{
		Tools: map[string]config.ToolConfig{
			"python": {Version: "3.14.6"},
		},
	}
	errMatch := checkConfigLockSync(cfgMatch, lf, lfPath)
	assert.NoError(t, errMatch)
}

