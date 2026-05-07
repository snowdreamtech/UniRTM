// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CacheRepository implements repository.CacheRepository for SQLite
// Validates Requirements: 2.2 (Store installation cache data)
type CacheRepository struct {
	db DBExecutor

	// Prepared statements for performance
	setStmt    *sql.Stmt
	getStmt    *sql.Stmt
	deleteStmt *sql.Stmt
	purgeStmt  *sql.Stmt
}

// NewCacheRepository creates a new SQLite cache repository
func NewCacheRepository(db DBExecutor) (*CacheRepository, error) {
	repo := &CacheRepository{db: db}

	// Prepare statements
	var err error

	// Use INSERT OR REPLACE for upsert behavior
	repo.setStmt, err = db.Prepare(`
		INSERT OR REPLACE INTO cache (key, value, expires_at, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare set statement: %w", err)
	}

	// Get only non-expired entries
	repo.getStmt, err = db.Prepare(`
		SELECT value, expires_at
		FROM cache
		WHERE key = ? AND expires_at > CURRENT_TIMESTAMP
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare get statement: %w", err)
	}

	repo.deleteStmt, err = db.Prepare(`
		DELETE FROM cache
		WHERE key = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare delete statement: %w", err)
	}

	repo.purgeStmt, err = db.Prepare(`
		DELETE FROM cache
		WHERE expires_at <= CURRENT_TIMESTAMP
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare purge statement: %w", err)
	}

	return repo, nil
}

// Set stores a cache entry with the specified TTL
func (r *CacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	_, err := r.setStmt.ExecContext(ctx, key, value, expiresAt)
	if err != nil {
		return fmt.Errorf("insert cache entry: %w", err)
	}

	return nil
}

// Get retrieves a cache entry
// Returns nil if the key does not exist or has expired
func (r *CacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	var expiresAt time.Time

	err := r.getStmt.QueryRowContext(ctx, key).Scan(&value, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Key not found or expired
			return nil, nil
		}
		return nil, fmt.Errorf("query cache entry: %w", err)
	}

	// Double-check expiration (should be handled by query, but defensive)
	if time.Now().After(expiresAt) {
		return nil, nil
	}

	return value, nil
}

// Delete removes a cache entry
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	_, err := r.deleteStmt.ExecContext(ctx, key)
	if err != nil {
		return fmt.Errorf("delete cache entry: %w", err)
	}

	return nil
}

// Purge removes all expired cache entries
func (r *CacheRepository) Purge(ctx context.Context) error {
	result, err := r.purgeStmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("purge expired cache entries: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	// Log the number of purged entries (optional, for debugging)
	_ = rowsAffected

	return nil
}

// Close closes all prepared statements
func (r *CacheRepository) Close() error {
	var errs []error

	if err := r.setStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close set statement: %w", err))
	}
	if err := r.getStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close get statement: %w", err))
	}
	if err := r.deleteStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close delete statement: %w", err))
	}
	if err := r.purgeStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close purge statement: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("close statements: %v", errs)
	}

	return nil
}
