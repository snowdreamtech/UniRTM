// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCacheRepository is a mock implementation of repository.CacheRepository
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Purge(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewCacheManager(t *testing.T) {
	t.Run("creates cache manager with valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024, // 1MB
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)
		assert.NotNil(t, cm)
		assert.Equal(t, tmpDir, cm.cacheDir)
		assert.Equal(t, int64(1024*1024), cm.maxCacheSize)
	})

	t.Run("returns error when cache directory is empty", func(t *testing.T) {
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     "",
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		assert.Error(t, err)
		assert.Nil(t, cm)
		assert.Contains(t, err.Error(), "cache directory is required")
	})

	t.Run("uses default max cache size when not specified", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 0, // Will use default
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)
		assert.Equal(t, int64(5*1024*1024*1024), cm.maxCacheSize) // 5GB default
	})

	t.Run("creates cache directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "cache", "subdir")
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     cacheDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)
		assert.NotNil(t, cm)

		// Verify directory was created
		info, err := os.Stat(cacheDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestCacheManager_Set(t *testing.T) {
	t.Run("stores cache entry successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		ttl := 1 * time.Hour

		mockRepo.On("Set", ctx, key, value, ttl).Return(nil)

		err = cm.Set(ctx, key, value, ttl)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		ttl := 1 * time.Hour

		expectedErr := errors.New("database error")
		mockRepo.On("Set", ctx, key, value, ttl).Return(expectedErr)

		err = cm.Set(ctx, key, value, ttl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store cache entry in database")
		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_Get(t *testing.T) {
	t.Run("retrieves cache entry successfully and records hit", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		expectedValue := []byte("test-value")

		mockRepo.On("Get", ctx, key).Return(expectedValue, nil)

		value, err := cm.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)

		// Verify hit was recorded
		stats := cm.GetStats()
		assert.Equal(t, int64(1), stats.Hits)
		assert.Equal(t, int64(0), stats.Misses)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns not found error and records miss", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "nonexistent-key"

		mockRepo.On("Get", ctx, key).Return(nil, repository.ErrNotFound)

		value, err := cm.Get(ctx, key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrNotFound))
		assert.Nil(t, value)

		// Verify miss was recorded
		stats := cm.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)

		mockRepo.AssertExpectations(t)
	})

	t.Run("records miss when value is nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "expired-key"

		mockRepo.On("Get", ctx, key).Return(nil, nil)

		value, err := cm.Get(ctx, key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, repository.ErrNotFound))
		assert.Nil(t, value)

		// Verify miss was recorded
		stats := cm.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)

		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_GetWithChecksum(t *testing.T) {
	t.Run("retrieves and verifies checksum successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		expectedChecksum := cm.calculateChecksum(value)

		mockRepo.On("Get", ctx, key).Return(value, nil)

		retrievedValue, err := cm.GetWithChecksum(ctx, key, expectedChecksum)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		mockRepo.AssertExpectations(t)
	})

	t.Run("deletes entry and returns error on checksum mismatch", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		wrongChecksum := "wrong-checksum"

		mockRepo.On("Get", ctx, key).Return(value, nil)
		mockRepo.On("Delete", ctx, key).Return(nil)

		retrievedValue, err := cm.GetWithChecksum(ctx, key, wrongChecksum)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum mismatch")
		assert.Nil(t, retrievedValue)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when delete fails after checksum mismatch", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"
		value := []byte("test-value")
		wrongChecksum := "wrong-checksum"

		deleteErr := errors.New("delete failed")
		mockRepo.On("Get", ctx, key).Return(value, nil)
		mockRepo.On("Delete", ctx, key).Return(deleteErr)

		retrievedValue, err := cm.GetWithChecksum(ctx, key, wrongChecksum)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum mismatch and failed to delete corrupted entry")
		assert.Nil(t, retrievedValue)

		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_Delete(t *testing.T) {
	t.Run("deletes cache entry successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"

		mockRepo.On("Delete", ctx, key).Return(nil)

		err = cm.Delete(ctx, key)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test-key"

		expectedErr := errors.New("database error")
		mockRepo.On("Delete", ctx, key).Return(expectedErr)

		err = cm.Delete(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete cache entry")
		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_PurgeExpired(t *testing.T) {
	t.Run("purges expired entries successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()

		mockRepo.On("Purge", ctx).Return(nil)

		err = cm.PurgeExpired(ctx)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()

		expectedErr := errors.New("database error")
		mockRepo.On("Purge", ctx).Return(expectedErr)

		err = cm.PurgeExpired(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "purge expired cache entries")
		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_PurgeAll(t *testing.T) {
	t.Run("purges all entries successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()

		mockRepo.On("Purge", ctx).Return(nil)

		err = cm.PurgeAll(ctx)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_GetStats(t *testing.T) {
	t.Run("returns accurate cache statistics", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()

		// Simulate some hits and misses
		mockRepo.On("Get", ctx, "key1").Return([]byte("value1"), nil)
		mockRepo.On("Get", ctx, "key2").Return([]byte("value2"), nil)
		mockRepo.On("Get", ctx, "key3").Return(nil, repository.ErrNotFound)

		_, _ = cm.Get(ctx, "key1") // hit
		_, _ = cm.Get(ctx, "key2") // hit
		_, _ = cm.Get(ctx, "key3") // miss

		stats := cm.GetStats()
		assert.Equal(t, int64(2), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)

		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_ResetStats(t *testing.T) {
	t.Run("resets cache statistics", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		ctx := context.Background()

		// Simulate some hits
		mockRepo.On("Get", ctx, "key1").Return([]byte("value1"), nil)
		_, _ = cm.Get(ctx, "key1")

		// Verify stats before reset
		stats := cm.GetStats()
		assert.Equal(t, int64(1), stats.Hits)

		// Reset stats
		cm.ResetStats()

		// Verify stats after reset
		stats = cm.GetStats()
		assert.Equal(t, int64(0), stats.Hits)
		assert.Equal(t, int64(0), stats.Misses)

		mockRepo.AssertExpectations(t)
	})
}

func TestCacheManager_GetCacheSize(t *testing.T) {
	t.Run("calculates cache size correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		// Create some test files
		testFile1 := filepath.Join(tmpDir, "file1.txt")
		testFile2 := filepath.Join(tmpDir, "file2.txt")

		err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
		require.NoError(t, err)

		err = os.WriteFile(testFile2, []byte("test content 2 longer"), 0644)
		require.NoError(t, err)

		size, err := cm.GetCacheSize()
		assert.NoError(t, err)
		assert.Greater(t, size, int64(0))
	})

	t.Run("returns zero for empty cache directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := new(MockAuditRepository)

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024,
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		size, err := cm.GetCacheSize()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})
}

func TestCacheManager_AutoCleanup(t *testing.T) {
	t.Run("performs cleanup when size exceeds threshold", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		// Set a very small threshold to trigger cleanup
		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 10, // 10 bytes
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		// Create a file that exceeds the threshold
		testFile := filepath.Join(tmpDir, "large-file.txt")
		err = os.WriteFile(testFile, []byte("this is a large file content"), 0644)
		require.NoError(t, err)

		ctx := context.Background()

		mockRepo.On("Purge", ctx).Return(nil)

		err = cm.AutoCleanup(ctx)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("does not cleanup when size is below threshold", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockRepo := new(MockCacheRepository)
		mockAuditRepo := &MockAuditRepository{}

		config := CacheManagerConfig{
			CacheDir:     tmpDir,
			MaxCacheSize: 1024 * 1024, // 1MB
		}

		cm, err := NewCacheManager(mockRepo, mockAuditRepo, config)
		require.NoError(t, err)

		// Create a small file
		testFile := filepath.Join(tmpDir, "small-file.txt")
		err = os.WriteFile(testFile, []byte("small"), 0644)
		require.NoError(t, err)

		ctx := context.Background()

		err = cm.AutoCleanup(ctx)
		assert.NoError(t, err)

		// Verify Purge was not called
		mockRepo.AssertNotCalled(t, "Purge", ctx)
	})
}
