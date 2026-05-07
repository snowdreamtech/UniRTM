package sqlite

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexRepository_Upsert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	entry := &repository.IndexEntry{
		Tool:        "node",
		Description: "Node.js JavaScript runtime",
		Homepage:    "https://nodejs.org",
		License:     "MIT",
		Backend:     "github",
		Metadata:    `{"repo":"nodejs/node"}`,
	}

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)
}

func TestIndexRepository_Upsert_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Insert initial entry
	entry := &repository.IndexEntry{
		Tool:        "python",
		Description: "Python programming language",
		Homepage:    "https://python.org",
		License:     "PSF",
		Backend:     "github",
		Metadata:    `{"repo":"python/cpython"}`,
	}

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)

	// Update the entry
	entry.Description = "Python 3 programming language"
	entry.Homepage = "https://www.python.org"

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)

	// Verify the update
	found, err := repo.FindByTool(ctx, "python")
	require.NoError(t, err)
	assert.Equal(t, "Python 3 programming language", found.Description)
	assert.Equal(t, "https://www.python.org", found.Homepage)
}

func TestIndexRepository_FindByTool(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entry
	entry := &repository.IndexEntry{
		Tool:        "go",
		Description: "Go programming language",
		Homepage:    "https://go.dev",
		License:     "BSD-3-Clause",
		Backend:     "github",
		Metadata:    `{"repo":"golang/go"}`,
	}

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)

	// Find the entry
	found, err := repo.FindByTool(ctx, "go")
	require.NoError(t, err)
	assert.Equal(t, entry.Tool, found.Tool)
	assert.Equal(t, entry.Description, found.Description)
	assert.Equal(t, entry.Homepage, found.Homepage)
	assert.Equal(t, entry.License, found.License)
	assert.Equal(t, entry.Backend, found.Backend)
	assert.NotZero(t, found.UpdatedAt)
}

func TestIndexRepository_FindByTool_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Try to find non-existent tool
	_, err = repo.FindByTool(ctx, "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestIndexRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create multiple entries
	entries := []*repository.IndexEntry{
		{
			Tool:        "node",
			Description: "Node.js runtime",
			Homepage:    "https://nodejs.org",
			License:     "MIT",
			Backend:     "github",
			Metadata:    `{}`,
		},
		{
			Tool:        "python",
			Description: "Python language",
			Homepage:    "https://python.org",
			License:     "PSF",
			Backend:     "github",
			Metadata:    `{}`,
		},
		{
			Tool:        "go",
			Description: "Go language",
			Homepage:    "https://go.dev",
			License:     "BSD-3-Clause",
			Backend:     "github",
			Metadata:    `{}`,
		},
	}

	for _, entry := range entries {
		err := repo.Upsert(ctx, entry)
		require.NoError(t, err)
	}

	// List all entries
	list, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Verify alphabetical order
	assert.Equal(t, "go", list[0].Tool)
	assert.Equal(t, "node", list[1].Tool)
	assert.Equal(t, "python", list[2].Tool)
}

func TestIndexRepository_Search(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entries with different descriptions
	entries := []*repository.IndexEntry{
		{
			Tool:        "node",
			Description: "Node.js JavaScript runtime",
			Homepage:    "https://nodejs.org",
			License:     "MIT",
			Backend:     "github",
			Metadata:    `{"tags":["javascript","runtime"]}`,
		},
		{
			Tool:        "deno",
			Description: "Deno JavaScript runtime",
			Homepage:    "https://deno.land",
			License:     "MIT",
			Backend:     "github",
			Metadata:    `{"tags":["javascript","typescript"]}`,
		},
		{
			Tool:        "python",
			Description: "Python programming language",
			Homepage:    "https://python.org",
			License:     "PSF",
			Backend:     "github",
			Metadata:    `{"tags":["python","language"]}`,
		},
		{
			Tool:        "go",
			Description: "Go programming language",
			Homepage:    "https://go.dev",
			License:     "BSD-3-Clause",
			Backend:     "github",
			Metadata:    `{"tags":["go","language"]}`,
		},
	}

	for _, entry := range entries {
		err := repo.Upsert(ctx, entry)
		require.NoError(t, err)
	}

	// Search by tool name
	results, err := repo.Search(ctx, "node")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "node", results[0].Tool)

	// Search by description
	results, err = repo.Search(ctx, "JavaScript")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search by metadata
	results, err = repo.Search(ctx, "typescript")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "deno", results[0].Tool)

	// Search for programming language
	results, err = repo.Search(ctx, "programming language")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search with no results
	results, err = repo.Search(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestIndexRepository_Search_CaseInsensitive(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entry
	entry := &repository.IndexEntry{
		Tool:        "Node",
		Description: "Node.js JavaScript Runtime",
		Homepage:    "https://nodejs.org",
		License:     "MIT",
		Backend:     "github",
		Metadata:    `{}`,
	}

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)

	// Search with different cases
	testCases := []string{"node", "NODE", "Node", "nOdE"}

	for _, query := range testCases {
		results, err := repo.Search(ctx, query)
		require.NoError(t, err)
		assert.Len(t, results, 1, "query: %s", query)
		assert.Equal(t, "Node", results[0].Tool)
	}
}

func TestIndexRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Create entry
	entry := &repository.IndexEntry{
		Tool:        "rust",
		Description: "Rust programming language",
		Homepage:    "https://rust-lang.org",
		License:     "MIT/Apache-2.0",
		Backend:     "github",
		Metadata:    `{}`,
	}

	err = repo.Upsert(ctx, entry)
	require.NoError(t, err)

	// Delete the entry
	err = repo.Delete(ctx, "rust")
	require.NoError(t, err)

	// Verify it's deleted
	_, err = repo.FindByTool(ctx, "rust")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestIndexRepository_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Try to delete non-existent tool
	err = repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
