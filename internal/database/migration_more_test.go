// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationManager_GetCurrentVersion_Error(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Close db to induce query error
	db.Close()
	_, err = mgr.GetCurrentVersion(ctx)
	require.Error(t, err)
}

func TestMigrationManager_ApplyMigration_Error(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Inject a bad migration
	badMigration := Migration{
		Version:     9999,
		Description: "Bad Migration",
		Up:          "INVALID SQL SYNTAX;",
	}
	err = mgr.applyMigration(ctx, badMigration)
	require.Error(t, err)
	db.Close()
}

func TestMigrationManager_Rollback_Error(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Close db to induce rollback error (the tx begins, or querying schema version fails)
	db.Close()
	err = mgr.Rollback(ctx)
	require.Error(t, err)
}

func TestMigrationManager_Rollback_NoMigrations(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")
	// Not applying migrations, so current version is 0
	err = mgr.Rollback(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no migrations to rollback")
}

func TestMigrationManager_Rollback_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Inject a fake version into the table
	_, err = db.Conn().ExecContext(ctx, `INSERT INTO schema_migrations (version, description) VALUES (?, ?)`, 9999, "Fake")
	require.NoError(t, err)

	err = mgr.Rollback(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMigrationManager_Rollback_NoDown(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Temporarily add a migration with no Down script to the global migrations slice
	testVersion := 9998
	migrations = append(migrations, Migration{
		Version:     testVersion,
		Description: "No Down",
		Up:          "CREATE TABLE temp_no_down (id INTEGER);",
		Down:        "",
	})

	// Defer cleanup of the global slice
	defer func() {
		migrations = migrations[:len(migrations)-1]
	}()

	// Apply migrations

	// Inject testVersion as the max version if not already
	_, err = db.Conn().ExecContext(ctx, `INSERT INTO schema_migrations (version, description) VALUES (?, ?)`, testVersion, "Fake No Down")
	require.NoError(t, err)

	err = mgr.Rollback(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "has no down migration")
}

func TestMigrationManager_ApplyMigrations_BeginTxError2(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(ctx, Config{Path: dbPath, WALMode: false})
	require.NoError(t, err)
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())
	db.Conn().ExecContext(ctx, "DELETE FROM schema_migrations")

	// Cancel context to make BeginTx fail (or Exec fail)
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = mgr.applyMigration(cancelCtx, Migration{
		Version:     9997,
		Description: "Cancelled",
		Up:          "CREATE TABLE temp_cancelled (id INTEGER);",
	})
	require.Error(t, err)
}
