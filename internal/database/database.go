// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // SQLite driver
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// Config contains database configuration
type Config struct {
	// Path is the path to the SQLite database file
	Path string

	// WALMode enables Write-Ahead Logging for better concurrent read performance
	WALMode bool
}

// Open opens a connection to the SQLite database and initializes the schema
func Open(ctx context.Context, config Config) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Open the database connection with busy_timeout and WAL mode
	// _txlock=immediate is required for concurrent writes to prevent SQLITE_BUSY deadlocks
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_txlock=immediate", config.Path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	// We allow multiple connections for better concurrency with WAL mode.
	// SQLite will handle write serialization internally.
	conn.SetMaxOpenConns(100)
	conn.SetMaxIdleConns(10)

	db := &DB{
		conn: conn,
		path: config.Path,
	}

	// Always enable WAL mode for better concurrency
	if err := db.enableWALMode(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	// Initialize schema and run migrations
	if err := db.initialize(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	return db, nil
}

// enableWALMode enables Write-Ahead Logging mode for better concurrent read performance
func (db *DB) enableWALMode(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("set WAL mode: %w", err)
	}
	return nil
}

// initialize initializes the database schema and applies migrations
func (db *DB) initialize(ctx context.Context) error {
	// Create migration manager
	migrationMgr := NewMigrationManager(db.conn)

	// Apply all pending migrations
	if err := migrationMgr.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, opts)
}

// Ping verifies the database connection is alive
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// GetSchemaVersion returns the current schema version
func (db *DB) GetSchemaVersion(ctx context.Context) (int, error) {
	migrationMgr := NewMigrationManager(db.conn)
	return migrationMgr.GetCurrentVersion(ctx)
}
