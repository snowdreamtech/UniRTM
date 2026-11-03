// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestNewRepositories_PrepareError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	// Close the DB immediately so that Prepare fails
	db.Close()
	cleanup()

	// NewAuditRepository
	_, err := NewAuditRepository(db.Conn())
	require.Error(t, err)
	require.Contains(t, err.Error(), "prepare log statement")

	// NewCacheRepository
	_, err = NewCacheRepository(db.Conn())
	require.Error(t, err)

	// NewIndexRepository
	_, err = NewIndexRepository(db.Conn())
	require.Error(t, err)

	// NewInstallationRepository
	_, err = NewInstallationRepository(db.Conn())
	require.Error(t, err)
}

func TestAuditRepository_CloseError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewAuditRepository(db.Conn())
	require.NoError(t, err)

	db.Close()
	err = repo.Close()
	require.NoError(t, err)
}

func TestCacheRepository_CloseError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)

	db.Close()
	_ = repo.Close()
}

func TestIndexRepository_CloseError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewIndexRepository(db.Conn())
	require.NoError(t, err)

	db.Close()
	_ = repo.Close()
}

func TestInstallationRepository_CloseError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	db.Close()
	_ = repo.Close()
}

func TestRepositories_ContextCanceled(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Audit
	auditRepo, _ := NewAuditRepository(db.Conn())
	err := auditRepo.Log(ctx, &repository.AuditEntry{})
	require.Error(t, err)

	_, err = auditRepo.Query(ctx, repository.AuditFilter{})
	require.Error(t, err)

	// Cache
	cacheRepo, _ := NewCacheRepository(db.Conn())
	err = cacheRepo.Set(ctx, "k", []byte("v"), 0)
	require.Error(t, err)

	_, err = cacheRepo.Get(ctx, "k")
	require.Error(t, err)

	err = cacheRepo.Purge(ctx)
	require.Error(t, err)

	// Index
	indexRepo, _ := NewIndexRepository(db.Conn())
	err = indexRepo.Upsert(ctx, &repository.IndexEntry{})
	require.Error(t, err)

	_, err = indexRepo.List(ctx)
	require.Error(t, err)

	_, err = indexRepo.Search(ctx, "query")
	require.Error(t, err)

	// Installation
	instRepo, _ := NewInstallationRepository(db.Conn())
	err = instRepo.Create(ctx, &repository.Installation{})
	require.Error(t, err)

	err = instRepo.Upsert(ctx, &repository.Installation{})
	require.Error(t, err)

	_, err = instRepo.List(ctx)
	require.Error(t, err)

	_, err = instRepo.ListByTool(ctx, "tool")
	require.Error(t, err)
}
