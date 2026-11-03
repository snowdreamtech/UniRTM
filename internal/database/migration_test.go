// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationManager_GetCurrentVersion(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())

	version, err := mgr.GetCurrentVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, CurrentSchemaVersion, version)
}

func TestMigrationManager_ApplyMigrations(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())

	// Verify migrations were applied
	version, err := mgr.GetCurrentVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, CurrentSchemaVersion, version)

	// Verify migration records exist
	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), count)
}

func TestMigrationManager_ApplyMigrations_Idempotent(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())

	// Apply migrations first time
	err = mgr.ApplyMigrations(ctx)
	require.NoError(t, err)

	version1, err := mgr.GetCurrentVersion(ctx)
	require.NoError(t, err)

	// Apply migrations second time (should be no-op)
	err = mgr.ApplyMigrations(ctx)
	require.NoError(t, err)

	version2, err := mgr.GetCurrentVersion(ctx)
	require.NoError(t, err)

	// Versions should be the same
	assert.Equal(t, version1, version2)

	// Migration count should be the same
	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), count)
}

func TestMigrationManager_TransactionRollback(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	// Insert test data
	_, err = db.Conn().ExecContext(ctx, `
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
		VALUES ('node', '20.0.0', 'github', 'generic', '/path/to/node', 'abc123')
	`)
	require.NoError(t, err)

	// Start a transaction and rollback
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
		VALUES ('python', '3.11.0', 'github', 'generic', '/path/to/python', 'def456')
	`)
	require.NoError(t, err)

	err = tx.Rollback()
	require.NoError(t, err)

	// Verify only the first record exists
	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify the rolled back record doesn't exist
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations WHERE tool='python'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMigrationManager_SchemaVersionTracking(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	// Verify schema version is tracked
	var version int
	var description string
	var appliedAt string

	err = db.Conn().QueryRowContext(ctx, `
		SELECT version, description, applied_at
		FROM schema_migrations
		WHERE version = ?
	`, CurrentSchemaVersion).Scan(&version, &description, &appliedAt)
	require.NoError(t, err)

	assert.Equal(t, CurrentSchemaVersion, version)
	assert.NotEmpty(t, description)
	assert.NotEmpty(t, appliedAt)
}

func TestMigrationManager_MultipleVersions(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())

	// Get all migration versions
	rows, err := db.Conn().QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	require.NoError(t, err)
	defer rows.Close()

	versions := []int{}
	for rows.Next() {
		var version int
		err := rows.Scan(&version)
		require.NoError(t, err)
		versions = append(versions, version)
	}

	// Verify versions are sequential
	for i, version := range versions {
		assert.Equal(t, i+1, version, "migration versions should be sequential")
	}

	// Verify current version matches the highest migration
	currentVersion, err := mgr.GetCurrentVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), currentVersion)
}
