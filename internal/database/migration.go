// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Migration represents a database schema migration
type Migration struct {
	Version     int
	Description string
	Up          string // SQL to apply the migration
	Down        string // SQL to rollback the migration
}

// MigrationManager handles database schema migrations
type MigrationManager struct {
	db *sql.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// GetCurrentVersion returns the current schema version from the database
func (m *MigrationManager) GetCurrentVersion(ctx context.Context) (int, error) {
	// Check if schema_migrations table exists
	var tableExists bool
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`
	err := m.db.QueryRowContext(ctx, query).Scan(&tableExists)
	if err != nil {
		return 0, fmt.Errorf("check schema_migrations table: %w", err)
	}

	if !tableExists {
		return 0, nil
	}

	// Get the latest version
	var version int
	query = `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	err = m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("get current version: %w", err)
	}

	return version, nil
}

// ApplyMigrations applies all pending migrations
func (m *MigrationManager) ApplyMigrations(ctx context.Context) error {
	currentVersion, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	// Find migrations to apply
	pendingMigrations := []Migration{}
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		return nil // No migrations to apply
	}

	// Apply each migration in a transaction
	for _, migration := range pendingMigrations {
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

// applyMigration applies a single migration in a transaction
func (m *MigrationManager) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the migration SQL
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		return fmt.Errorf("execute migration SQL: %w", err)
	}

	// Record the migration
	query := `INSERT INTO schema_migrations (version, description) VALUES (?, ?)`
	if _, err := tx.ExecContext(ctx, query, migration.Version, migration.Description); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the last applied migration
func (m *MigrationManager) Rollback(ctx context.Context) error {
	currentVersion, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the migration to rollback
	var targetMigration *Migration
	for i := range migrations {
		if migrations[i].Version == currentVersion {
			targetMigration = &migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found", currentVersion)
	}

	if targetMigration.Down == "" {
		return fmt.Errorf("migration %d has no down migration", currentVersion)
	}

	// Rollback in a transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute the rollback SQL
	if _, err := tx.ExecContext(ctx, targetMigration.Down); err != nil {
		return fmt.Errorf("execute rollback SQL: %w", err)
	}

	// Remove the migration record
	query := `DELETE FROM schema_migrations WHERE version = ?`
	if _, err := tx.ExecContext(ctx, query, currentVersion); err != nil {
		return fmt.Errorf("remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
