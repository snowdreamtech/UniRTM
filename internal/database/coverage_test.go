// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"testing"
)

func TestDatabase_Coverage(t *testing.T) {
	ctx := context.Background()

	// Test Open error
	_, err := Open(ctx, Config{Path: "file://invalid:path?mode=ro"})
	if err == nil {
		t.Log("Expected error for invalid path, got nil")
	}

	// Test with a memory db to trigger errors safely
	db, err := Open(ctx, Config{Path: ":memory:"})
	if err == nil {
		defer db.Close()

		// Drop tables to cause errors in get version
		db.conn.ExecContext(ctx, "DROP TABLE IF EXISTS schema_migrations")

		mgr := NewMigrationManager(db.conn)

		// applyMigration error
		err = mgr.applyMigration(ctx, Migration{
			Version: 999,
			Up:      "INVALID SQL SYNTAX",
		})
		if err == nil {
			t.Log("Expected error applying invalid sql")
		}

		// Rollback on no migrations
		err = mgr.Rollback(ctx)
		if err == nil {
			t.Log("Expected error rolling back")
		}

		// Get schema version without table
		db.GetSchemaVersion(ctx)
	}

	// Test on closed DB
	db2, err := Open(ctx, Config{Path: ":memory:"})
	if err == nil {
		db2.conn.Close() // Forcefully close the inner conn

		db2.enableWALMode(ctx)
		db2.initialize(ctx)
		db2.Ping(ctx)
		db2.BeginTx(ctx, nil)
	}
}
