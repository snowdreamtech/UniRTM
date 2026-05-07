// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build standalone
// +build standalone

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// This is a standalone test file that can be run independently
// to verify the Index Manager implementation without dependencies
// on other service files that may have compilation errors.

// MockIndexRepo is a simple in-memory implementation for testing
type MockIndexRepo struct {
	entries map[string]*repository.IndexEntry
}

func NewMockIndexRepo() *MockIndexRepo {
	return &MockIndexRepo{
		entries: make(map[string]*repository.IndexEntry),
	}
}

func (m *MockIndexRepo) Upsert(ctx context.Context, entry *repository.IndexEntry) error {
	m.entries[entry.Tool] = entry
	return nil
}

func (m *MockIndexRepo) FindByTool(ctx context.Context, tool string) (*repository.IndexEntry, error) {
	entry, ok := m.entries[tool]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return entry, nil
}

func (m *MockIndexRepo) List(ctx context.Context) ([]*repository.IndexEntry, error) {
	result := make([]*repository.IndexEntry, 0, len(m.entries))
	for _, entry := range m.entries {
		result = append(result, entry)
	}
	return result, nil
}

func (m *MockIndexRepo) Search(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
	// Simple search implementation - case-insensitive substring match
	result := make([]*repository.IndexEntry, 0)
	for _, entry := range m.entries {
		if contains(entry.Tool, query) || contains(entry.Description, query) {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (m *MockIndexRepo) Delete(ctx context.Context, tool string) error {
	delete(m.entries, tool)
	return nil
}

func contains(s, substr string) bool {
	// Simple case-insensitive substring search
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	// Convert to lowercase for case-insensitive comparison
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// MockAuditRepo is a simple mock for audit logging
type MockAuditRepo struct{}

func (m *MockAuditRepo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	return nil
}

func (m *MockAuditRepo) Query(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	return []*repository.AuditEntry{}, nil
}

func TestIndexManager_Standalone_BasicOperations(t *testing.T) {
	ctx := context.Background()

	// Create manager
	repo := NewMockIndexRepo()
	auditRepo := &MockAuditRepo{}
	manager, err := service.NewIndexManager(repo, auditRepo, nil, service.IndexManagerConfig{
		StaleTimeout: 7 * 24 * time.Hour,
	})
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Test UpsertTool
	err = manager.UpsertTool(ctx, "node", "Node.js runtime", "https://nodejs.org", "MIT", "github", &service.ToolMetadata{
		AvailableVersions: []string{"20.0.0", "18.0.0"},
		Tags:              []string{"runtime", "javascript"},
	})
	require.NoError(t, err)

	// Test GetTool
	entry, err := manager.GetTool(ctx, "node")
	require.NoError(t, err)
	assert.Equal(t, "node", entry.Tool)
	assert.Equal(t, "Node.js runtime", entry.Description)
	assert.Equal(t, "github", entry.Backend)

	// Test ListTools
	tools, err := manager.ListTools(ctx)
	require.NoError(t, err)
	assert.Len(t, tools, 1)

	// Test SearchTools
	results, err := manager.SearchTools(ctx, service.SearchOptions{Query: "Node"})
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Test DeleteTool
	err = manager.DeleteTool(ctx, "node")
	require.NoError(t, err)

	// Verify deletion
	_, err = manager.GetTool(ctx, "node")
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestIndexManager_Standalone_StaleDetection(t *testing.T) {
	ctx := context.Background()

	repo := NewMockIndexRepo()
	auditRepo := &MockAuditRepo{}
	manager, err := service.NewIndexManager(repo, auditRepo, nil, service.IndexManagerConfig{
		StaleTimeout: 7 * 24 * time.Hour,
	})
	require.NoError(t, err)

	// Empty index should be stale
	isStale, err := manager.IsStale(ctx)
	require.NoError(t, err)
	assert.True(t, isStale)

	// Add a fresh entry
	err = manager.UpsertTool(ctx, "node", "Node.js", "https://nodejs.org", "MIT", "github", nil)
	require.NoError(t, err)

	// Should not be stale now
	isStale, err = manager.IsStale(ctx)
	require.NoError(t, err)
	assert.False(t, isStale)

	// Test prompt for update
	shouldPrompt, message, err := manager.PromptForUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, shouldPrompt)
	assert.Empty(t, message)
}

func TestIndexManager_Standalone_OfflineCapability(t *testing.T) {
	ctx := context.Background()

	repo := NewMockIndexRepo()
	auditRepo := &MockAuditRepo{}
	manager, err := service.NewIndexManager(repo, auditRepo, nil, service.IndexManagerConfig{})
	require.NoError(t, err)

	// Should support offline
	assert.True(t, manager.SupportsOffline())

	// Empty index - not offline capable
	capable, err := manager.IsOfflineCapable(ctx)
	require.NoError(t, err)
	assert.False(t, capable)

	// Add entry
	err = manager.UpsertTool(ctx, "node", "Node.js", "https://nodejs.org", "MIT", "github", nil)
	require.NoError(t, err)

	// Now offline capable
	capable, err = manager.IsOfflineCapable(ctx)
	require.NoError(t, err)
	assert.True(t, capable)
}

func TestIndexManager_Standalone_FilterByBackend(t *testing.T) {
	ctx := context.Background()

	repo := NewMockIndexRepo()
	auditRepo := &MockAuditRepo{}
	manager, err := service.NewIndexManager(repo, auditRepo, nil, service.IndexManagerConfig{})
	require.NoError(t, err)

	// Add tools with different backends
	err = manager.UpsertTool(ctx, "node", "Node.js", "https://nodejs.org", "MIT", "github", nil)
	require.NoError(t, err)

	err = manager.UpsertTool(ctx, "python", "Python", "https://python.org", "PSF", "aqua", nil)
	require.NoError(t, err)

	err = manager.UpsertTool(ctx, "go", "Go", "https://go.dev", "BSD", "github", nil)
	require.NoError(t, err)

	// Filter by github
	githubTools, err := manager.FilterByBackend(ctx, "github")
	require.NoError(t, err)
	assert.Len(t, githubTools, 2)

	// Filter by aqua
	aquaTools, err := manager.FilterByBackend(ctx, "aqua")
	require.NoError(t, err)
	assert.Len(t, aquaTools, 1)
	assert.Equal(t, "python", aquaTools[0].Tool)
}
