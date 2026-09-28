// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
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
	"github.com/stretchr/testify/require"
)

func TestWhichCommandStructure(t *testing.T) {
	assert.Equal(t, "which <tool> [version]", whichCmd.Use)
	assert.NotEmpty(t, whichCmd.Short)
	assert.NotNil(t, whichCmd.RunE)
}

func TestIsExecutableFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular non-executable file
	file1 := filepath.Join(tmpDir, "file1.txt")
	os.WriteFile(file1, []byte(""), 0644)
	assert.False(t, isExecutableFile(file1))

	// Create an executable file
	file2 := filepath.Join(tmpDir, "file2.sh")
	os.WriteFile(file2, []byte(""), 0755)

	// Archive extensions
	file3 := filepath.Join(tmpDir, "file3.tar.gz")
	os.WriteFile(file3, []byte(""), 0755)
	assert.False(t, isExecutableFile(file3))
}

func TestRunWhich(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := env.GetDatabasePath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	require.NoError(t, err)

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	installPath := filepath.Join(tmpDir, "installs", "dummy", "1.0.0")
	os.MkdirAll(filepath.Join(installPath, "bin"), 0755)

	// Since we don't mock providers completely, the fallback might look into the disk.
	// But actually, without a provider it might skip. However, "dummy" provider doesn't exist.
	// Wait, we can test the case where the tool is NOT found.

	inst := &repository.Installation{
		Tool:        "dummy",
		Version:     "1.0.0",
		Backend:     "native",
		InstallPath: installPath,
		InstalledAt: time.Now(),
	}
	err = repo.Create(context.Background(), inst)
	require.NoError(t, err)
	db.Close()

	cmd := whichCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Since "dummy" tool doesn't have a real provider in DefaultRegistry that lists executables,
	// it will likely fall back to "not found".
	err = runWhich(cmd, []string{"dummy"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found: dummy")
}

func TestRunWhich_WithVersionArg(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	cmd := whichCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runWhich(cmd, []string{"nonexistent", "2.0.0"})
	assert.Error(t, err)
}

func TestRunWhich_Alias(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := env.GetDatabasePath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	require.NoError(t, err)

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	installPath := filepath.Join(tmpDir, "installs", "node", "20.0.0")
	binDir := filepath.Join(installPath, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	nodeBin := filepath.Join(binDir, "node")
	f, err := os.Create(nodeBin)
	require.NoError(t, err)
	f.Close()
	os.Chmod(nodeBin, 0755)

	inst := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "node",
		InstallPath: installPath,
		InstalledAt: time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), inst))
	db.Close()

	cmd := whichCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// "nodejs" is an alias of "node"
	err = runWhich(cmd, []string{"nodejs"})
	assert.NoError(t, err)
}

func TestRunWhich_AliasPython(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := env.GetDatabasePath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	require.NoError(t, err)

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	installPath := filepath.Join(tmpDir, "installs", "python", "3.12.0")
	binDir := filepath.Join(installPath, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	pyBin := filepath.Join(binDir, "python")
	f, err := os.Create(pyBin)
	require.NoError(t, err)
	f.Close()
	os.Chmod(pyBin, 0755)

	inst := &repository.Installation{
		Tool:        "python",
		Version:     "3.12.0",
		Backend:     "python",
		InstallPath: installPath,
		InstalledAt: time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), inst))
	db.Close()

	cmd := whichCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// "python3" should resolve to "python"
	err = runWhich(cmd, []string{"python3"})
	assert.NoError(t, err)
}
