// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package repository

import (
	"context"
	"errors"
	"time"
)

// Common repository errors
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")
)

// Installation represents an installed tool
type Installation struct {
	ID          int64     `db:"id"`
	Tool        string    `db:"tool"`
	Version     string    `db:"version"`
	Backend     string    `db:"backend"`
	Provider    string    `db:"provider"`
	InstallPath string    `db:"install_path"`
	Checksum    string    `db:"checksum"`
	InstalledAt time.Time `db:"installed_at"`
	Metadata    string    `db:"metadata"` // JSON-encoded metadata
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string    `db:"key"`
	Value     []byte    `db:"value"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID              int64     `db:"id"`
	Timestamp       time.Time `db:"timestamp"`
	Operation       string    `db:"operation"` // install, uninstall, activate, etc.
	Tool            string    `db:"tool"`
	Version         string    `db:"version"`
	Status          string    `db:"status"` // success, failure
	Error           string    `db:"error"`
	Duration        int64     `db:"duration_ms"`
	GpgVerification string    `db:"gpg_verification"`
	Metadata        string    `db:"metadata"` // JSON-encoded metadata
}

// IndexEntry represents a tool in the index
type IndexEntry struct {
	Tool        string    `db:"tool"`
	Description string    `db:"description"`
	Homepage    string    `db:"homepage"`
	License     string    `db:"license"`
	Backend     string    `db:"backend"`
	UpdatedAt   time.Time `db:"updated_at"`
	Metadata    string    `db:"metadata"` // JSON-encoded metadata
}

// AuditFilter defines filters for querying audit logs
type AuditFilter struct {
	// StartTime filters audit entries after this time (inclusive)
	StartTime *time.Time
	// EndTime filters audit entries before this time (inclusive)
	EndTime *time.Time
	// Operation filters by operation type (install, uninstall, activate, etc.)
	Operation string
	// Tool filters by tool name
	Tool string
	// Status filters by status (success, failure)
	Status string
	// Limit limits the number of results returned (0 = no limit)
	Limit int
	// Offset skips the first N results (for pagination)
	Offset int
}

// InstallationRepository manages installation records in the database
// Validates Requirements: 2.2 (Store installation cache data)
type InstallationRepository interface {
	// Create records a new installation
	Create(ctx context.Context, installation *Installation) error

	// Upsert creates or updates an installation record
	Upsert(ctx context.Context, installation *Installation) error

	// FindByToolAndVersion finds an installation by tool and version
	FindByToolAndVersion(ctx context.Context, tool string, version string) (*Installation, error)

	// ListByTool lists all installations for a specific tool
	ListByTool(ctx context.Context, tool string) ([]*Installation, error)

	// List lists all installations
	List(ctx context.Context) ([]*Installation, error)

	// Delete removes an installation record
	Delete(ctx context.Context, tool string, version string) error
}

// CacheRepository manages cache entries with TTL support
// Validates Requirements: 2.2 (Store installation cache data)
type CacheRepository interface {
	// Set stores a cache entry with the specified TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Get retrieves a cache entry
	// Returns nil if the key does not exist or has expired
	Get(ctx context.Context, key string) ([]byte, error)

	// Delete removes a cache entry
	Delete(ctx context.Context, key string) error

	// Purge removes all expired cache entries
	Purge(ctx context.Context) error
}

// AuditRepository manages audit log entries with query filters
// Validates Requirements: 2.5 (Store audit logs)
type AuditRepository interface {
	// Log records an audit log entry
	Log(ctx context.Context, entry *AuditEntry) error

	// Query queries audit logs with filters
	Query(ctx context.Context, filter AuditFilter) ([]*AuditEntry, error)
}

// IndexRepository manages tool index entries for tool metadata
// Validates Requirements: 2.4 (Store tool indexes)
type IndexRepository interface {
	// Upsert creates or updates a tool index entry
	Upsert(ctx context.Context, entry *IndexEntry) error

	// FindByTool finds a tool index entry by tool name
	FindByTool(ctx context.Context, tool string) (*IndexEntry, error)

	// List lists all tool index entries
	List(ctx context.Context) ([]*IndexEntry, error)

	// Search searches tool index by name, description, or tags
	// The query string is matched against tool name, description, and metadata
	Search(ctx context.Context, query string) ([]*IndexEntry, error)

	// Delete removes a tool index entry
	Delete(ctx context.Context, tool string) error
}
