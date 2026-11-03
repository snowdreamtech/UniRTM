// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite" // Use pure-Go SQLite driver (same as production code) to avoid CGO compilation overhead in CI
)

func TestSQLiteTransactionManager_BeginError(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)

	m := NewSQLiteTransactionManager(db)

	// close db so BeginTx fails
	db.Close()

	_, err = m.Begin(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction:")
}

func TestSQLiteTransactionManager_BeginRepoErrors(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)
	defer db.Close()

	m := NewSQLiteTransactionManager(db)
	errMock := fmt.Errorf("mock error")

	origInst := newInstallationRepo
	origCache := newCacheRepo
	origAudit := newAuditRepo
	origIndex := newIndexRepo
	defer func() {
		newInstallationRepo = origInst
		newCacheRepo = origCache
		newAuditRepo = origAudit
		newIndexRepo = origIndex
	}()

	mockSuccessInst := func(db sqlite.DBExecutor) (*sqlite.InstallationRepository, error) { return nil, nil }
	mockSuccessCache := func(db sqlite.DBExecutor) (*sqlite.CacheRepository, error) { return nil, nil }
	mockSuccessAudit := func(db sqlite.DBExecutor) (*sqlite.AuditRepository, error) { return nil, nil }
	mockSuccessIndex := func(db sqlite.DBExecutor) (*sqlite.IndexRepository, error) { return nil, nil }

	t.Run("installation repo error", func(t *testing.T) {
		newInstallationRepo = func(db sqlite.DBExecutor) (*sqlite.InstallationRepository, error) { return nil, errMock }
		newCacheRepo = mockSuccessCache
		newAuditRepo = mockSuccessAudit
		newIndexRepo = mockSuccessIndex

		_, err := m.Begin(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create installation repository:")
	})

	t.Run("cache repo error", func(t *testing.T) {
		newInstallationRepo = mockSuccessInst
		newCacheRepo = func(db sqlite.DBExecutor) (*sqlite.CacheRepository, error) { return nil, errMock }
		newAuditRepo = mockSuccessAudit
		newIndexRepo = mockSuccessIndex

		_, err := m.Begin(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create cache repository:")
	})

	t.Run("audit repo error", func(t *testing.T) {
		newInstallationRepo = mockSuccessInst
		newCacheRepo = mockSuccessCache
		newAuditRepo = func(db sqlite.DBExecutor) (*sqlite.AuditRepository, error) { return nil, errMock }
		newIndexRepo = mockSuccessIndex

		_, err := m.Begin(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create audit repository:")
	})

	t.Run("index repo error", func(t *testing.T) {
		newInstallationRepo = mockSuccessInst
		newCacheRepo = mockSuccessCache
		newAuditRepo = mockSuccessAudit
		newIndexRepo = func(db sqlite.DBExecutor) (*sqlite.IndexRepository, error) { return nil, errMock }

		_, err := m.Begin(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create index repository:")
	})
}
