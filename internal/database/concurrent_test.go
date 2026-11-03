// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simulateToolInstall simulates the exact write pattern of a tool installation:
// an UPSERT into the installations table followed by an audit_log insert,
// both inside the same transaction. This mirrors the real-world repository call.
func simulateToolInstall(ctx context.Context, db *DB, tool, version string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// UPSERT installation record (mirrors installation_repository.go)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum)
		VALUES (?, ?, 'github', 'generic', ?, ?)
		ON CONFLICT(tool, version) DO UPDATE SET
			install_path = excluded.install_path,
			checksum     = excluded.checksum
	`, tool, version,
		fmt.Sprintf("/tools/%s/%s", tool, version),
		fmt.Sprintf("sha256-%s-%s", tool, version),
	)
	if err != nil {
		return fmt.Errorf("upsert installation: %w", err)
	}

	// Write audit log
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_log (operation, tool, version, status, duration_ms)
		VALUES ('install', ?, ?, 'success', ?)
	`, tool, version, 100)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return tx.Commit()
}

// TestConcurrentWrites_8Workers mirrors the real CI scenario with 8 parallel jobs.
func TestConcurrentWrites_8Workers(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "concurrent_8.db")})
	require.NoError(t, err)
	defer db.Close()

	tools := []string{
		"node", "go", "python", "ruff",
		"shellcheck", "hadolint", "gitleaks", "actionlint",
	}

	var wg sync.WaitGroup
	var errCount atomic.Int64

	for _, tool := range tools {
		wg.Add(1)
		go func(tool string) {
			defer wg.Done()
			if err := simulateToolInstall(ctx, db, tool, "1.0.0"); err != nil {
				t.Logf("ERROR installing %s: %v", tool, err)
				errCount.Add(1)
			}
		}(tool)
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "expected 0 errors from 8 concurrent installs")

	// Verify all 8 rows were committed
	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 8, count, "all 8 tools should be persisted")
}

// TestConcurrentWrites_16Workers doubles the load to 16 parallel goroutines.
func TestConcurrentWrites_16Workers(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "concurrent_16.db")})
	require.NoError(t, err)
	defer db.Close()

	const workers = 16
	var wg sync.WaitGroup
	var errCount atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tool := fmt.Sprintf("tool-%d", i)
			if err := simulateToolInstall(ctx, db, tool, "2.0.0"); err != nil {
				t.Logf("ERROR installing %s: %v", tool, err)
				errCount.Add(1)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "expected 0 errors from 16 concurrent installs")

	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, workers, count, "all 16 tools should be persisted")
}

// TestConcurrentWrites_32Workers pushes to extreme concurrency: 32 goroutines.
func TestConcurrentWrites_32Workers(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "concurrent_32.db")})
	require.NoError(t, err)
	defer db.Close()

	const workers = 32
	var wg sync.WaitGroup
	var errCount atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tool := fmt.Sprintf("extreme-tool-%d", i)
			if err := simulateToolInstall(ctx, db, tool, "3.0.0"); err != nil {
				t.Logf("ERROR installing %s: %v", tool, err)
				errCount.Add(1)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "expected 0 errors from 32 concurrent installs")

	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, workers, count, "all 32 tools should be persisted")
}

// TestConcurrentUpsert_Idempotency verifies that concurrent UPSERT of the SAME
// tool+version produces exactly 1 row with no errors — critical for retry safety.
func TestConcurrentUpsert_Idempotency(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "upsert_idempotent.db")})
	require.NoError(t, err)
	defer db.Close()

	const workers = 16
	var wg sync.WaitGroup
	var errCount atomic.Int64

	// All goroutines race to upsert the exact same (tool, version) pair.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := simulateToolInstall(ctx, db, "node", "20.0.0"); err != nil {
				errCount.Add(1)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "concurrent upserts of same key should never error")

	// UPSERT should result in exactly 1 row, not 16.
	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations WHERE tool='node'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "idempotent upsert should yield exactly 1 row")
}

// TestConcurrentMixedReadWrite fires readers and writers simultaneously to
// ensure WAL mode allows reads to proceed without blocking on write transactions.
func TestConcurrentMixedReadWrite(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "mixed_rw.db")})
	require.NoError(t, err)
	defer db.Close()

	// Seed one row so reads have something to find.
	require.NoError(t, simulateToolInstall(ctx, db, "seed-tool", "1.0.0"))

	const writers = 8
	const readers = 8
	var wg sync.WaitGroup
	var writeErr, readErr atomic.Int64

	// Launch writers
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := simulateToolInstall(ctx, db, fmt.Sprintf("write-tool-%d", i), "1.0.0"); err != nil {
				writeErr.Add(1)
			}
		}(i)
	}

	// Launch concurrent readers
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var count int
			if err := db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&count); err != nil {
				readErr.Add(1)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(0), writeErr.Load(), "expected 0 write errors in mixed R/W test")
	assert.Equal(t, int64(0), readErr.Load(), "expected 0 read errors in mixed R/W test")
}

// TestConcurrentAuditLog verifies that audit log insertions under high concurrency
// never lose records or produce errors — audit rows don't have UNIQUE constraints
// so they should all succeed.
func TestConcurrentAuditLog(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "audit_concurrent.db")})
	require.NoError(t, err)
	defer db.Close()

	const workers = 16
	var wg sync.WaitGroup
	var errCount atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := db.Conn().ExecContext(ctx, `
				INSERT INTO audit_log (operation, tool, version, status, duration_ms)
				VALUES ('install', ?, '1.0.0', 'success', ?)
			`, fmt.Sprintf("audit-tool-%d", i), i*10)
			if err != nil {
				errCount.Add(1)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "expected 0 errors writing audit logs concurrently")

	var count int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, workers, count, "all audit log entries should be written")
}

// TestBusyTimeout_ExceedGracefully confirms that if every slot in the connection
// pool is occupied by a sleeping transaction, new callers wait rather than fail
// immediately — proving the 30s busy_timeout is effective.
func TestBusyTimeout_ExceedGracefully(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "busy_timeout.db")})
	require.NoError(t, err)
	defer db.Close()

	// Occupy the single connection briefly with a slow write.
	holdDone := make(chan struct{})
	go func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			close(holdDone)
			return
		}
		time.Sleep(200 * time.Millisecond) // Hold lock for 200ms
		_ = tx.Rollback()
		close(holdDone)
	}()

	// Give the goroutine a moment to acquire the lock before we race it.
	time.Sleep(10 * time.Millisecond)

	// This write must succeed; it should queue behind the lock holder.
	err = simulateToolInstall(ctx, db, "queued-tool", "1.0.0")
	assert.NoError(t, err, "write should succeed after waiting for lock release")

	<-holdDone
}

// TestDatabaseConcurrentWrites_DataIntegrity is the most critical test: it verifies
// that after all concurrent writes complete, no rows are missing and counts match.
func TestDatabaseConcurrentWrites_DataIntegrity(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "integrity.db")})
	require.NoError(t, err)
	defer db.Close()

	// Simulate a realistic CI matrix: 32 unique tools, installed concurrently.
	realWorldTools := []struct{ name, version string }{
		{"node", "22.0.0"}, {"go", "1.24.0"}, {"python", "3.12.0"}, {"ruff", "0.11.0"},
		{"shellcheck", "0.10.0"}, {"hadolint", "2.14.0"}, {"gitleaks", "8.20.0"}, {"actionlint", "1.7.0"},
		{"dotenv-linter", "4.0.0"}, {"cli", "2.90.0"}, {"mvdan-sh", "3.13.0"}, {"anchore-syft", "1.44.0"},
		{"osv-scanner", "2.3.0"}, {"checkmake", "0.3.0"}, {"zizmor", "1.24.0"}, {"editorconfig-checker", "3.6.0"},
		{"prettier", "3.8.0"}, {"eslint", "9.3.0"}, {"markdownlint-cli2", "0.22.0"}, {"commitizen", "4.3.0"},
		{"taplo-cli", "0.7.0"}, {"yamllint", "1.38.0"}, {"stylelint", "17.10.0"}, {"pnpm", "9.0.0"},
		{"bats", "1.13.0"}, {"sort-package-json", "2.10.0"}, {"dockerfile-utils", "0.16.0"}, {"clang-format", "18.0.0"},
		{"pre-commit", "3.8.0"}, {"commitlint", "19.0.0"}, {"cz-conventional", "3.3.0"}, {"rust", "1.80.0"},
	}

	var wg sync.WaitGroup
	var errCount atomic.Int64

	for _, tool := range realWorldTools {
		wg.Add(1)
		go func(name, version string) {
			defer wg.Done()
			if err := simulateToolInstall(ctx, db, name, version); err != nil {
				t.Logf("ERROR [%s@%s]: %v", name, version, err)
				errCount.Add(1)
			}
		}(tool.name, tool.version)
	}

	wg.Wait()
	assert.Equal(t, int64(0), errCount.Load(), "zero errors expected across all concurrent installs")

	var installed, audited int
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM installations`).Scan(&installed)
	require.NoError(t, err)
	err = db.Conn().QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&audited)
	require.NoError(t, err)

	assert.Equal(t, len(realWorldTools), installed, "all installations must be persisted")
	assert.Equal(t, len(realWorldTools), audited, "all audit log entries must be persisted")
}
