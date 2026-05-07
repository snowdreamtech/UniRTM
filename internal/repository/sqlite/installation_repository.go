package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// InstallationRepository implements repository.InstallationRepository for SQLite
// Validates Requirements: 2.2 (Store installation cache data)
type InstallationRepository struct {
	db DBExecutor

	// Prepared statements for performance
	createStmt               *sql.Stmt
	findByToolAndVersionStmt *sql.Stmt
	listStmt                 *sql.Stmt
	deleteStmt               *sql.Stmt
}

// NewInstallationRepository creates a new SQLite installation repository
func NewInstallationRepository(db DBExecutor) (*InstallationRepository, error) {
	repo := &InstallationRepository{db: db}

	// Prepare statements
	var err error

	repo.createStmt, err = db.Prepare(`
		INSERT INTO installations (tool, version, backend, provider, install_path, checksum, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare create statement: %w", err)
	}

	repo.findByToolAndVersionStmt, err = db.Prepare(`
		SELECT id, tool, version, backend, provider, install_path, checksum, installed_at, metadata
		FROM installations
		WHERE tool = ? AND version = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare find statement: %w", err)
	}

	repo.listStmt, err = db.Prepare(`
		SELECT id, tool, version, backend, provider, install_path, checksum, installed_at, metadata
		FROM installations
		ORDER BY installed_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare list statement: %w", err)
	}

	repo.deleteStmt, err = db.Prepare(`
		DELETE FROM installations
		WHERE tool = ? AND version = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare delete statement: %w", err)
	}

	return repo, nil
}

// Create records a new installation
func (r *InstallationRepository) Create(ctx context.Context, installation *repository.Installation) error {
	result, err := r.createStmt.ExecContext(
		ctx,
		installation.Tool,
		installation.Version,
		installation.Backend,
		installation.Provider,
		installation.InstallPath,
		installation.Checksum,
		installation.Metadata,
	)
	if err != nil {
		// Check for unique constraint violation
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return fmt.Errorf("installation already exists for %s@%s: %w", installation.Tool, installation.Version, repository.ErrAlreadyExists)
			}
		}
		return fmt.Errorf("insert installation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	installation.ID = id
	return nil
}

// FindByToolAndVersion finds an installation by tool and version
func (r *InstallationRepository) FindByToolAndVersion(ctx context.Context, tool string, version string) (*repository.Installation, error) {
	installation := &repository.Installation{}

	err := r.findByToolAndVersionStmt.QueryRowContext(ctx, tool, version).Scan(
		&installation.ID,
		&installation.Tool,
		&installation.Version,
		&installation.Backend,
		&installation.Provider,
		&installation.InstallPath,
		&installation.Checksum,
		&installation.InstalledAt,
		&installation.Metadata,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("installation not found for %s@%s: %w", tool, version, repository.ErrNotFound)
		}
		return nil, fmt.Errorf("query installation: %w", err)
	}

	return installation, nil
}

// List lists all installations
func (r *InstallationRepository) List(ctx context.Context) ([]*repository.Installation, error) {
	rows, err := r.listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query installations: %w", err)
	}
	defer rows.Close()

	installations := []*repository.Installation{}
	for rows.Next() {
		installation := &repository.Installation{}
		err := rows.Scan(
			&installation.ID,
			&installation.Tool,
			&installation.Version,
			&installation.Backend,
			&installation.Provider,
			&installation.InstallPath,
			&installation.Checksum,
			&installation.InstalledAt,
			&installation.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan installation: %w", err)
		}
		installations = append(installations, installation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate installations: %w", err)
	}

	return installations, nil
}

// Delete removes an installation record
func (r *InstallationRepository) Delete(ctx context.Context, tool string, version string) error {
	result, err := r.deleteStmt.ExecContext(ctx, tool, version)
	if err != nil {
		return fmt.Errorf("delete installation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("installation not found for %s@%s: %w", tool, version, repository.ErrNotFound)
	}

	return nil
}

// Close closes all prepared statements
func (r *InstallationRepository) Close() error {
	var errs []error

	if err := r.createStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close create statement: %w", err))
	}
	if err := r.findByToolAndVersionStmt.Close(); err != nil {
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
