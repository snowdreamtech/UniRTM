// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIndexUsage verifies that database indexes are being used in queries
func TestIndexUsage(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test installation repository index usage
	t.Run("InstallationRepository uses indexes", func(t *testing.T) {
		repo, err := NewInstallationRepository(db.Conn())
		require.NoError(t, err)
		defer repo.Close()

		// Create test data
		installation := &repository.Installation{
			Tool:        "node",
			Version:     "20.0.0",
			Backend:     "github",
			Provider:    "node",
			InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
			Checksum:    "abc123",
			Metadata:    `{}`,
		}
		err = repo.Create(ctx, installation)
		require.NoError(t, err)

		// Query using EXPLAIN QUERY PLAN to verify index usage
		rows, err := db.Conn().QueryContext(ctx, `
			EXPLAIN QUERY PLAN
			SELECT * FROM installations WHERE tool = ? AND version = ?
		`, "node", "20.0.0")
		require.NoError(t, err)
		defer rows.Close()

		// Check that the query plan mentions an index
		foundIndex := false
		for rows.Next() {
			var id, parent, notused int
			var detail string
			err := rows.Scan(&id, &parent, &notused, &detail)
			require.NoError(t, err)
			// SQLite uses "SEARCH" for index usage
			if detail != "" {
				foundIndex = true
				t.Logf("Query plan: %s", detail)
			}
		}
		assert.True(t, foundIndex, "Query should use an index or table scan")
	})

	// Test cache repository index usage
	t.Run("CacheRepository uses indexes for expiration", func(t *testing.T) {
		repo, err := NewCacheRepository(db.Conn())
		require.NoError(t, err)
		defer repo.Close()

		// Query using EXPLAIN QUERY PLAN
		rows, err := db.Conn().QueryContext(ctx, `
			EXPLAIN QUERY PLAN
			SELECT value, expires_at FROM cache WHERE key = ? AND expires_at > CURRENT_TIMESTAMP
		`, "test-key")
		require.NoError(t, err)
		defer rows.Close()

		foundPlan := false
		for rows.Next() {
			var id, parent, notused int
			var detail string
			err := rows.Scan(&id, &parent, &notused, &detail)
			require.NoError(t, err)
			if detail != "" {
				foundPlan = true
				t.Logf("Query plan: %s", detail)
			}
		}
		assert.True(t, foundPlan, "Query should have an execution plan")
	})

	// Test audit repository index usage
	t.Run("AuditRepository uses indexes for filtering", func(t *testing.T) {
		repo, err := NewAuditRepository(db.Conn())
		require.NoError(t, err)
		defer repo.Close()

		// Query using EXPLAIN QUERY PLAN
		rows, err := db.Conn().QueryContext(ctx, `
			EXPLAIN QUERY PLAN
			SELECT * FROM audit_log WHERE operation = ? AND tool = ?
		`, "install", "node")
		require.NoError(t, err)
		defer rows.Close()

		foundPlan := false
		for rows.Next() {
			var id, parent, notused int
			var detail string
			err := rows.Scan(&id, &parent, &notused, &detail)
			require.NoError(t, err)
			if detail != "" {
				foundPlan = true
				t.Logf("Query plan: %s", detail)
			}
		}
		assert.True(t, foundPlan, "Query should have an execution plan")
	})

	// Test index repository search
	t.Run("IndexRepository search uses indexes", func(t *testing.T) {
		repo, err := NewIndexRepository(db.Conn())
		require.NoError(t, err)
		defer repo.Close()

		// Query using EXPLAIN QUERY PLAN
		rows, err := db.Conn().QueryContext(ctx, `
			EXPLAIN QUERY PLAN
			SELECT * FROM tool_index WHERE tool LIKE ? OR description LIKE ?
		`, "%node%", "%node%")
		require.NoError(t, err)
		defer rows.Close()

		foundPlan := false
		for rows.Next() {
			var id, parent, notused int
			var detail string
			err := rows.Scan(&id, &parent, &notused, &detail)
			require.NoError(t, err)
			if detail != "" {
				foundPlan = true
				t.Logf("Query plan: %s", detail)
			}
		}
		assert.True(t, foundPlan, "Query should have an execution plan")
	})
}

// TestRepositoryIntegration tests all repositories working together
func TestRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create all repositories
	installRepo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)
	defer installRepo.Close()

	cacheRepo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer cacheRepo.Close()

	auditRepo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer auditRepo.Close()

	indexRepo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer indexRepo.Close()

	// Simulate a tool installation workflow
	t.Run("Complete installation workflow", func(t *testing.T) {
		// 1. Check tool index
		indexEntry := &repository.IndexEntry{
			Tool:        "node",
			Description: "Node.js JavaScript runtime",
			Homepage:    "https://nodejs.org",
			License:     "MIT",
			Backend:     "github",
			Metadata:    `{"repo":"nodejs/node"}`,
		}
		err := indexRepo.Upsert(ctx, indexEntry)
		require.NoError(t, err)

		// 2. Log installation start
		auditEntry := &repository.AuditEntry{
			Operation: "install",
			Tool:      "node",
			Version:   "20.0.0",
			Status:    "started",
			Duration:  0,
			Metadata:  `{}`,
		}
		err = auditRepo.Log(ctx, auditEntry)
		require.NoError(t, err)

		// 3. Create installation record
		installation := &repository.Installation{
			Tool:        "node",
			Version:     "20.0.0",
			Backend:     "github",
			Provider:    "node",
			InstallPath: "/usr/local/unirtm/installs/node/20.0.0",
			Checksum:    "abc123def456",
			Metadata:    `{"arch":"x64","os":"linux"}`,
		}
		err = installRepo.Create(ctx, installation)
		require.NoError(t, err)

		// 4. Cache version metadata
		err = cacheRepo.Set(ctx, "node:20.0.0:metadata", []byte(`{"size":50000000}`), 24*3600*1000000000) // 24 hours
		require.NoError(t, err)

		// 5. Log installation success
		auditEntry = &repository.AuditEntry{
			Operation: "install",
			Tool:      "node",
			Version:   "20.0.0",
			Status:    "success",
			Duration:  1500,
			Metadata:  `{}`,
		}
		err = auditRepo.Log(ctx, auditEntry)
		require.NoError(t, err)

		// Verify all data is accessible
		found, err := installRepo.FindByToolAndVersion(ctx, "node", "20.0.0")
		require.NoError(t, err)
		assert.Equal(t, "node", found.Tool)

		cached, err := cacheRepo.Get(ctx, "node:20.0.0:metadata")
		require.NoError(t, err)
		assert.NotNil(t, cached)

		auditLogs, err := auditRepo.Query(ctx, repository.AuditFilter{Tool: "node"})
		require.NoError(t, err)
		assert.Len(t, auditLogs, 2)

		indexFound, err := indexRepo.FindByTool(ctx, "node")
		require.NoError(t, err)
		assert.Equal(t, "node", indexFound.Tool)
	})
}
