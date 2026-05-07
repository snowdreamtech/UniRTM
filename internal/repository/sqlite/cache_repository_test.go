// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheRepository_Set_Get(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Set a cache entry
	key := "test-key"
	value := []byte("test-value")
	ttl := 1 * time.Hour

	err = repo.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Get the cache entry
	retrieved, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)
}

func TestCacheRepository_Get_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Try to get non-existent key
	retrieved, err := repo.Get(ctx, "nonexistent-key")
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestCacheRepository_Get_Expired(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Set a cache entry with very short TTL
	key := "expiring-key"
	value := []byte("expiring-value")
	ttl := 100 * time.Millisecond

	err = repo.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Try to get expired entry
	retrieved, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestCacheRepository_Set_Upsert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	key := "upsert-key"
	value1 := []byte("value-1")
	value2 := []byte("value-2")
	ttl := 1 * time.Hour

	// Set initial value
	err = repo.Set(ctx, key, value1, ttl)
	require.NoError(t, err)

	// Update with new value
	err = repo.Set(ctx, key, value2, ttl)
	require.NoError(t, err)

	// Get should return the updated value
	retrieved, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value2, retrieved)
}

func TestCacheRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Set a cache entry
	key := "delete-key"
	value := []byte("delete-value")
	ttl := 1 * time.Hour

	err = repo.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Delete the entry
	err = repo.Delete(ctx, key)
	require.NoError(t, err)

	// Verify it's deleted
	retrieved, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestCacheRepository_Purge(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Set multiple cache entries with different TTLs
	entries := []struct {
		key   string
		value []byte
		ttl   time.Duration
	}{
		{"key1", []byte("value1"), 100 * time.Millisecond}, // Will expire
		{"key2", []byte("value2"), 100 * time.Millisecond}, // Will expire
		{"key3", []byte("value3"), 1 * time.Hour},          // Won't expire
	}

	for _, entry := range entries {
		err := repo.Set(ctx, entry.key, entry.value, entry.ttl)
		require.NoError(t, err)
	}

	// Wait for some entries to expire
	time.Sleep(200 * time.Millisecond)

	// Purge expired entries
	err = repo.Purge(ctx)
	require.NoError(t, err)

	// Verify expired entries are gone
	retrieved1, err := repo.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Nil(t, retrieved1)

	retrieved2, err := repo.Get(ctx, "key2")
	require.NoError(t, err)
	assert.Nil(t, retrieved2)

	// Verify non-expired entry still exists
	retrieved3, err := repo.Get(ctx, "key3")
	require.NoError(t, err)
	assert.Equal(t, []byte("value3"), retrieved3)
}

func TestCacheRepository_BinaryData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo, err := NewCacheRepository(db.Conn())
	require.NoError(t, err)
	defer repo.Close()

	ctx := context.Background()

	// Test with binary data
	key := "binary-key"
	value := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	ttl := 1 * time.Hour

	err = repo.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)
}
