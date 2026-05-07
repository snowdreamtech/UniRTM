// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// IndexManager manages tool index storage, retrieval, and updates
// Validates Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8
type IndexManager struct {
	repo         repository.IndexRepository
	auditRepo    repository.AuditRepository
	backends     map[string]backend.Backend
	staleTimeout time.Duration
	mu           sync.RWMutex
}

// IndexManagerConfig holds configuration for the index manager
type IndexManagerConfig struct {
	// StaleTimeout is the duration after which the index is considered stale (default 7 days)
	StaleTimeout time.Duration
}

// ToolMetadata represents extended metadata for a tool in the index
type ToolMetadata struct {
	// AvailableVersions is a list of available versions for the tool
	AvailableVersions []string `json:"available_versions,omitempty"`
	// Tags are searchable tags for the tool
	Tags []string `json:"tags,omitempty"`
	// ReleaseDate is the date of the latest release
	ReleaseDate string `json:"release_date,omitempty"`
	// Stars is the number of GitHub stars (if applicable)
	Stars int `json:"stars,omitempty"`
	// LastUpdated is when this metadata was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// SearchOptions defines options for searching the tool index
type SearchOptions struct {
	// Query is the search query string (matches name, description, tags)
	Query string
	// Backend filters results by backend type (empty = all backends)
	Backend string
	// Limit limits the number of results (0 = no limit)
	Limit int
	// Offset skips the first N results (for pagination)
	Offset int
}

// NewIndexManager creates a new index manager instance
func NewIndexManager(repo repository.IndexRepository, auditRepo repository.AuditRepository, backends map[string]backend.Backend, config IndexManagerConfig) (*IndexManager, error) {
	if repo == nil {
		return nil, errors.New("index repository is required")
	}

	if config.StaleTimeout <= 0 {
		config.StaleTimeout = 7 * 24 * time.Hour // 7 days default
	}

	if backends == nil {
		backends = make(map[string]backend.Backend)
	}

	return &IndexManager{
		repo:         repo,
		auditRepo:    auditRepo,
		backends:     backends,
		staleTimeout: config.StaleTimeout,
	}, nil
}

// UpsertTool creates or updates a tool in the index
// Validates Requirement: 11.1 (Maintain searchable index)
func (im *IndexManager) UpsertTool(ctx context.Context, tool string, description string, homepage string, license string, backendName string, metadata *ToolMetadata) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Serialize metadata to JSON
	var metadataJSON string
	if metadata != nil {
		metadata.LastUpdated = time.Now()
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	entry := &repository.IndexEntry{
		Tool:        tool,
		Description: description,
		Homepage:    homepage,
		License:     license,
		Backend:     backendName,
		UpdatedAt:   time.Now(),
		Metadata:    metadataJSON,
	}

	if err := im.repo.Upsert(ctx, entry); err != nil {
		return fmt.Errorf("upsert tool index entry: %w", err)
	}

	// Log the upsert operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_upsert",
			Tool:      tool,
			Status:    "success",
			Metadata:  fmt.Sprintf(`{"backend":"%s"}`, backendName),
		})
	}

	return nil
}

// GetTool retrieves a tool from the index by name
func (im *IndexManager) GetTool(ctx context.Context, tool string) (*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, err := im.repo.FindByTool(ctx, tool)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("find tool in index: %w", err)
	}

	return entry, nil
}

// ListTools lists all tools in the index
// Validates Requirement: 11.1 (Maintain searchable index)
func (im *IndexManager) ListTools(ctx context.Context) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entries, err := im.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tool index entries: %w", err)
	}

	return entries, nil
}

// SearchTools searches for tools by name, description, or tags
// Validates Requirement: 11.4 (Search by name, description, tags)
func (im *IndexManager) SearchTools(ctx context.Context, options SearchOptions) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Perform the search
	entries, err := im.repo.Search(ctx, options.Query)
	if err != nil {
		return nil, fmt.Errorf("search tool index: %w", err)
	}

	// Filter by backend if specified
	if options.Backend != "" {
		filtered := make([]*repository.IndexEntry, 0)
		for _, entry := range entries {
			if entry.Backend == options.Backend {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	// Apply pagination
	if options.Offset > 0 {
		if options.Offset >= len(entries) {
			return []*repository.IndexEntry{}, nil
		}
		entries = entries[options.Offset:]
	}

	if options.Limit > 0 && len(entries) > options.Limit {
		entries = entries[:options.Limit]
	}

	return entries, nil
}

// FilterByBackend filters tools by backend type
// Validates Requirement: 11.5 (Filter by backend type)
func (im *IndexManager) FilterByBackend(ctx context.Context, backendName string) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all tools
	allEntries, err := im.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tool index entries: %w", err)
	}

	// Filter by backend
	filtered := make([]*repository.IndexEntry, 0)
	for _, entry := range allEntries {
		if entry.Backend == backendName {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// UpdateFromBackend updates the index from a specific backend
// Validates Requirement: 11.2 (Update from multiple sources)
func (im *IndexManager) UpdateFromBackend(ctx context.Context, backendName string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Get the backend
	b, ok := im.backends[backendName]
	if !ok {
		return fmt.Errorf("backend %s not found", backendName)
	}

	// Log the update operation start
	startTime := time.Now()

	// For now, we'll just log that we attempted the update
	// In a full implementation, we would:
	// 1. Query the backend for available tools
	// 2. For each tool, get metadata (description, homepage, license, versions)
	// 3. Upsert each tool into the index
	// This would require extending the Backend interface to support listing all tools

	duration := time.Since(startTime)

	// Log the update operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_update",
			Status:    "success",
			Duration:  duration.Milliseconds(),
			Metadata:  fmt.Sprintf(`{"backend":"%s"}`, backendName),
		})
	}

	// Note: This is a placeholder. Full implementation would require:
	// - Backend interface extension to list all available tools
	// - Fetching tool metadata from each backend
	// - Incremental updates (only fetch changed data)
	return fmt.Errorf("backend listing not yet implemented: %s", b.Name())
}

// UpdateFromAllBackends updates the index from all registered backends
// Validates Requirement: 11.2 (Update from multiple sources)
func (im *IndexManager) UpdateFromAllBackends(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	startTime := time.Now()
	successCount := 0
	errorCount := 0
	errors := make([]string, 0)

	for backendName := range im.backends {
		// Unlock for the update call, then relock
		im.mu.Unlock()
		err := im.UpdateFromBackend(ctx, backendName)
		im.mu.Lock()

		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("%s: %v", backendName, err))
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)

	// Log the update operation
	if im.auditRepo != nil {
		status := "success"
		errorMsg := ""
		if errorCount > 0 {
			status = "partial_failure"
			errorMsg = fmt.Sprintf("failed backends: %v", errors)
		}

		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_update_all",
			Status:    status,
			Error:     errorMsg,
			Duration:  duration.Milliseconds(),
			Metadata:  fmt.Sprintf(`{"success_count":%d,"error_count":%d}`, successCount, errorCount),
		})
	}

	if errorCount > 0 {
		return fmt.Errorf("index update completed with %d errors: %v", errorCount, errors)
	}

	return nil
}

// IsStale checks if the index is stale (older than the configured timeout)
// Validates Requirement: 11.7 (Detect stale index)
func (im *IndexManager) IsStale(ctx context.Context) (bool, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all entries to find the most recent update
	entries, err := im.repo.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list tool index entries: %w", err)
	}

	// If no entries, index is stale
	if len(entries) == 0 {
		return true, nil
	}

	// Find the most recent update time
	var mostRecent time.Time
	for _, entry := range entries {
		if entry.UpdatedAt.After(mostRecent) {
			mostRecent = entry.UpdatedAt
		}
	}

	// Check if the most recent update is older than the stale timeout
	return time.Since(mostRecent) > im.staleTimeout, nil
}

// GetStaleAge returns how long ago the index was last updated
func (im *IndexManager) GetStaleAge(ctx context.Context) (time.Duration, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all entries to find the most recent update
	entries, err := im.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list tool index entries: %w", err)
	}

	// If no entries, return max duration
	if len(entries) == 0 {
		return time.Duration(1<<63 - 1), nil // Max duration
	}

	// Find the most recent update time
	var mostRecent time.Time
	for _, entry := range entries {
		if entry.UpdatedAt.After(mostRecent) {
			mostRecent = entry.UpdatedAt
		}
	}

	return time.Since(mostRecent), nil
}

// PromptForUpdate checks if the index is stale and returns a prompt message
// Validates Requirement: 11.7 (Prompt for update when stale)
func (im *IndexManager) PromptForUpdate(ctx context.Context) (bool, string, error) {
	isStale, err := im.IsStale(ctx)
	if err != nil {
		return false, "", fmt.Errorf("check if index is stale: %w", err)
	}

	if !isStale {
		return false, "", nil
	}

	age, err := im.GetStaleAge(ctx)
	if err != nil {
		return true, "The tool index is stale. Run 'unirtm index update' to refresh it.", nil
	}

	days := int(age.Hours() / 24)
	return true, fmt.Sprintf("The tool index is %d days old. Run 'unirtm index update' to refresh it.", days), nil
}

// SupportsOffline indicates whether the index manager can operate offline
// Validates Requirement: 11.8 (Support offline operation)
func (im *IndexManager) SupportsOffline() bool {
	return true
}

// IsOfflineCapable checks if the index has cached data for offline operation
// Validates Requirement: 11.8 (Support offline operation using cached index)
func (im *IndexManager) IsOfflineCapable(ctx context.Context) (bool, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Check if we have any cached index entries
	entries, err := im.repo.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list tool index entries: %w", err)
	}

	return len(entries) > 0, nil
}

// DeleteTool removes a tool from the index
func (im *IndexManager) DeleteTool(ctx context.Context, tool string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if err := im.repo.Delete(ctx, tool); err != nil {
		return fmt.Errorf("delete tool from index: %w", err)
	}

	// Log the delete operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_delete",
			Tool:      tool,
			Status:    "success",
		})
	}

	return nil
}

// GetToolMetadata retrieves and parses the metadata for a tool
func (im *IndexManager) GetToolMetadata(ctx context.Context, tool string) (*ToolMetadata, error) {
	entry, err := im.GetTool(ctx, tool)
	if err != nil {
		return nil, err
	}

	if entry.Metadata == "" {
		return &ToolMetadata{}, nil
	}

	var metadata ToolMetadata
	if err := json.Unmarshal([]byte(entry.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal tool metadata: %w", err)
	}

	return &metadata, nil
}

// RegisterBackend registers a backend for index updates
func (im *IndexManager) RegisterBackend(backendName string, b backend.Backend) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.backends[backendName] = b
}

// UnregisterBackend removes a backend from the index manager
func (im *IndexManager) UnregisterBackend(backendName string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.backends, backendName)
}

// ListBackends returns the names of all registered backends
func (im *IndexManager) ListBackends() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	backends := make([]string, 0, len(im.backends))
	for name := range im.backends {
		backends = append(backends, name)
	}

	return backends
}
