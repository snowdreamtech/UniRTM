package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "unirtm-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := database.Open(context.Background(), database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestInstallationRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		Provider:    "node",
		InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
		Checksum:    "abc123def456",
		Metadata:    `{"arch":"x64","os":"linux"}`,
	}

	err = repo.Create(ctx, installation)
	require.NoError(t, err)
	assert.NotZero(t, installation.ID)
}

func TestInstallationRepository_Create_Duplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		Provider:    "node",
		InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
		Checksum:    "abc123def456",
		Metadata:    `{}`,
	}

	// First create should succeed
	err = repo.Create(ctx, installation)
	require.NoError(t, err)

	// Second create with same tool and version should fail
	err = repo.Create(ctx, installation)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrAlreadyExists)
}

func TestInstallationRepository_FindByToolAndVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create installation
	installation := &repository.Installation{
		Tool:        "python",
		Version:     "3.11.0",
		Backend:     "github",
		Provider:    "python",
		InstallPath: "/usr/local/unirtm/installs/python/3.11.0",
		Checksum:    "xyz789abc123",
		Metadata:    `{"arch":"arm64"}`,
	}

	err = repo.Create(ctx, installation)
	require.NoError(t, err)

	// Find the installation
	found, err := repo.FindByToolAndVersion(ctx, "python", "3.11.0")
	require.NoError(t, err)
	assert.Equal(t, installation.Tool, found.Tool)
	assert.Equal(t, installation.Version, found.Version)
	assert.Equal(t, installation.Backend, found.Backend)
	assert.Equal(t, installation.Checksum, found.Checksum)
	assert.NotZero(t, found.InstalledAt)
}

func TestInstallationRepository_FindByToolAndVersion_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Try to find non-existent installation
	_, err = repo.FindByToolAndVersion(ctx, "nonexistent", "1.0.0")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestInstallationRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create multiple installations
	installations := []*repository.Installation{
		{
			Tool:        "node",
			Version:     "18.0.0",
			Backend:     "github",
			Provider:    "node",
			InstallPath: "/usr/local/unirtm/installs/node/18.0.0",
			Checksum:    "aaa111",
			Metadata:    `{}`,
		},
		{
			Tool:        "node",
			Version:     "20.0.0",
			Backend:     "github",
			Provider:    "node",
			InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
			Checksum:    "bbb222",
			Metadata:    `{}`,
		},
		{
			Tool:        "python",
			Version:     "3.11.0",
			Backend:     "github",
			Provider:    "python",
			InstallPath: "/usr/local/unirtm/installs/python/3.11.0",
			Checksum:    "ccc333",
			Metadata:    `{}`,
		},
	}

	for _, inst := range installations {
		err := repo.Create(ctx, inst)
		require.NoError(t, err)
		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// List all installations
	list, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Verify order (most recent first)
	assert.Equal(t, "python", list[0].Tool)
	assert.Equal(t, "node", list[1].Tool)
	assert.Equal(t, "20.0.0", list[1].Version)
}

func TestInstallationRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create installation
	installation := &repository.Installation{
		Tool:        "go",
		Version:     "1.21.0",
		Backend:     "github",
		Provider:    "go",
		InstallPath: "/usr/local/unirtm/installs/go/1.21.0",
		Checksum:    "def456",
		Metadata:    `{}`,
	}

	err = repo.Create(ctx, installation)
	require.NoError(t, err)

	// Delete the installation
	err = repo.Delete(ctx, "go", "1.21.0")
	require.NoError(t, err)

	// Verify it's deleted
	_, err = repo.FindByToolAndVersion(ctx, "go", "1.21.0")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestInstallationRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Try to delete non-existent installation
	err = repo.Delete(ctx, "nonexistent", "1.0.0")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
