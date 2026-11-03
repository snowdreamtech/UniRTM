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

// TestDatabaseIntegration tests the complete database workflow
func TestDatabaseIntegration(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "integration.db")

	// Requirement 2.1: Initialize SQLite database on first run
	t.Run("initialize database on first run", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Verify database was created
		assert.FileExists(t, dbPath)

		// Verify schema version is set
		version, err := db.GetSchemaVersion(ctx)
		require.NoError(t, err)
		assert.Equal(t, CurrentSchemaVersion, version)
	})

	// Requirement 2.6: Perform automatic migrations when schema changes
	t.Run("automatic migrations on schema changes", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Verify migrations were applied
		var count int
		err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, len(migrations), count)

		// Verify all tables exist
		tables := []string{"installations", "cache", "audit_log", "tool_index"}
		for _, table := range tables {
			var tableCount int
			query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
			err := db.Conn().QueryRowContext(ctx, query, table).Scan(&tableCount)
			require.NoError(t, err)
			assert.Equal(t, 1, tableCount, "table %s should exist", table)
		}
	})

	// Test all table operations
	t.Run("installations table operations", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Insert installation
		_, err = db.Conn().ExecContext(ctx, `
			INSERT INTO installations (tool, version, backend, provider, install_path, checksum, metadata)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, "node", "20.0.0", "github", "generic", "/path/to/node", "abc123", `{"arch":"amd64"}`)
		require.NoError(t, err)

		// Query installation
		var tool, version, backend string
		err = db.Conn().QueryRowContext(ctx, `
			SELECT tool, version, backend FROM installations WHERE tool = ?
		`, "node").Scan(&tool, &version, &backend)
		require.NoError(t, err)
		assert.Equal(t, "node", tool)
		assert.Equal(t, "20.0.0", version)
		assert.Equal(t, "github", backend)
	})

	t.Run("cache table operations", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Insert cache entry
		_, err = db.Conn().ExecContext(ctx, `
			INSERT INTO cache (key, value, expires_at)
			VALUES (?, ?, datetime('now', '+1 day'))
		`, "test-key", []byte("test-value"))
		require.NoError(t, err)

		// Query cache entry
		var value []byte
		err = db.Conn().QueryRowContext(ctx, `
			SELECT value FROM cache WHERE key = ?
		`, "test-key").Scan(&value)
		require.NoError(t, err)
		assert.Equal(t, []byte("test-value"), value)
	})

	t.Run("audit_log table operations", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Insert audit log entry
		_, err = db.Conn().ExecContext(ctx, `
			INSERT INTO audit_log (operation, tool, version, status, duration_ms)
			VALUES (?, ?, ?, ?, ?)
		`, "install", "python", "3.11.0", "success", 1500)
		require.NoError(t, err)

		// Query audit log
		var operation, status string
		var durationMs int
		err = db.Conn().QueryRowContext(ctx, `
			SELECT operation, status, duration_ms FROM audit_log WHERE tool = ?
		`, "python").Scan(&operation, &status, &durationMs)
		require.NoError(t, err)
		assert.Equal(t, "install", operation)
		assert.Equal(t, "success", status)
		assert.Equal(t, 1500, durationMs)
	})

	t.Run("tool_index table operations", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Insert tool index entry
		_, err = db.Conn().ExecContext(ctx, `
			INSERT INTO tool_index (tool, description, homepage, license, backend)
			VALUES (?, ?, ?, ?, ?)
		`, "go", "The Go programming language", "https://go.dev", "BSD-3-Clause", "github")
		require.NoError(t, err)

		// Query tool index
		var tool, description, license string
		err = db.Conn().QueryRowContext(ctx, `
			SELECT tool, description, license FROM tool_index WHERE tool = ?
		`, "go").Scan(&tool, &description, &license)
		require.NoError(t, err)
		assert.Equal(t, "go", tool)
		assert.Equal(t, "The Go programming language", description)
		assert.Equal(t, "BSD-3-Clause", license)
	})

	// Test indexes are working
	t.Run("verify indexes improve query performance", func(t *testing.T) {
		db, err := Open(ctx, Config{
			Path:    dbPath,
			WALMode: true,
		})
		require.NoError(t, err)
		defer db.Close()

		// Query using indexed column should use the index
		rows, err := db.Conn().QueryContext(ctx, `
			EXPLAIN QUERY PLAN
			SELECT * FROM installations WHERE tool = 'node'
		`)
		require.NoError(t, err)
		defer rows.Close()

		// The query plan should mention the index
		foundIndex := false
		for rows.Next() {
			var id, parent, notused int
			var detail string
			err := rows.Scan(&id, &parent, &notused, &detail)
			require.NoError(t, err)
			if contains(detail, "idx_installations_tool") {
				foundIndex = true
				break
			}
		}
		assert.True(t, foundIndex, "query should use idx_installations_tool index")
	})
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
