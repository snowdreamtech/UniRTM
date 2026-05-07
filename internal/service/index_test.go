package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// MockIndexRepository is a mock implementation of repository.IndexRepository
type MockIndexRepository struct {
	upsertFunc     func(ctx context.Context, entry *repository.IndexEntry) error
	findByToolFunc func(ctx context.Context, tool string) (*repository.IndexEntry, error)
	listFunc       func(ctx context.Context) ([]*repository.IndexEntry, error)
	searchFunc     func(ctx context.Context, query string) ([]*repository.IndexEntry, error)
	deleteFunc     func(ctx context.Context, tool string) error
}

func (m *MockIndexRepository) Upsert(ctx context.Context, entry *repository.IndexEntry) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, entry)
	}
	return nil
}

func (m *MockIndexRepository) FindByTool(ctx context.Context, tool string) (*repository.IndexEntry, error) {
	if m.findByToolFunc != nil {
		return m.findByToolFunc(ctx, tool)
	}
	return nil, repository.ErrNotFound
}

func (m *MockIndexRepository) List(ctx context.Context) ([]*repository.IndexEntry, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []*repository.IndexEntry{}, nil
}

func (m *MockIndexRepository) Search(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query)
	}
	return []*repository.IndexEntry{}, nil
}

func (m *MockIndexRepository) Delete(ctx context.Context, tool string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tool)
	}
	return nil
}

// MockBackend is a mock implementation of backend.Backend
type MockBackend struct {
	name string
}

func (m *MockBackend) Name() string {
	return m.name
}

func (m *MockBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	return nil, nil
}

func (m *MockBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	return nil, nil
}

func (m *MockBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	return nil, nil
}

func (m *MockBackend) SupportsChecksum() bool {
	return true
}

func (m *MockBackend) SupportsGPG() bool {
	return false
}

func TestNewIndexManager(t *testing.T) {
	tests := []struct {
		name      string
		repo      repository.IndexRepository
		auditRepo repository.AuditRepository
		backends  map[string]backend.Backend
		config    IndexManagerConfig
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			repo:      &MockIndexRepository{},
			auditRepo: &MockAuditRepository{},
			backends:  map[string]backend.Backend{},
			config:    IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantErr:   false,
		},
		{
			name:      "nil repository",
			repo:      nil,
			auditRepo: &MockAuditRepository{},
			backends:  map[string]backend.Backend{},
			config:    IndexManagerConfig{},
			wantErr:   true,
		},
		{
			name:      "default stale timeout",
			repo:      &MockIndexRepository{},
			auditRepo: &MockAuditRepository{},
			backends:  map[string]backend.Backend{},
			config:    IndexManagerConfig{StaleTimeout: 0},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewIndexManager(tt.repo, tt.auditRepo, tt.backends, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				if tt.config.StaleTimeout == 0 {
					assert.Equal(t, 7*24*time.Hour, manager.staleTimeout)
				}
			}
		})
	}
}

func TestIndexManager_UpsertTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		tool        string
		description string
		homepage    string
		license     string
		backend     string
		metadata    *ToolMetadata
		setupMock   func(*MockIndexRepository)
		wantErr     bool
	}{
		{
			name:        "successful upsert",
			tool:        "node",
			description: "Node.js runtime",
			homepage:    "https://nodejs.org",
			license:     "MIT",
			backend:     "github",
			metadata: &ToolMetadata{
				AvailableVersions: []string{"20.0.0", "18.0.0"},
				Tags:              []string{"runtime", "javascript"},
			},
			setupMock: func(m *MockIndexRepository) {
				m.upsertFunc = func(ctx context.Context, entry *repository.IndexEntry) error {
					assert.Equal(t, "node", entry.Tool)
					assert.Equal(t, "Node.js runtime", entry.Description)
					assert.Equal(t, "github", entry.Backend)
					assert.NotEmpty(t, entry.Metadata)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "upsert without metadata",
			tool:        "python",
			description: "Python runtime",
			homepage:    "https://python.org",
			license:     "PSF",
			backend:     "aqua",
			metadata:    nil,
			setupMock: func(m *MockIndexRepository) {
				m.upsertFunc = func(ctx context.Context, entry *repository.IndexEntry) error {
					assert.Equal(t, "python", entry.Tool)
					assert.Empty(t, entry.Metadata)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "repository error",
			tool:        "go",
			description: "Go runtime",
			homepage:    "https://go.dev",
			license:     "BSD",
			backend:     "github",
			metadata:    nil,
			setupMock: func(m *MockIndexRepository) {
				m.upsertFunc = func(ctx context.Context, entry *repository.IndexEntry) error {
					return errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			err = manager.UpsertTool(ctx, tt.tool, tt.description, tt.homepage, tt.license, tt.backend, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIndexManager_GetTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		tool      string
		setupMock func(*MockIndexRepository)
		wantErr   bool
		wantEntry *repository.IndexEntry
	}{
		{
			name: "tool found",
			tool: "node",
			setupMock: func(m *MockIndexRepository) {
				m.findByToolFunc = func(ctx context.Context, tool string) (*repository.IndexEntry, error) {
					return &repository.IndexEntry{
						Tool:        "node",
						Description: "Node.js runtime",
						Backend:     "github",
					}, nil
				}
			},
			wantErr: false,
			wantEntry: &repository.IndexEntry{
				Tool:        "node",
				Description: "Node.js runtime",
				Backend:     "github",
			},
		},
		{
			name: "tool not found",
			tool: "nonexistent",
			setupMock: func(m *MockIndexRepository) {
				m.findByToolFunc = func(ctx context.Context, tool string) (*repository.IndexEntry, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr:   true,
			wantEntry: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			entry, err := manager.GetTool(ctx, tt.tool)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEntry.Tool, entry.Tool)
				assert.Equal(t, tt.wantEntry.Description, entry.Description)
			}
		})
	}
}

func TestIndexManager_SearchTools(t *testing.T) {
	ctx := context.Background()

	allTools := []*repository.IndexEntry{
		{Tool: "node", Description: "Node.js runtime", Backend: "github"},
		{Tool: "python", Description: "Python runtime", Backend: "aqua"},
		{Tool: "go", Description: "Go runtime", Backend: "github"},
		{Tool: "ruby", Description: "Ruby runtime", Backend: "aqua"},
	}

	tests := []struct {
		name      string
		options   SearchOptions
		setupMock func(*MockIndexRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "search all",
			options: SearchOptions{
				Query: "runtime",
			},
			setupMock: func(m *MockIndexRepository) {
				m.searchFunc = func(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "filter by backend",
			options: SearchOptions{
				Query:   "runtime",
				Backend: "github",
			},
			setupMock: func(m *MockIndexRepository) {
				m.searchFunc = func(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "with limit",
			options: SearchOptions{
				Query: "runtime",
				Limit: 2,
			},
			setupMock: func(m *MockIndexRepository) {
				m.searchFunc = func(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "with offset",
			options: SearchOptions{
				Query:  "runtime",
				Offset: 2,
			},
			setupMock: func(m *MockIndexRepository) {
				m.searchFunc = func(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "offset beyond results",
			options: SearchOptions{
				Query:  "runtime",
				Offset: 10,
			},
			setupMock: func(m *MockIndexRepository) {
				m.searchFunc = func(ctx context.Context, query string) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			results, err := manager.SearchTools(ctx, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
			}
		})
	}
}

func TestIndexManager_FilterByBackend(t *testing.T) {
	ctx := context.Background()

	allTools := []*repository.IndexEntry{
		{Tool: "node", Backend: "github"},
		{Tool: "python", Backend: "aqua"},
		{Tool: "go", Backend: "github"},
		{Tool: "ruby", Backend: "aqua"},
	}

	tests := []struct {
		name        string
		backend     string
		setupMock   func(*MockIndexRepository)
		wantCount   int
		wantBackend string
		wantErr     bool
	}{
		{
			name:    "filter github",
			backend: "github",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount:   2,
			wantBackend: "github",
			wantErr:     false,
		},
		{
			name:    "filter aqua",
			backend: "aqua",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount:   2,
			wantBackend: "aqua",
			wantErr:     false,
		},
		{
			name:    "filter nonexistent backend",
			backend: "nonexistent",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return allTools, nil
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			results, err := manager.FilterByBackend(ctx, tt.backend)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
				for _, entry := range results {
					assert.Equal(t, tt.wantBackend, entry.Backend)
				}
			}
		})
	}
}

func TestIndexManager_IsStale(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(*MockIndexRepository)
		config    IndexManagerConfig
		wantStale bool
		wantErr   bool
	}{
		{
			name: "fresh index",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node", UpdatedAt: time.Now()},
					}, nil
				}
			},
			config:    IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantStale: false,
			wantErr:   false,
		},
		{
			name: "stale index",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node", UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
					}, nil
				}
			},
			config:    IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantStale: true,
			wantErr:   false,
		},
		{
			name: "empty index",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{}, nil
				}
			},
			config:    IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantStale: true,
			wantErr:   false,
		},
		{
			name: "multiple entries - use most recent",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node", UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
						{Tool: "python", UpdatedAt: time.Now()},
					}, nil
				}
			},
			config:    IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantStale: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, tt.config)
			require.NoError(t, err)

			isStale, err := manager.IsStale(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStale, isStale)
			}
		})
	}
}

func TestIndexManager_PromptForUpdate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setupMock      func(*MockIndexRepository)
		config         IndexManagerConfig
		wantPrompt     bool
		wantMessageNil bool
		wantErr        bool
	}{
		{
			name: "fresh index - no prompt",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node", UpdatedAt: time.Now()},
					}, nil
				}
			},
			config:         IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantPrompt:     false,
			wantMessageNil: true,
			wantErr:        false,
		},
		{
			name: "stale index - prompt",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node", UpdatedAt: time.Now().Add(-10 * 24 * time.Hour)},
					}, nil
				}
			},
			config:         IndexManagerConfig{StaleTimeout: 7 * 24 * time.Hour},
			wantPrompt:     true,
			wantMessageNil: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, tt.config)
			require.NoError(t, err)

			shouldPrompt, message, err := manager.PromptForUpdate(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPrompt, shouldPrompt)
				if tt.wantMessageNil {
					assert.Empty(t, message)
				} else {
					assert.NotEmpty(t, message)
				}
			}
		})
	}
}

func TestIndexManager_IsOfflineCapable(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(*MockIndexRepository)
		wantOk    bool
		wantErr   bool
	}{
		{
			name: "has cached entries",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{
						{Tool: "node"},
					}, nil
				}
			},
			wantOk:  true,
			wantErr: false,
		},
		{
			name: "no cached entries",
			setupMock: func(m *MockIndexRepository) {
				m.listFunc = func(ctx context.Context) ([]*repository.IndexEntry, error) {
					return []*repository.IndexEntry{}, nil
				}
			},
			wantOk:  false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			ok, err := manager.IsOfflineCapable(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOk, ok)
			}
		})
	}
}

func TestIndexManager_GetToolMetadata(t *testing.T) {
	ctx := context.Background()

	metadata := &ToolMetadata{
		AvailableVersions: []string{"20.0.0", "18.0.0"},
		Tags:              []string{"runtime", "javascript"},
		Stars:             50000,
	}
	metadataJSON, _ := json.Marshal(metadata)

	tests := []struct {
		name      string
		tool      string
		setupMock func(*MockIndexRepository)
		wantErr   bool
	}{
		{
			name: "tool with metadata",
			tool: "node",
			setupMock: func(m *MockIndexRepository) {
				m.findByToolFunc = func(ctx context.Context, tool string) (*repository.IndexEntry, error) {
					return &repository.IndexEntry{
						Tool:     "node",
						Metadata: string(metadataJSON),
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "tool without metadata",
			tool: "python",
			setupMock: func(m *MockIndexRepository) {
				m.findByToolFunc = func(ctx context.Context, tool string) (*repository.IndexEntry, error) {
					return &repository.IndexEntry{
						Tool:     "python",
						Metadata: "",
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "tool not found",
			tool: "nonexistent",
			setupMock: func(m *MockIndexRepository) {
				m.findByToolFunc = func(ctx context.Context, tool string) (*repository.IndexEntry, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			meta, err := manager.GetToolMetadata(ctx, tt.tool)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, meta)
			}
		})
	}
}

func TestIndexManager_BackendManagement(t *testing.T) {
	mockRepo := &MockIndexRepository{}
	manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
	require.NoError(t, err)

	// Test RegisterBackend
	githubBackend := &MockBackend{name: "github"}
	manager.RegisterBackend("github", githubBackend)

	backends := manager.ListBackends()
	assert.Contains(t, backends, "github")

	// Test UnregisterBackend
	manager.UnregisterBackend("github")

	backends = manager.ListBackends()
	assert.NotContains(t, backends, "github")
}

func TestIndexManager_DeleteTool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		tool      string
		setupMock func(*MockIndexRepository)
		wantErr   bool
	}{
		{
			name: "successful delete",
			tool: "node",
			setupMock: func(m *MockIndexRepository) {
				m.deleteFunc = func(ctx context.Context, tool string) error {
					assert.Equal(t, "node", tool)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "repository error",
			tool: "python",
			setupMock: func(m *MockIndexRepository) {
				m.deleteFunc = func(ctx context.Context, tool string) error {
					return errors.New("database error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockIndexRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			manager, err := NewIndexManager(mockRepo, &MockAuditRepository{}, nil, IndexManagerConfig{})
			require.NoError(t, err)

			err = manager.DeleteTool(ctx, tt.tool)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
