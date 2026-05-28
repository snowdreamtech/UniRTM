// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package transaction

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "unirtm-transaction-test-*")
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

func TestNewSQLiteTransactionManager(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	assert.NotNil(t, tm)
}

func TestTransactionManager_Begin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	tx, err := tm.Begin(ctx)
	require.NoError(t, err)
	require.NotNil(t, tx)

	// Verify all repositories are available
	assert.NotNil(t, tx.InstallationRepo())
	assert.NotNil(t, tx.CacheRepo())
	assert.NotNil(t, tx.AuditRepo())
	assert.NotNil(t, tx.IndexRepo())

	// Rollback to clean up
	err = tx.Rollback()
	require.NoError(t, err)
}

func TestTransaction_Commit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Start transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Create an installation within the transaction
	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		Provider:    "node",
		InstallPath: "/opt/unirtm/node/20.0.0",
		Checksum:    "abc123",
		Metadata:    "{}",
	}

	err = tx.InstallationRepo().Create(ctx, installation)
	require.NoError(t, err)
	assert.Greater(t, installation.ID, int64(0))

	// Commit the transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify the installation was persisted
	// Create a new connection to verify
	found, err := db.Conn().QueryContext(ctx, "SELECT COUNT(*) FROM installations WHERE tool = ? AND version = ?", "node", "20.0.0")
	require.NoError(t, err)
	defer found.Close()

	var count int
	require.True(t, found.Next())
	err = found.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestTransaction_Rollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Start transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Create an installation within the transaction
	installation := &repository.Installation{
		Tool:        "python",
		Version:     "3.11.0",
		Backend:     "github",
		Provider:    "python",
		InstallPath: "/opt/unirtm/python/3.11.0",
		Checksum:    "def456",
		Metadata:    "{}",
	}

	err = tx.InstallationRepo().Create(ctx, installation)
	require.NoError(t, err)

	// Rollback the transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify the installation was NOT persisted
	found, err := db.Conn().QueryContext(ctx, "SELECT COUNT(*) FROM installations WHERE tool = ? AND version = ?", "python", "3.11.0")
	require.NoError(t, err)
	defer found.Close()

	var count int
	require.True(t, found.Next())
	err = found.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTransaction_MultipleOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Start transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Create installation
	installation := &repository.Installation{
		Tool:        "go",
		Version:     "1.21.0",
		Backend:     "github",
		Provider:    "go",
		InstallPath: "/opt/unirtm/go/1.21.0",
		Checksum:    "ghi789",
		Metadata:    "{}",
	}
	err = tx.InstallationRepo().Create(ctx, installation)
	require.NoError(t, err)

	// Create cache entry
	err = tx.CacheRepo().Set(ctx, "test-key", []byte("test-value"), 1*time.Hour)
	require.NoError(t, err)

	// Create audit log
	auditEntry := &repository.AuditEntry{
		Operation: "install",
		Tool:      "go",
		Version:   "1.21.0",
		Status:    "success",
		Duration:  1000,
		Metadata:  "{}",
	}
	err = tx.AuditRepo().Log(ctx, auditEntry)
	require.NoError(t, err)

	// Create index entry
	indexEntry := &repository.IndexEntry{
		Tool:        "go",
		Description: "The Go programming language",
		Homepage:    "https://go.dev",
		License:     "BSD-3-Clause",
		Backend:     "github",
		Metadata:    "{}",
	}
	err = tx.IndexRepo().Upsert(ctx, indexEntry)
	require.NoError(t, err)

	// Commit all operations
	err = tx.Commit()
	require.NoError(t, err)

	// Verify all operations were persisted
	// Check installation
	var installCount int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM installations WHERE tool = ?", "go").Scan(&installCount)
	require.NoError(t, err)
	assert.Equal(t, 1, installCount)

	// Check cache
	var cacheCount int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM cache WHERE key = ?", "test-key").Scan(&cacheCount)
	require.NoError(t, err)
	assert.Equal(t, 1, cacheCount)

	// Check audit log
	var auditCount int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_log WHERE tool = ?", "go").Scan(&auditCount)
	require.NoError(t, err)
	assert.Equal(t, 1, auditCount)

	// Check index
	var indexCount int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM tool_index WHERE tool = ?", "go").Scan(&indexCount)
	require.NoError(t, err)
	assert.Equal(t, 1, indexCount)
}

func TestTransaction_RollbackOnError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Start transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Create first installation
	installation1 := &repository.Installation{
		Tool:        "rust",
		Version:     "1.70.0",
		Backend:     "github",
		Provider:    "rust",
		InstallPath: "/opt/unirtm/rust/1.70.0",
		Checksum:    "jkl012",
		Metadata:    "{}",
	}
	err = tx.InstallationRepo().Create(ctx, installation1)
	require.NoError(t, err)

	// Try to create duplicate installation (should fail)
	installation2 := &repository.Installation{
		Tool:        "rust",
		Version:     "1.70.0", // Same tool and version
		Backend:     "github",
		Provider:    "rust",
		InstallPath: "/opt/unirtm/rust/1.70.0",
		Checksum:    "jkl012",
		Metadata:    "{}",
	}
	err = tx.InstallationRepo().Create(ctx, installation2)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrAlreadyExists)

	// Rollback the transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify nothing was persisted
	var count int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM installations WHERE tool = ?", "rust").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTransaction_IsolationBetweenTransactions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Start first transaction
	tx1, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Create installation in first transaction
	installation1 := &repository.Installation{
		Tool:        "java",
		Version:     "17.0.0",
		Backend:     "github",
		Provider:    "java",
		InstallPath: "/opt/unirtm/java/17.0.0",
		Checksum:    "mno345",
		Metadata:    "{}",
	}
	err = tx1.InstallationRepo().Create(ctx, installation1)
	require.NoError(t, err)

	// Commit first transaction
	err = tx1.Commit()
	require.NoError(t, err)

	// Start second transaction after first commits
	tx2, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Second transaction should see the committed data
	installations, err := tx2.InstallationRepo().List(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(installations), "Second transaction should see committed data")
	assert.Equal(t, "java", installations[0].Tool)

	// Rollback second transaction
	err = tx2.Rollback()
	require.NoError(t, err)
}

func TestTransaction_ContextCancellation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Cancel the context
	cancel()

	// Try to create an installation with cancelled context
	installation := &repository.Installation{
		Tool:        "ruby",
		Version:     "3.2.0",
		Backend:     "github",
		Provider:    "ruby",
		InstallPath: "/opt/unirtm/ruby/3.2.0",
		Checksum:    "pqr678",
		Metadata:    "{}",
	}

	// This should fail due to context cancellation
	err = tx.InstallationRepo().Create(ctx, installation)
	require.Error(t, err)

	// Rollback should still work, but may return an error if already rolled back by context cancellation
	err = tx.Rollback()
	if err != nil {
		assert.Contains(t, err.Error(), "transaction has already been committed or rolled back")
	}
}

func TestTransaction_CommitAfterRollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Rollback first
	err = tx.Rollback()
	require.NoError(t, err)

	// Try to commit after rollback (should fail)
	err = tx.Commit()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sql: transaction has already been committed or rolled back")
}

func TestTransaction_RollbackAfterCommit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Commit first
	err = tx.Commit()
	require.NoError(t, err)

	// Try to rollback after commit (should fail or be ignored)
	err = tx.Rollback()
	if err != nil {
		assert.Contains(t, err.Error(), "sql: transaction has already been committed or rolled back")
	}
}
