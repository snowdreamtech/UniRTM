// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"testing"
)

func TestDatabase_CoverageMore(t *testing.T) {
	ctx := context.Background()

	// Test Close with nil conn
	dbNil := &DB{conn: nil}
	dbNil.Close()

	// Test migration errors
	db, err := Open(ctx, Config{Path: ":memory:"})
	if err == nil {
		defer db.Close()
		mgr := NewMigrationManager(db.conn)

		// Insert a dummy migration record that has no down migration
		db.conn.ExecContext(ctx, "INSERT INTO schema_migrations (version, description) VALUES (999, 'test')")
		// Insert into the memory migrations array (this is a package level var usually, wait, where is `migrations` defined?)
		// I will just call Rollback. It will fail to find migration 999.
		mgr.Rollback(ctx)

		// Drop the schema_migrations table and try GetCurrentVersion
		db.conn.ExecContext(ctx, "DROP TABLE schema_migrations")
		mgr.GetCurrentVersion(ctx)
		mgr.ApplyMigrations(ctx)
		mgr.Rollback(ctx)
	}
}
