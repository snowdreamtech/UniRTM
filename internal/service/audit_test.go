package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuditRepository is a mock implementation of repository.AuditRepository for testing
type MockAuditRepository struct {
	logFunc   func(ctx context.Context, entry *repository.AuditEntry) error
	queryFunc func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error)
	entries   []*repository.AuditEntry
}

func (m *MockAuditRepository) Log(ctx context.Context, entry *repository.AuditEntry) error {
	if m.logFunc != nil {
		return m.logFunc(ctx, entry)
	}
	// Default behavior: store entry
	entry.ID = int64(len(m.entries) + 1)
	entry.Timestamp = time.Now()
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditRepository) Query(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, filter)
	}
	// Default behavior: return all entries
	return m.entries, nil
}

func TestNewAuditService(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	require.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
}

func TestAuditService_LogOperation(t *testing.T) {
	tests := []struct {
		name      string
		entry     *AuditLogEntry
		repoError error
		wantErr   bool
	}{
		{
			name: "successful operation with metadata",
			entry: &AuditLogEntry{
				Operation: OperationInstall,
				Tool:      "node",
				Version:   "20.0.0",
				Status:    StatusSuccess,
				Duration:  1500,
				Metadata: map[string]interface{}{
					"backend": "github",
					"size":    1024000,
				},
			},
			wantErr: false,
		},
		{
			name: "failed operation with error",
			entry: &AuditLogEntry{
				Operation: OperationInstall,
				Tool:      "python",
				Version:   "3.11.0",
				Status:    StatusFailure,
				Error:     "download failed: connection timeout",
				Duration:  5000,
			},
			wantErr: false,
		},
		{
			name: "operation without tool",
			entry: &AuditLogEntry{
				Operation: OperationCachePurge,
				Status:    StatusSuccess,
				Duration:  200,
				Metadata: map[string]interface{}{
					"purged_count": 10,
				},
			},
			wantErr: false,
		},
		{
			name: "repository error",
			entry: &AuditLogEntry{
				Operation: OperationInstall,
				Tool:      "go",
				Version:   "1.21.0",
				Status:    StatusSuccess,
				Duration:  1000,
			},
			repoError: errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockAuditRepository{
				logFunc: func(ctx context.Context, entry *repository.AuditEntry) error {
					if tt.repoError != nil {
						return tt.repoError
					}
					// Verify entry fields
					assert.Equal(t, string(tt.entry.Operation), entry.Operation)
					assert.Equal(t, tt.entry.Tool, entry.Tool)
					assert.Equal(t, tt.entry.Version, entry.Version)
					assert.Equal(t, string(tt.entry.Status), entry.Status)
					assert.Equal(t, tt.entry.Error, entry.Error)
					assert.Equal(t, tt.entry.Duration, entry.Duration)

					// Verify metadata JSON
					if tt.entry.Metadata != nil {
						var metadata map[string]interface{}
						err := json.Unmarshal([]byte(entry.Metadata), &metadata)
						require.NoError(t, err)
						// JSON unmarshaling converts numbers to float64, so we need to compare carefully
						for key, expectedValue := range tt.entry.Metadata {
							actualValue, ok := metadata[key]
							require.True(t, ok, "key %s not found in metadata", key)
							// Handle numeric type conversions
							switch v := expectedValue.(type) {
							case int:
								assert.Equal(t, float64(v), actualValue, "key %s", key)
							case int64:
								assert.Equal(t, float64(v), actualValue, "key %s", key)
							default:
								assert.Equal(t, expectedValue, actualValue, "key %s", key)
							}
						}
					}

					return nil
				},
			}

			service := NewAuditService(repo)
			err := service.LogOperation(context.Background(), tt.entry)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuditService_LogInstall(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		version  string
		duration time.Duration
		err      error
		wantOp   OperationType
		wantStat OperationStatus
	}{
		{
			name:     "successful install",
			tool:     "node",
			version:  "20.0.0",
			duration: 2 * time.Second,
			err:      nil,
			wantOp:   OperationInstall,
			wantStat: StatusSuccess,
		},
		{
			name:     "failed install",
			tool:     "python",
			version:  "3.11.0",
			duration: 5 * time.Second,
			err:      errors.New("download failed"),
			wantOp:   OperationInstall,
			wantStat: StatusFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockAuditRepository{}
			service := NewAuditService(repo)

			err := service.LogInstall(context.Background(), tt.tool, tt.version, tt.duration, tt.err)
			require.NoError(t, err)

			require.Len(t, repo.entries, 1)
			entry := repo.entries[0]
			assert.Equal(t, string(tt.wantOp), entry.Operation)
			assert.Equal(t, tt.tool, entry.Tool)
			assert.Equal(t, tt.version, entry.Version)
			assert.Equal(t, string(tt.wantStat), entry.Status)
			assert.Equal(t, tt.duration.Milliseconds(), entry.Duration)

			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), entry.Error)
			} else {
				assert.Empty(t, entry.Error)
			}
		})
	}
}

func TestAuditService_LogUninstall(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogUninstall(context.Background(), "node", "20.0.0", 500*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationUninstall), entry.Operation)
	assert.Equal(t, "node", entry.Tool)
	assert.Equal(t, "20.0.0", entry.Version)
	assert.Equal(t, string(StatusSuccess), entry.Status)
	assert.Equal(t, int64(500), entry.Duration)
}

func TestAuditService_LogActivate(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogActivate(context.Background(), "python", "3.11.0", 100*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationActivate), entry.Operation)
	assert.Equal(t, "python", entry.Tool)
	assert.Equal(t, "3.11.0", entry.Version)
	assert.Equal(t, string(StatusSuccess), entry.Status)
}

func TestAuditService_LogDeactivate(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogDeactivate(context.Background(), "go", "1.21.0", 50*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationDeactivate), entry.Operation)
	assert.Equal(t, "go", entry.Tool)
	assert.Equal(t, "1.21.0", entry.Version)
}

func TestAuditService_LogUpdate(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogUpdate(context.Background(), "node", "18.0.0", "20.0.0", 3*time.Second, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationUpdate), entry.Operation)
	assert.Equal(t, "node", entry.Tool)
	assert.Equal(t, "20.0.0", entry.Version)
	assert.Equal(t, string(StatusSuccess), entry.Status)

	// Verify metadata contains old and new versions
	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(entry.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, "18.0.0", metadata["old_version"])
	assert.Equal(t, "20.0.0", metadata["new_version"])
}

func TestAuditService_LogCachePurge(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogCachePurge(context.Background(), 200*time.Millisecond, 15, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationCachePurge), entry.Operation)
	assert.Equal(t, string(StatusSuccess), entry.Status)

	// Verify metadata contains purged count
	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(entry.Metadata), &metadata)
	require.NoError(t, err)
	// JSON numbers are float64
	assert.Equal(t, float64(15), metadata["purged_count"])
}

func TestAuditService_LogConfigLoad(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogConfigLoad(context.Background(), "/path/to/config.toml", 50*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationConfigLoad), entry.Operation)

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(entry.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, "/path/to/config.toml", metadata["config_path"])
}

func TestAuditService_LogConfigUpdate(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogConfigUpdate(context.Background(), "/path/to/config.toml", 100*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationConfigUpdate), entry.Operation)
}

func TestAuditService_LogIndexUpdate(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogIndexUpdate(context.Background(), "github", 2*time.Second, 50, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationIndexUpdate), entry.Operation)

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(entry.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, "github", metadata["backend"])
	assert.Equal(t, float64(50), metadata["tool_count"])
}

func TestAuditService_LogVersionResolve(t *testing.T) {
	repo := &MockAuditRepository{}
	service := NewAuditService(repo)

	err := service.LogVersionResolve(context.Background(), "node", "latest", "20.0.0", 100*time.Millisecond, nil)
	require.NoError(t, err)

	require.Len(t, repo.entries, 1)
	entry := repo.entries[0]
	assert.Equal(t, string(OperationVersionResolve), entry.Operation)
	assert.Equal(t, "node", entry.Tool)
	assert.Equal(t, "20.0.0", entry.Version)

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(entry.Metadata), &metadata)
	require.NoError(t, err)
	assert.Equal(t, "latest", metadata["version_spec"])
	assert.Equal(t, "20.0.0", metadata["resolved_version"])
}

func TestAuditService_QueryAuditLogs(t *testing.T) {
	expectedEntries := []*repository.AuditEntry{
		{
			ID:        1,
			Operation: "install",
			Tool:      "node",
			Version:   "20.0.0",
			Status:    "success",
		},
		{
			ID:        2,
			Operation: "uninstall",
			Tool:      "python",
			Version:   "3.11.0",
			Status:    "success",
		},
	}

	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			return expectedEntries, nil
		},
	}

	service := NewAuditService(repo)
	filter := repository.AuditFilter{Limit: 10}

	entries, err := service.QueryAuditLogs(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, expectedEntries, entries)
}

func TestAuditService_GetRecentLogs(t *testing.T) {
	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			assert.Equal(t, 10, filter.Limit)
			return []*repository.AuditEntry{}, nil
		},
	}

	service := NewAuditService(repo)
	_, err := service.GetRecentLogs(context.Background(), 10)
	require.NoError(t, err)
}

func TestAuditService_GetLogsByOperation(t *testing.T) {
	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			assert.Equal(t, "install", filter.Operation)
			assert.Equal(t, 20, filter.Limit)
			return []*repository.AuditEntry{}, nil
		},
	}

	service := NewAuditService(repo)
	_, err := service.GetLogsByOperation(context.Background(), OperationInstall, 20)
	require.NoError(t, err)
}

func TestAuditService_GetLogsByTool(t *testing.T) {
	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			assert.Equal(t, "node", filter.Tool)
			assert.Equal(t, 15, filter.Limit)
			return []*repository.AuditEntry{}, nil
		},
	}

	service := NewAuditService(repo)
	_, err := service.GetLogsByTool(context.Background(), "node", 15)
	require.NoError(t, err)
}

func TestAuditService_GetLogsByStatus(t *testing.T) {
	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			assert.Equal(t, "failure", filter.Status)
			assert.Equal(t, 25, filter.Limit)
			return []*repository.AuditEntry{}, nil
		},
	}

	service := NewAuditService(repo)
	_, err := service.GetLogsByStatus(context.Background(), StatusFailure, 25)
	require.NoError(t, err)
}

func TestAuditService_GetLogsByTimeRange(t *testing.T) {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	repo := &MockAuditRepository{
		queryFunc: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
			require.NotNil(t, filter.StartTime)
			require.NotNil(t, filter.EndTime)
			assert.Equal(t, startTime, *filter.StartTime)
			assert.Equal(t, endTime, *filter.EndTime)
			assert.Equal(t, 30, filter.Limit)
			return []*repository.AuditEntry{}, nil
		},
	}

	service := NewAuditService(repo)
	_, err := service.GetLogsByTimeRange(context.Background(), startTime, endTime, 30)
	require.NoError(t, err)
}

func TestAuditService_LogOperation_InvalidMetadata(t *testing.T) {
	// Test that invalid metadata (e.g., circular references) doesn't crash
	// but logs a warning and continues with empty metadata
	repo := &MockAuditRepository{
		logFunc: func(ctx context.Context, entry *repository.AuditEntry) error {
			// Metadata should be empty or "{}" when marshaling fails
			assert.NotEmpty(t, entry.Metadata)
			return nil
		},
	}

	service := NewAuditService(repo)

	// Create metadata with a channel (which can't be marshaled to JSON)
	entry := &AuditLogEntry{
		Operation: OperationInstall,
		Tool:      "test",
		Version:   "1.0.0",
		Status:    StatusSuccess,
		Duration:  100,
		Metadata: map[string]interface{}{
			"channel": make(chan int), // This will fail JSON marshaling
		},
	}

	// Should not return error, but log warning internally
	err := service.LogOperation(context.Background(), entry)
	require.NoError(t, err)
}
