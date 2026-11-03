// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestRunLink(t *testing.T) {
	// Set up isolated environment
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Ensure DB parent directory is initialized
	os.MkdirAll(filepath.Dir(env.GetDatabasePath()), 0755)

	// Create a dummy installation path
	installPath := filepath.Join(tmpDir, "dummy_tool")
	err := os.MkdirAll(installPath, 0755)
	assert.NoError(t, err)

	// Set backend parameter explicitly
	linkBackend = "custom"

	// Test successful link
	err = runLink(linkCmd, []string{"mytool", "1.0.0", installPath})
	assert.NoError(t, err)

	// Test missing path
	err = runLink(linkCmd, []string{"mytool", "2.0.0", filepath.Join(tmpDir, "non_existent")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")

	// Test linking a file (it should warn but still work)
	dummyFile := filepath.Join(tmpDir, "dummy_file.txt")
	err = os.WriteFile(dummyFile, []byte("test"), 0644)
	assert.NoError(t, err)

	err = runLink(linkCmd, []string{"mytool", "3.0.0", dummyFile})
	assert.NoError(t, err)
}
