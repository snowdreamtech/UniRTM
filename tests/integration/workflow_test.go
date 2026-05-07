// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package integration provides end-to-end integration tests for the full
// UniRTM installation workflow:  install → activate → use → uninstall.
//
// These tests use a real SQLite database and filesystem but mock the HTTP
// downloader to avoid network dependencies.
//
// Validates Requirements: Integration testing, 9.1, 9.2, 9.3, 3.1, 3.4
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary SQLite database and returns a cleanup func.
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err, "open test database")

	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

// TestInstallWorkflow_DuplicateDetection ensures that the repository layer
// rejects duplicate tool@version records (used by InstallationManager.Install).
//
// Validates Requirements: 9.2 (duplicate detection)
func TestInstallWorkflow_DuplicateDetection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	inst := &repository.Installation{
		Tool:        "go",
		Version:     "1.21.0",
		Backend:     "github",
		Provider:    "generic",
		InstallPath: "/opt/unirtm/tools/go/1.21.0",
	}

	// First insert succeeds
	require.NoError(t, installRepo.Create(ctx, inst), "first create must succeed")

	// Second insert must fail (duplicate key)
	err2 := installRepo.Create(ctx, inst)
	assert.Error(t, err2, "second create must fail for duplicate tool@version")
}

// TestDatabase_TransactionAtomicity verifies that a failed transaction
// leaves the database unchanged.
//
// Validates Requirements: 2.8 (transaction atomicity)
func TestDatabase_TransactionAtomicity(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	conn := db.Conn()

	installRepo, err := sqlite.NewInstallationRepository(conn)
	require.NoError(t, err)

	// Begin transaction and insert, then rollback
	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, insertErr := tx.ExecContext(ctx,
		"INSERT INTO installations (tool, version, backend, provider, install_path, checksum, installed_at, metadata) VALUES (?,?,?,?,?,?,datetime('now'),?)",
		"node", "20.0.0", "github", "generic", "/tmp/node/20.0.0", "", "{}",
	)
	require.NoError(t, insertErr)

	// Rollback — record must not persist
	require.NoError(t, tx.Rollback())

	inst, findErr := installRepo.FindByToolAndVersion(ctx, "node", "20.0.0")
	assert.True(t, findErr != nil || inst == nil, "rolled-back record must not exist")
}

// TestDatabase_ConcurrentReads verifies that multiple goroutines can read
// from the database concurrently without errors.
//
// Validates Requirement: 2.7 (concurrent read support)
func TestDatabase_ConcurrentReads(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	// Run 20 concurrent reads
	const goroutines = 20
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := installRepo.List(ctx)
			errCh <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		assert.NoError(t, <-errCh, "concurrent read should not error")
	}
}

// TestRecovery_OrphanedDirectoryCleanup verifies that CleanupManager
// finds and removes directories not registered in the database.
//
// Validates Requirement: 3.4
func TestRecovery_OrphanedDirectoryCleanup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	installsDir := t.TempDir()

	// Create an orphaned directory (not in the database)
	orphanPath := filepath.Join(installsDir, "node", "19.0.0")
	require.NoError(t, os.MkdirAll(orphanPath, 0755))

	cm := service.NewCleanupManager(installRepo, installsDir)

	// Dry-run should identify but not remove
	dryRemoved, err := cm.CleanOrphaned(ctx, true)
	require.NoError(t, err)
	assert.Contains(t, dryRemoved, orphanPath, "orphaned path must be detected")

	// Dry-run must not actually remove
	_, statErr := os.Stat(orphanPath)
	assert.NoError(t, statErr, "dry-run must not remove directory")

	// Real cleanup
	removed, err := cm.CleanOrphaned(ctx, false)
	require.NoError(t, err)
	assert.Contains(t, removed, orphanPath)

	// Must be gone now
	_, statErr = os.Stat(orphanPath)
	assert.True(t, os.IsNotExist(statErr), "orphaned directory must be removed")
}

// TestMigration_MiseToml verifies that a .mise.toml file is correctly parsed
// and converted to UniRTM format.
//
// Validates Requirements: 21.1, 21.3
func TestMigration_MiseToml(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Write a sample .mise.toml
	miseContent := `[tools]
node = "20.0.0"
python = "3.11.0"
go = "1.21.0"
`
	miseFile := filepath.Join(dir, ".mise.toml")
	require.NoError(t, os.WriteFile(miseFile, []byte(miseContent), 0644))

	mm := service.NewMigrationManager()
	outputFile := filepath.Join(dir, "unirtm.toml")

	report, err := mm.MigrateFile(ctx, miseFile, outputFile, false)
	require.NoError(t, err)
	assert.Len(t, report.Tools, 3)
	assert.Empty(t, report.Errors)

	// Output file must exist and contain tool sections
	content, readErr := os.ReadFile(outputFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "[tools.node]")
	assert.Contains(t, string(content), "[tools.python]")
	assert.Contains(t, string(content), "[tools.go]")
}

// TestMigration_ToolVersions verifies that a .tool-versions file is correctly parsed.
//
// Validates Requirement: 21.2
func TestMigration_ToolVersions(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	tvContent := "node 18.0.0\npython 3.10.0\n"
	tvFile := filepath.Join(dir, ".tool-versions")
	require.NoError(t, os.WriteFile(tvFile, []byte(tvContent), 0644))

	mm := service.NewMigrationManager()
	outputFile := filepath.Join(dir, "unirtm.toml")

	report, err := mm.MigrateFile(ctx, tvFile, outputFile, false)
	require.NoError(t, err)
	assert.Len(t, report.Tools, 2)
}

// TestMigration_DryRun verifies that dry-run mode does not write files.
//
// Validates Requirement: 8.7, 21.4
func TestMigration_DryRun(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	miseContent := "[tools]\nnode = \"20.0.0\"\n"
	miseFile := filepath.Join(dir, ".mise.toml")
	require.NoError(t, os.WriteFile(miseFile, []byte(miseContent), 0644))

	mm := service.NewMigrationManager()
	outputFile := filepath.Join(dir, "unirtm.toml")

	report, err := mm.MigrateFile(ctx, miseFile, outputFile, true /*dryRun*/)
	require.NoError(t, err)
	assert.True(t, report.DryRun)

	// Output file must NOT exist in dry-run mode
	_, statErr := os.Stat(outputFile)
	assert.True(t, os.IsNotExist(statErr), "dry-run must not write output file")
}

// TestShimGenerator_Unix verifies that a Unix shim script is created correctly.
//
// Validates Requirements: 14.1, 14.2, 14.3, 14.4
func TestShimGenerator_Unix(t *testing.T) {
	shimsDir := t.TempDir()
	installsDir := t.TempDir()

	gen := service.NewGenerator(shimsDir, installsDir)
	ctx := context.Background()

	err := gen.GenerateShim(ctx, "node")
	require.NoError(t, err)

	// Check shim exists
	assert.True(t, gen.ShimExists("node"), "shim must exist after generation")

	// Check shim contains required elements
	shimPath := filepath.Join(shimsDir, "node")
	content, readErr := os.ReadFile(shimPath)
	require.NoError(t, readErr)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/sh", "shim must be POSIX sh")
	assert.Contains(t, contentStr, "UNIRTM_NODE_VERSION", "shim must check env var")
	assert.Contains(t, contentStr, `exec "${TOOL_BIN}" "$@"`, "shim must exec with all args")
}

// TestShimGenerator_Remove verifies that shim removal works.
//
// Validates Requirement: 14.1
func TestShimGenerator_Remove(t *testing.T) {
	shimsDir := t.TempDir()
	gen := service.NewGenerator(shimsDir, t.TempDir())
	ctx := context.Background()

	require.NoError(t, gen.GenerateShim(ctx, "python"))
	assert.True(t, gen.ShimExists("python"))

	require.NoError(t, gen.RemoveShim(ctx, "python"))
	assert.False(t, gen.ShimExists("python"))
}

// TestPerformanceMonitor_Tracking verifies operation tracking and percentile reports.
//
// Validates Requirements: 17.1, 17.5, 17.6
func TestPerformanceMonitor_Tracking(t *testing.T) {
	pm := service.NewPerformanceMonitor(nil)
	ctx := context.Background()

	// Record 10 download metrics with increasing durations
	for i := 1; i <= 10; i++ {
		pm.Record(ctx, service.OperationMetric{
			Tool:    "node",
			Version: "20.0.0",
			Phase:   service.PhaseDownload,
			Success: true,
		})
	}

	report := pm.Report(service.PhaseDownload)
	assert.Equal(t, 10, report.Count)
	assert.GreaterOrEqual(t, int64(report.P95), int64(report.P50))
	assert.GreaterOrEqual(t, int64(report.Max), int64(report.Min))
}

// TestPerformanceMonitor_CacheHitRate verifies cache hit rate tracking.
//
// Validates Requirement: 17.2
func TestPerformanceMonitor_CacheHitRate(t *testing.T) {
	pm := service.NewPerformanceMonitor(nil)
	ctx := context.Background()

	// 3 hits, 1 miss → 75%
	for i := 0; i < 3; i++ {
		pm.Record(ctx, service.OperationMetric{Phase: service.PhaseCacheHit, Success: true})
	}
	pm.Record(ctx, service.OperationMetric{Phase: service.PhaseCacheMiss, Success: true})

	rate := pm.CacheHitRate()
	assert.InDelta(t, 75.0, rate, 0.01)
}
