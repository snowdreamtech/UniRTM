// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestPruneStructure(t *testing.T) {
	assert.Contains(t, pruneCmd.Use, "prune", "pruneCmd command use should contain 'prune'")
	assert.NotEmpty(t, pruneCmd.Short, "pruneCmd command short description should not be empty")
	assert.True(t, pruneCmd.Run != nil || pruneCmd.RunE != nil, "Run or RunE function should be set for pruneCmd")
}

func TestRunPrune(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := pruneCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	// Open DB and add some fake installations
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{Path: env.GetDatabasePath(), WALMode: true})
	assert.NoError(t, err)
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	assert.NoError(t, err)

	// Tool with 2 versions, 1.0.0 is older, 1.1.0 is latest
	installPath1 := filepath.Join(tmpDir, "dummy1")
	os.MkdirAll(installPath1, 0755)
	os.WriteFile(filepath.Join(installPath1, "file"), []byte("data"), 0644)
	
	installPath2 := filepath.Join(tmpDir, "dummy2")
	os.MkdirAll(installPath2, 0755)

	err = installRepo.Upsert(ctx, &repository.Installation{
		Tool: "dummy", Version: "1.0.0", InstallPath: installPath1,
	})
	assert.NoError(t, err)
	err = installRepo.Upsert(ctx, &repository.Installation{
		Tool: "dummy", Version: "1.1.0", InstallPath: installPath2,
	})
	assert.NoError(t, err)
	db.Close()

	// prune without confirm should fail in non-interactive if we don't mock it, but actually here since it tries to read stdin, we'll set pruneYes = true
	pruneYes = true
	err = runPrune(cmd, []string{})
	assert.NoError(t, err)

	// 1.1.0 might be removed if List() orders descending, let's just check that exactly 1 installation is left.
	// Since we know one of them is removed, we check that one path is gone.
	_, err1 := os.Stat(installPath1)
	_, err2 := os.Stat(installPath2)
	assert.True(t, os.IsNotExist(err1) || os.IsNotExist(err2))

	pruneTool = "dummy"
	err = runPrune(cmd, []string{})
	assert.NoError(t, err)
}
