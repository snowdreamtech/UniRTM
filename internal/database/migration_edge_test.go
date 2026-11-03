// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // SQLite driver

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationManager_Rollback_EdgeCases(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "rollback_test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: false,
	})
	require.NoError(t, err)
	defer db.Close()

	conn := db.Conn()
	mgr := NewMigrationManager(conn)

	t.Run("rollback when no migrations applied returns error", func(t *testing.T) {
		// Create a fresh DB without schema
		freshPath := filepath.Join(t.TempDir(), "fresh.db")
		freshDB, err := sql.Open("sqlite", freshPath)
		require.NoError(t, err)
		defer freshDB.Close()

		freshMgr := NewMigrationManager(freshDB)
		err = freshMgr.Rollback(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no migrations to rollback")
	})

	t.Run("rollback when migration has no down SQL returns error", func(t *testing.T) {
		// The real migrations have no Down SQL, so Rollback should fail
		err := mgr.Rollback(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no down migration")
	})

	t.Run("rollback works when migration has down SQL", func(t *testing.T) {
		// Create a DB with a schema_migrations table and inject a test migration with Down SQL
		testPath := filepath.Join(t.TempDir(), "test_down.db")
		testDB, err := sql.Open("sqlite", testPath)
		require.NoError(t, err)
		defer testDB.Close()

		// Create schema_migrations table and a test table
		_, err = testDB.ExecContext(ctx, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY,
				description TEXT NOT NULL,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);
			INSERT INTO schema_migrations (version, description) VALUES (1, 'test');
		`)
		require.NoError(t, err)

		// Temporarily inject a migration with Down SQL
		origMigrations := migrations
		migrations = []Migration{
			{
				Version:     1,
				Description: "test migration",
				Up:          "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);",
				Down:        "DROP TABLE IF EXISTS test_table;",
			},
		}
		defer func() { migrations = origMigrations }()

		testMgr := NewMigrationManager(testDB)
		err = testMgr.Rollback(ctx)
		require.NoError(t, err)

		// Version should be 0 now
		version, err := testMgr.GetCurrentVersion(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, version)
	})

	t.Run("rollback returns error when migration version not in list", func(t *testing.T) {
		testPath := filepath.Join(t.TempDir(), "test_notfound.db")
		testDB, err := sql.Open("sqlite", testPath)
		require.NoError(t, err)
		defer testDB.Close()

		// schema_migrations table with version that doesn't match any migration
		_, err = testDB.ExecContext(ctx, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY,
				description TEXT NOT NULL,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO schema_migrations (version, description) VALUES (999, 'unknown');
		`)
		require.NoError(t, err)

		// migrations list has no version 999
		origMigrations := migrations
		migrations = []Migration{
			{Version: 1, Description: "first", Up: "SELECT 1;", Down: ""},
		}
		defer func() { migrations = origMigrations }()

		testMgr := NewMigrationManager(testDB)
		err = testMgr.Rollback(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
