package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditRepository_Log(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	entry := &repository.AuditEntry{
		Operation: "install",
		Tool:      "node",
		Version:   "20.0.0",
		Status:    "success",
		Error:     "",
		Duration:  1500,
		Metadata:  `{"backend":"github"}`,
	}

	err = repo.Log(ctx, entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)
}

func TestAuditRepository_Query_All(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create multiple audit entries
	entries := []*repository.AuditEntry{
		{
			Operation: "install",
			Tool:      "node",
			Version:   "18.0.0",
			Status:    "success",
			Duration:  1000,
			Metadata:  `{}`,
		},
		{
			Operation: "install",
			Tool:      "python",
			Version:   "3.11.0",
			Status:    "success",
			Duration:  2000,
			Metadata:  `{}`,
		},
		{
			Operation: "uninstall",
			Tool:      "node",
			Version:   "16.0.0",
			Status:    "success",
			Duration:  500,
			Metadata:  `{}`,
		},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Query all entries
	filter := repository.AuditFilter{}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify order (most recent first)
	assert.Equal(t, "uninstall", results[0].Operation)
	assert.Equal(t, "python", results[1].Tool)
	assert.Equal(t, "node", results[2].Tool)
}

func TestAuditRepository_Query_ByOperation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entries with different operations
	entries := []*repository.AuditEntry{
		{Operation: "install", Tool: "node", Version: "18.0.0", Status: "success", Duration: 1000, Metadata: `{}`},
		{Operation: "install", Tool: "python", Version: "3.11.0", Status: "success", Duration: 2000, Metadata: `{}`},
		{Operation: "uninstall", Tool: "node", Version: "16.0.0", Status: "success", Duration: 500, Metadata: `{}`},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
	}

	// Query by operation
	filter := repository.AuditFilter{
		Operation: "install",
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "install", results[0].Operation)
	assert.Equal(t, "install", results[1].Operation)
}

func TestAuditRepository_Query_ByTool(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entries for different tools
	entries := []*repository.AuditEntry{
		{Operation: "install", Tool: "node", Version: "18.0.0", Status: "success", Duration: 1000, Metadata: `{}`},
		{Operation: "install", Tool: "node", Version: "20.0.0", Status: "success", Duration: 1500, Metadata: `{}`},
		{Operation: "install", Tool: "python", Version: "3.11.0", Status: "success", Duration: 2000, Metadata: `{}`},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
	}

	// Query by tool
	filter := repository.AuditFilter{
		Tool: "node",
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "node", results[0].Tool)
	assert.Equal(t, "node", results[1].Tool)
}

func TestAuditRepository_Query_ByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entries with different statuses
	entries := []*repository.AuditEntry{
		{Operation: "install", Tool: "node", Version: "18.0.0", Status: "success", Duration: 1000, Metadata: `{}`},
		{Operation: "install", Tool: "python", Version: "3.11.0", Status: "failure", Error: "network error", Duration: 500, Metadata: `{}`},
		{Operation: "install", Tool: "go", Version: "1.21.0", Status: "success", Duration: 1500, Metadata: `{}`},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
	}

	// Query by status
	filter := repository.AuditFilter{
		Status: "failure",
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "failure", results[0].Status)
	assert.Equal(t, "python", results[0].Tool)
	assert.Equal(t, "network error", results[0].Error)
}

func TestAuditRepository_Query_ByTimeRange(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entries
	entries := []*repository.AuditEntry{
		{Operation: "install", Tool: "node", Version: "18.0.0", Status: "success", Duration: 1000, Metadata: `{}`},
		{Operation: "install", Tool: "python", Version: "3.11.0", Status: "success", Duration: 2000, Metadata: `{}`},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
	}

	// Query by time range - use a very wide range to ensure we capture the entries
	// SQLite uses UTC for CURRENT_TIMESTAMP
	startTime := time.Now().UTC().Add(-24 * time.Hour)
	endTime := time.Now().UTC().Add(24 * time.Hour)

	filter := repository.AuditFilter{
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestAuditRepository_Query_WithPagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create multiple entries
	for i := 0; i < 10; i++ {
		entry := &repository.AuditEntry{
			Operation: "install",
			Tool:      "node",
			Version:   "18.0.0",
			Status:    "success",
			Duration:  1000,
			Metadata:  `{}`,
		}
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
		time.Sleep(5 * time.Millisecond)
	}

	// Query with limit
	filter := repository.AuditFilter{
		Limit: 5,
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 5)

	// Query with offset
	filter = repository.AuditFilter{
		Limit:  5,
		Offset: 5,
	}
	results, err = repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestAuditRepository_Query_MultipleFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create diverse entries
	entries := []*repository.AuditEntry{
		{Operation: "install", Tool: "node", Version: "18.0.0", Status: "success", Duration: 1000, Metadata: `{}`},
		{Operation: "install", Tool: "node", Version: "20.0.0", Status: "failure", Error: "error", Duration: 500, Metadata: `{}`},
		{Operation: "uninstall", Tool: "node", Version: "16.0.0", Status: "success", Duration: 300, Metadata: `{}`},
		{Operation: "install", Tool: "python", Version: "3.11.0", Status: "success", Duration: 2000, Metadata: `{}`},
	}

	for _, entry := range entries {
		err := repo.Log(ctx, entry)
		require.NoError(t, err)
	}

	// Query with multiple filters
	filter := repository.AuditFilter{
		Operation: "install",
		Tool:      "node",
		Status:    "success",
	}
	results, err := repo.Query(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "install", results[0].Operation)
	assert.Equal(t, "node", results[0].Tool)
	assert.Equal(t, "success", results[0].Status)
	assert.Equal(t, "18.0.0", results[0].Version)
}
