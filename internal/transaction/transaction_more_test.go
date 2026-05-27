package transaction

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteTransactionManager_BeginError(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)

	m := NewSQLiteTransactionManager(db)
	
	// close db so BeginTx fails
	db.Close()

	_, err = m.Begin(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction:")
}
