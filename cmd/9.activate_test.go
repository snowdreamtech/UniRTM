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
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveShellType(t *testing.T) {
	tests := []struct {
		name      string
		shellType string
		want      string
		wantErr   bool
	}{
		{"bash", "bash", "bash", false},
		{"zsh", "zsh", "zsh", false},
		{"fish", "fish", "fish", false},
		{"powershell", "powershell", "powershell", false},
		{"pwsh", "pwsh", "powershell", false},
		{"unsupported", "unknown", "", true},
		{"empty (fallback)", "", "bash", false}, // Assuming DetectShell fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveShellType(tt.shellType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.shellType != "" { // Only check if we explicitly provided it
					assert.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func TestRunActivate_SpecificTool(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", InstallPath: filepath.Join(tmpDir, "dummy-tool")})
	require.NoError(t, err)

	activateShell = "bash"
	activateScope = "global"

	cmd := activateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runActivate(cmd, []string{"dummy-tool", "20.0.0"})
	assert.NoError(t, err)
}

func TestRunActivate_AllTools(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", InstallPath: filepath.Join(tmpDir, "dummy-tool")})
	require.NoError(t, err)
	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool2", Version: "1.0.0", InstallPath: filepath.Join(tmpDir, "dummy-tool2")})
	require.NoError(t, err)

	activateShell = "bash"
	activateScope = "global"

	cmd := activateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runActivate(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunActivate_LatestTool(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", InstallPath: filepath.Join(tmpDir, "dummy-tool")})
	require.NoError(t, err)

	activateShell = "bash"
	activateScope = "global"

	cmd := activateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runActivate(cmd, []string{"dummy-tool"})
	assert.NoError(t, err)
}

func TestRunActivate_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	activateShell = "bash"
	activateScope = "global"

	// 1. Tool not found specific version
	err = runActivate(activateCmd, []string{"dummy-tool", "20.0.0"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// 2. Tool not found latest version
	err = runActivate(activateCmd, []string{"dummy-tool"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no installed version of dummy-tool found")

	// 3. Shell error
	activateShell = "invalid-shell"
	err = runActivate(activateCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported shell")
}
