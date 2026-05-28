// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestFormatListSize(t *testing.T) {
	assert.Equal(t, "0 B", formatListSize(0))
	assert.Equal(t, "0 B", formatListSize(-10))
	assert.Equal(t, "500 B", formatListSize(500))
	assert.Equal(t, "1.0 KB", formatListSize(1024))
	assert.Equal(t, "1.5 MB", formatListSize(1024*1536))
	assert.Equal(t, "2.0 GB", formatListSize(1024*1024*2048))
}

func TestIsPathUnder(t *testing.T) {
	assert.True(t, isPathUnder("/a/b/c", "/a/b"))
	assert.False(t, isPathUnder("/a/b", "/a/b/c"))
	assert.False(t, isPathUnder("/a/x", "/a/b"))
}

func TestDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("12345"), 0644)
	assert.NoError(t, err)

	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
	err = os.WriteFile(filepath.Join(tmpDir, "sub", "file2.txt"), []byte("1234567890"), 0644)
	assert.NoError(t, err)

	size, err := dirSize(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), size)

	_, err = dirSize("/path/does/not/exist/surely")
	assert.Error(t, err)
}

func TestResolveActiveVersions(t *testing.T) {
	tmpDir := t.TempDir()
	installsDir := filepath.Join(tmpDir, "installs")
	shimsDir := filepath.Join(tmpDir, "shims")

	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	os.MkdirAll(filepath.Join(installsDir, "node", "20.0.0", "bin"), 0755)
	os.MkdirAll(shimsDir, 0755)

	// Create a binary
	binFile := filepath.Join(installsDir, "node", "20.0.0", "bin", "node")
	os.WriteFile(binFile, []byte("dummy"), 0755)

	// Create a shim symlink pointing to the binary
	shimFile := filepath.Join(shimsDir, "node")
	os.Symlink(binFile, shimFile)

	installations := []*repository.Installation{
		{Tool: "node", Version: "20.0.0", InstallPath: filepath.Join(installsDir, "node", "20.0.0")},
		{Tool: "node", Version: "18.0.0", InstallPath: filepath.Join(installsDir, "node", "18.0.0")},
	}

	active := resolveActiveVersions(shimsDir, installations)
	assert.Equal(t, "20.0.0", active["node"])
}

func TestRunList(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	dbPath := env.GetDatabasePath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	assert.NoError(t, err)

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	assert.NoError(t, err)

	inst1 := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "native",
		InstallPath: "/tmp/node/20.0.0",
		InstalledAt: time.Now(),
	}
	inst2 := &repository.Installation{
		Tool:        "go",
		Version:     "1.21.0",
		Backend:     "native",
		InstallPath: "/tmp/go/1.21.0",
		InstalledAt: time.Now(),
	}

	err = repo.Create(context.Background(), inst1)
	assert.NoError(t, err)
	err = repo.Create(context.Background(), inst2)
	assert.NoError(t, err)
	db.Close()

	// Test runList normally
	listToolFilter = ""
	listCurrentOnly = false
	err = runList(listCmd, []string{})
	assert.NoError(t, err)

	// Test runList with tool filter
	listToolFilter = "node"
	err = runList(listCmd, []string{})
	assert.NoError(t, err)

	// Test runList with JSON output
	listToolFilter = ""
	jsonOutput = true
	err = runList(listCmd, []string{})
	assert.NoError(t, err)
	jsonOutput = false

	// Test runList empty DB (clear db)
	os.Remove(dbPath)
	db2, _ := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	sqlite.NewInstallationRepository(db2.Conn())
	db2.Close()

	err = runList(listCmd, []string{})
	assert.NoError(t, err)
}
