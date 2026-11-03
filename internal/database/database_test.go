// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with WAL mode",
			config: Config{
				Path:    filepath.Join(t.TempDir(), "test.db"),
				WALMode: true,
			},
			wantErr: false,
		},
		{
			name: "valid config without WAL mode",
			config: Config{
				Path:    filepath.Join(t.TempDir(), "test.db"),
				WALMode: false,
			},
			wantErr: false,
		},
		{
			name: "nested directory path",
			config: Config{
				Path:    filepath.Join(t.TempDir(), "nested", "dir", "test.db"),
				WALMode: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db, err := Open(ctx, tt.config)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)
			defer db.Close()

			// Verify database file was created
			_, err = os.Stat(tt.config.Path)
			require.NoError(t, err)

			// Verify connection is alive
			err = db.Ping(ctx)
			require.NoError(t, err)

			// Verify schema version
			version, err := db.GetSchemaVersion(ctx)
			require.NoError(t, err)
			assert.Equal(t, CurrentSchemaVersion, version)
		})
	}
}

func TestDatabaseInitialization(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	// Verify all tables exist
	tables := []string{"installations", "cache", "audit_log", "tool_index", "schema_migrations"}
	for _, table := range tables {
		var count int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
		err := db.Conn().QueryRowContext(ctx, query, table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}

	// Verify indexes exist
	indexes := []string{
		"idx_installations_tool",
		"idx_installations_installed_at",
		"idx_cache_expires_at",
		"idx_audit_log_timestamp",
		"idx_audit_log_operation",
		"idx_audit_log_tool",
		"idx_tool_index_backend",
		"idx_tool_index_updated_at",
	}
	for _, index := range indexes {
		var count int
		query := `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`
		err := db.Conn().QueryRowContext(ctx, query, index).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "index %s should exist", index)
	}
}

func TestDatabaseReopen(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Open database first time
	db1, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)

	// Insert test data
	_, err = db1.Conn().ExecContext(ctx, `
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
		VALUES ('node', '20.0.0', 'github', 'generic', '/path/to/node', 'abc123')
	`)
	require.NoError(t, err)
	db1.Close()

	// Reopen database
	db2, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db2.Close()

	// Verify data persisted
	var count int
	err = db2.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify schema version is still correct
	version, err := db2.GetSchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, CurrentSchemaVersion, version)
}

func TestDatabaseTransaction(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	t.Run("successful transaction", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
			VALUES ('python', '3.11.0', 'github', 'generic', '/path/to/python', 'def456')
		`)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify data was committed
		var count int
		err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations WHERE tool='python'`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("rolled back transaction", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
			VALUES ('go', '1.21.0', 'github', 'generic', '/path/to/go', 'ghi789')
		`)
		require.NoError(t, err)

		err = tx.Rollback()
		require.NoError(t, err)

		// Verify data was not committed
		var count int
		err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations WHERE tool='go'`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestDatabaseWALMode(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	// Verify WAL mode is enabled
	var journalMode string
	err = db.Conn().QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)
}

func TestDatabaseConcurrentReads(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	// Insert test data
	_, err = db.Conn().ExecContext(ctx, `
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
		VALUES ('node', '20.0.0', 'github', 'generic', '/path/to/node', 'abc123')
	`)
	require.NoError(t, err)

	// Perform concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			var count int
			err := db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
