package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestRunIndexStatusAndClear(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	os.MkdirAll(filepath.Dir(env.GetDatabasePath()), 0755)

	// Pre-populate some dummy entries
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{
		Path:    env.GetDatabasePath(),
		WALMode: true,
	})
	assert.NoError(t, err)

	db.Conn().ExecContext(ctx, "INSERT INTO tool_index (tool, backend, homepage, license, description, metadata, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"dummy", "native", "", "", "A dummy tool", "{}", time.Now().Format(time.RFC3339))
	db.Close()

	// Test status
	err = runIndexStatus(indexStatusCmd, []string{})
	assert.NoError(t, err)

	// Test clear
	err = runIndexClear(indexClearCmd, []string{})
	assert.NoError(t, err)

	// Verify clear
	db2, _ := database.Open(ctx, database.Config{Path: env.GetDatabasePath(), WALMode: true})
	indexRepo2, _ := sqlite.NewIndexRepository(db2.Conn())
	entries, _ := indexRepo2.List(ctx)
	assert.Empty(t, entries)
	db2.Close()

	// Test status empty
	err = runIndexStatus(indexStatusCmd, []string{})
	assert.NoError(t, err)
}
