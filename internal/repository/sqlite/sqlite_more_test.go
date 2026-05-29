// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite" // Use pure-Go SQLite driver (same as production code) to avoid CGO compilation overhead in CI
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
)

func getClosedDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)
	// create table so prepare statement doesn't fail immediately in some constructors if they do table checking
	// actually constructors don't create tables in SQLite backend usually, or do they?
	// The constructor prepares statements, so the tables MUST exist before prepare.
	// Oh wait! If the tables don't exist, Prepare will fail during NewXXXRepository!
	return db
}

func getPreparedDB(t *testing.T, initSQL string) *sql.DB {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	assert.NoError(t, err)
	_, err = db.Exec(initSQL)
	assert.NoError(t, err)
	return db
}

const auditSQL = `
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation TEXT NOT NULL,
    tool TEXT NOT NULL,
    version TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    duration_ms INTEGER NOT NULL,
    gpg_verification BOOLEAN NOT NULL DEFAULT 0,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func TestAuditRepository_ClosedDB(t *testing.T) {
	db := getPreparedDB(t, auditSQL)
	r, err := NewAuditRepository(db)
	assert.NoError(t, err)

	err = r.Close()
	assert.NoError(t, err)
	db.Close() // close the underlying db

	// these should error
	err = r.Log(context.Background(), &repository.AuditEntry{})
	assert.Error(t, err)

	_, err = r.Query(context.Background(), repository.AuditFilter{})
	assert.Error(t, err)
}

const cacheSQL = `
CREATE TABLE cache (
    key TEXT PRIMARY KEY,
    value BLOB NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

func TestCacheRepository_ClosedDB(t *testing.T) {
	db := getPreparedDB(t, cacheSQL)
	r, err := NewCacheRepository(db)
	assert.NoError(t, err)
	if r == nil {
		return
	}

	err = r.Close()
	assert.NoError(t, err)
	db.Close() // close underlying db

	err = r.Set(context.Background(), "k", []byte("v"), time.Hour)
	assert.Error(t, err)

	_, err = r.Get(context.Background(), "k")
	assert.Error(t, err)

	err = r.Delete(context.Background(), "k")
	assert.Error(t, err)

	err = r.Purge(context.Background())
	assert.Error(t, err)
}

const indexSQL = `
CREATE TABLE tool_index (
    tool TEXT PRIMARY KEY,
    description TEXT,
    homepage TEXT,
    license TEXT,
    backend TEXT,
    metadata TEXT,
    updated_at DATETIME NOT NULL
);`

func TestIndexRepository_ClosedDB(t *testing.T) {
	db := getPreparedDB(t, indexSQL)
	r, err := NewIndexRepository(db)
	assert.NoError(t, err)
	if r == nil {
		return
	}

	err = r.Close()
	assert.NoError(t, err)
	db.Close()

	err = r.Upsert(context.Background(), &repository.IndexEntry{})
	assert.Error(t, err)

	_, err = r.FindByTool(context.Background(), "t")
	assert.Error(t, err)

	_, err = r.List(context.Background())
	assert.Error(t, err)

	_, err = r.Search(context.Background(), "t")
	assert.Error(t, err)

	err = r.Delete(context.Background(), "t")
	assert.Error(t, err)
}

const installSQL = `
CREATE TABLE installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool TEXT NOT NULL,
    version TEXT NOT NULL,
    install_path TEXT NOT NULL,
    checksum TEXT,
    backend TEXT,
    provider TEXT,
    metadata TEXT,
    installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    UNIQUE (tool, version)
);`

func TestInstallationRepository_ClosedDB(t *testing.T) {
	db := getPreparedDB(t, installSQL)
	r, err := NewInstallationRepository(db)
	assert.NoError(t, err)
	if r == nil {
		return
	}

	err = r.Close()
	assert.NoError(t, err)
	db.Close()

	err = r.Create(context.Background(), &repository.Installation{})
	assert.Error(t, err)

	err = r.Upsert(context.Background(), &repository.Installation{})
	assert.Error(t, err)

	_, err = r.FindByToolAndVersion(context.Background(), "t", "v")
	assert.Error(t, err)

	_, err = r.List(context.Background())
	assert.Error(t, err)

	_, err = r.ListByTool(context.Background(), "t")
	assert.Error(t, err)

	err = r.Delete(context.Background(), "t", "v")
	assert.Error(t, err)
}
