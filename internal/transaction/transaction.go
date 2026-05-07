// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
)

// TransactionManager manages database transactions
// Validates Requirements: 2.8 (Use transactions for all write operations)
type TransactionManager interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) (Transaction, error)
}

// Transaction represents an active database transaction with repository access
// Validates Requirements: 2.8 (Use transactions for all write operations), 3.3 (Support explicit commit operations)
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// InstallationRepo returns the installation repository for this transaction
	InstallationRepo() repository.InstallationRepository

	// CacheRepo returns the cache repository for this transaction
	CacheRepo() repository.CacheRepository

	// AuditRepo returns the audit repository for this transaction
	AuditRepo() repository.AuditRepository

	// IndexRepo returns the index repository for this transaction
	IndexRepo() repository.IndexRepository
}

// sqliteTransactionManager implements TransactionManager for SQLite
type sqliteTransactionManager struct {
	db *sql.DB
}

// NewSQLiteTransactionManager creates a new SQLite transaction manager
func NewSQLiteTransactionManager(db *sql.DB) TransactionManager {
	return &sqliteTransactionManager{
		db: db,
	}
}

// Begin starts a new transaction
func (m *sqliteTransactionManager) Begin(ctx context.Context) (Transaction, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	// Create transaction-scoped repositories
	installationRepo, err := sqlite.NewInstallationRepository(tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create installation repository: %w", err)
	}

	cacheRepo, err := sqlite.NewCacheRepository(tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create cache repository: %w", err)
	}

	auditRepo, err := sqlite.NewAuditRepository(tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create audit repository: %w", err)
	}

	indexRepo, err := sqlite.NewIndexRepository(tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create index repository: %w", err)
	}

	return &sqliteTransaction{
		tx:               tx,
		installationRepo: installationRepo,
		cacheRepo:        cacheRepo,
		auditRepo:        auditRepo,
		indexRepo:        indexRepo,
	}, nil
}

// sqliteTransaction implements Transaction for SQLite
type sqliteTransaction struct {
	tx               *sql.Tx
	installationRepo repository.InstallationRepository
	cacheRepo        repository.CacheRepository
	auditRepo        repository.AuditRepository
	indexRepo        repository.IndexRepository
}

// Commit commits the transaction
func (t *sqliteTransaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction
func (t *sqliteTransaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}
	return nil
}

// InstallationRepo returns the installation repository for this transaction
func (t *sqliteTransaction) InstallationRepo() repository.InstallationRepository {
	return t.installationRepo
}

// CacheRepo returns the cache repository for this transaction
func (t *sqliteTransaction) CacheRepo() repository.CacheRepository {
	return t.cacheRepo
}

// AuditRepo returns the audit repository for this transaction
func (t *sqliteTransaction) AuditRepo() repository.AuditRepository {
	return t.auditRepo
}

// IndexRepo returns the index repository for this transaction
func (t *sqliteTransaction) IndexRepo() repository.IndexRepository {
	return t.indexRepo
}
