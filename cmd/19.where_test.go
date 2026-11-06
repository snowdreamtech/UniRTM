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

func TestWhereStructure(t *testing.T) {
	assert.Contains(t, whereCmd.Use, "where", "whereCmd command use should contain 'where'")
	assert.NotEmpty(t, whereCmd.Short, "whereCmd command short description should not be empty")
	assert.True(t, whereCmd.Run != nil || whereCmd.RunE != nil, "Run or RunE function should be set for whereCmd")
}

func TestWhere_CustomToolAlias(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := env.GetDatabasePath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := database.Open(context.Background(), database.Config{Path: dbPath, WALMode: true})
	require.NoError(t, err)

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	installPath := filepath.Join(tmpDir, "installs", "github-mvdan-sh", "3.14.1")
	binDir := filepath.Join(installPath, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	inst := &repository.Installation{
		Tool:        "github:mvdan/sh",
		Version:     "3.14.1",
		Backend:     "github",
		InstallPath: installPath,
		InstalledAt: time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), inst))
	db.Close()

	cfgPath := filepath.Join(tmpDir, ".unirtm.toml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
[tools]
"github:mvdan/sh" = { version = "3.14.1", alias = "shfmt" }
`), 0644))

	origWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origWd)

	cmd := whereCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runWhere(cmd, []string{"shfmt"})
	assert.NoError(t, err)
}
