package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

// IndexRepository implements repository.IndexRepository for SQLite
// Validates Requirements: 2.4 (Store tool indexes)
type IndexRepository struct {
	db DBExecutor

	// Prepared statements for performance
	upsertStmt     *sql.Stmt
	findByToolStmt *sql.Stmt
	listStmt       *sql.Stmt
	deleteStmt     *sql.Stmt
}

// NewIndexRepository creates a new SQLite index repository
func NewIndexRepository(db DBExecutor) (*IndexRepository, error) {
	repo := &IndexRepository{db: db}

	// Prepare statements
	var err error

	// Use INSERT OR REPLACE for upsert behavior
	repo.upsertStmt, err = db.Prepare(`
		INSERT OR REPLACE INTO tool_index (tool, description, homepage, license, backend, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare upsert statement: %w", err)
	}

	repo.findByToolStmt, err = db.Prepare(`
		SELECT tool, description, homepage, license, backend, updated_at, metadata
		FROM tool_index
		WHERE tool = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare find statement: %w", err)
	}

	repo.listStmt, err = db.Prepare(`
		SELECT tool, description, homepage, license, backend, updated_at, metadata
		FROM tool_index
		ORDER BY tool ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare list statement: %w", err)
	}

	repo.deleteStmt, err = db.Prepare(`
		DELETE FROM tool_index
		WHERE tool = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare delete statement: %w", err)
	}

	return repo, nil
}

// Upsert creates or updates a tool index entry
func (r *IndexRepository) Upsert(ctx context.Context, entry *repository.IndexEntry) error {
	_, err := r.upsertStmt.ExecContext(
		ctx,
		entry.Tool,
		entry.Description,
		entry.Homepage,
		entry.License,
		entry.Backend,
		entry.Metadata,
	)
	if err != nil {
		return fmt.Errorf("upsert tool index: %w", err)
	}

	return nil
}

// FindByTool finds a tool index entry by tool name
func (r *IndexRepository) FindByTool(ctx context.Context, tool string) (*repository.IndexEntry, error) {
	entry := &repository.IndexEntry{}

	err := r.findByToolStmt.QueryRowContext(ctx, tool).Scan(
		&entry.Tool,
		&entry.Description,
		&entry.Homepage,
		&entry.License,
		&entry.Backend,
		&entry.UpdatedAt,
		&entry.Metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tool index not found for %s: %w", tool, repository.ErrNotFound)
		}
		return nil, fmt.Errorf("query tool index: %w", err)
	}

	return entry, nil
}

// List lists all tool index entries
func (r *IndexRepository) List(ctx context.Context) ([]*repository.IndexEntry, error) {
	rows, err := r.listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query tool index: %w", err)
	}
	defer rows.Close()

	entries := []*repository.IndexEntry{}
	for rows.Next() {
		entry := &repository.IndexEntry{}
		err := rows.Scan(
			&entry.Tool,
			&entry.Description,
			&entry.Homepage,
			&entry.License,
			&entry.Backend,
			&entry.UpdatedAt,
			&entry.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tool index: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tool index: %w", err)
	}

	return entries, nil
}

// Search searches tool index by name, description, or tags
// The query string is matched against tool name, description, and metadata
func (r *IndexRepository) Search(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
	// Use LIKE for case-insensitive search
	// Search in tool name, description, and metadata
	searchQuery := `
		SELECT tool, description, homepage, license, backend, updated_at, metadata
		FROM tool_index
		WHERE tool LIKE ? OR description LIKE ? OR metadata LIKE ?
		ORDER BY tool ASC
	`

	// Prepare search pattern with wildcards
	pattern := "%" + query + "%"

	rows, err := r.db.QueryContext(ctx, searchQuery, pattern, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("search tool index: %w", err)
	}
	defer rows.Close()

	entries := []*repository.IndexEntry{}
	for rows.Next() {
		entry := &repository.IndexEntry{}
		err := rows.Scan(
			&entry.Tool,
			&entry.Description,
			&entry.Homepage,
			&entry.License,
			&entry.Backend,
			&entry.UpdatedAt,
			&entry.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tool index: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tool index: %w", err)
	}

	return entries, nil
}

// Delete removes a tool index entry
func (r *IndexRepository) Delete(ctx context.Context, tool string) error {
	result, err := r.deleteStmt.ExecContext(ctx, tool)
	if err != nil {
		return fmt.Errorf("delete tool index: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool index not found for %s: %w", tool, repository.ErrNotFound)
	}

	return nil
}

// Close closes all prepared statements
func (r *IndexRepository) Close() error {
	var errs []error

	if err := r.upsertStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close upsert statement: %w", err))
	}
	if err := r.findByToolStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close find statement: %w", err))
	}
	if err := r.listStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close list statement: %w", err))
	}
	if err := r.deleteStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close delete statement: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("close statements: %v", errs)
	}

	return nil
}
