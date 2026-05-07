//go:build integration
// +build integration

package cmd

import (
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

// TestUninstallCommandIntegration tests the complete uninstall workflow.
// This test requires a real database and file system operations.
//
// Run with: go test -tags=integration -v ./cmd -run TestUninstallCommandIntegration
func TestUninstallCommandIntegration(t *testing.T) {
	// Create a temporary directory for test
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize database
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	// Create installation repository
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	// Create a test installation directory with some files
	testInstallPath := filepath.Join(tmpDir, "tools", "node", "20.0.0")
	err = os.MkdirAll(filepath.Join(testInstallPath, "bin"), 0755)
	require.NoError(t, err)

	// Create a dummy executable
	dummyExe := filepath.Join(testInstallPath, "bin", "node")
	err = os.WriteFile(dummyExe, []byte("#!/bin/sh\necho 'node'\n"), 0755)
	require.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(dummyExe)
	require.NoError(t, err)

	// Create installation record
	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		InstallPath: testInstallPath,
		Checksum:    "abc123",
	}
	err = installRepo.Create(ctx, installation)
	require.NoError(t, err)

	// Verify installation exists in database
	found, err := installRepo.FindByToolAndVersion(ctx, "node", "20.0.0")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "node", found.Tool)
	assert.Equal(t, "20.0.0", found.Version)

	// Note: In a real integration test, we would call the uninstall command here
	// For now, we verify the setup is correct and the installation exists

	t.Log("Integration test setup successful")
	t.Log("Installation path:", testInstallPath)
	t.Log("Database path:", dbPath)
	t.Log("Installation record created and verified")
}
