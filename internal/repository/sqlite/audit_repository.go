// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

// AuditRepository implements repository.AuditRepository for SQLite
// Validates Requirements: 2.5 (Store audit logs)
type AuditRepository struct {
	db DBExecutor

	// Prepared statements for performance
	logStmt *sql.Stmt
}

// NewAuditRepository creates a new SQLite audit repository
func NewAuditRepository(db DBExecutor) (*AuditRepository, error) {
	repo := &AuditRepository{db: db}

	// Prepare statements
	var err error

	repo.logStmt, err = db.Prepare(`
		INSERT INTO audit_log (operation, tool, version, status, error, duration_ms, gpg_verification, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare log statement: %w", err)
	}

	return repo, nil
}

// Log records an audit log entry
func (r *AuditRepository) Log(ctx context.Context, entry *repository.AuditEntry) error {
	result, err := r.logStmt.ExecContext(
		ctx,
		entry.Operation,
		entry.Tool,
		entry.Version,
		entry.Status,
		entry.Error,
		entry.Duration,
		entry.GpgVerification,
		entry.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	entry.ID = id
	return nil
}

// Query queries audit logs with filters
func (r *AuditRepository) Query(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	// Build dynamic query based on filters
	query := `
		SELECT id, timestamp, operation, tool, version, status, error, duration_ms, gpg_verification, metadata
		FROM audit_log
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply filters
	if filter.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, *filter.StartTime)
	}
	if filter.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, *filter.EndTime)
	}
	if filter.Operation != "" {
		query += " AND operation = ?"
		args = append(args, filter.Operation)
	}
	if filter.Tool != "" {
		query += " AND tool = ?"
		args = append(args, filter.Tool)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	// Order by timestamp descending (most recent first)
	query += " ORDER BY timestamp DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	entries := []*repository.AuditEntry{}
	for rows.Next() {
		entry := &repository.AuditEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Operation,
			&entry.Tool,
			&entry.Version,
			&entry.Status,
			&entry.Error,
			&entry.Duration,
			&entry.GpgVerification,
			&entry.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return entries, nil
}

// Close closes all prepared statements
func (r *AuditRepository) Close() error {
	if err := r.logStmt.Close(); err != nil {
		return fmt.Errorf("close log statement: %w", err)
	}
	return nil
}

// buildQueryWithFilters is a helper to build dynamic queries (not used in current implementation)
// Kept for reference if more complex query building is needed
func buildQueryWithFilters(baseQuery string, filters map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	for key, value := range filters {
		if value != nil {
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	return baseQuery, args
}
