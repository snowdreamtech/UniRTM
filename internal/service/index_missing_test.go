package service

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestIndexManager_MissingMethods(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockIndexRepository{}
	manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
	require.NoError(t, err)

	// UpsertTool
	mockRepo.upsertFunc = func(ctx context.Context, entry *repository.IndexEntry) error {
		return nil
	}
	manager.UpsertTool(ctx, "toolA", "DescriptionA", "http://example.com/toolA", "MIT", "backendA", nil)

	// seedDefaultTools
	manager.seedDefaultTools(ctx)

	// UpdateFromBackend
	manager.UpdateFromBackend(ctx, "native")

	// UpdateFromAllBackends
	manager.UpdateFromAllBackends(ctx)

	// SupportsOffline
	manager.SupportsOffline()
}
