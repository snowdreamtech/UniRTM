// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

// Schema contains the SQL schema definitions for all tables
const Schema = `
-- installations table
CREATE TABLE IF NOT EXISTS installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool TEXT NOT NULL,
    version TEXT NOT NULL,
    backend TEXT NOT NULL,
    provider TEXT NOT NULL,
    install_path TEXT NOT NULL,
    checksum TEXT NOT NULL,
    installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT,
    UNIQUE(tool, version)
);

CREATE INDEX IF NOT EXISTS idx_installations_tool ON installations(tool);
CREATE INDEX IF NOT EXISTS idx_installations_installed_at ON installations(installed_at);

-- cache table
CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value BLOB NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cache_expires_at ON cache(expires_at);

-- audit_log table
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    operation TEXT NOT NULL,
    tool TEXT,
    version TEXT,
    status TEXT NOT NULL,
    error TEXT,
    duration_ms INTEGER,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_operation ON audit_log(operation);
CREATE INDEX IF NOT EXISTS idx_audit_log_tool ON audit_log(tool);

-- tool_index table
CREATE TABLE IF NOT EXISTS tool_index (
    tool TEXT PRIMARY KEY,
    description TEXT,
    homepage TEXT,
    license TEXT,
    backend TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_tool_index_backend ON tool_index(backend);
CREATE INDEX IF NOT EXISTS idx_tool_index_updated_at ON tool_index(updated_at);

-- schema_migrations table for tracking migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    description TEXT NOT NULL
);
`

// CurrentSchemaVersion is the current schema version
const CurrentSchemaVersion = 2

// migrations contains all schema migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema",
		Up:          Schema,
		Down:        "", // Initial schema has no down migration
	},
	{
		Version:     2,
		Description: "Add gpg_verification to audit_log",
		Up:          "ALTER TABLE audit_log ADD COLUMN gpg_verification TEXT;",
		Down:        "", // No down migration for adding a column in SQLite easily
	},
}
