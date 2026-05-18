// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

// Mock implementations for testing Update Manager

type mockUpdateBackend struct {
	name             string
	versions         map[string]*backend.VersionInfo
	supportsChecksum bool
	supportsGPG      bool
	listErr          error
	resolveErr       error
	downloadErr      error
}

func (m *mockUpdateBackend) Name() string {
	return m.name
}

func (m *mockUpdateBackend) AttestationType() string {
	return ""
}

func (m *mockUpdateBackend) Dependencies() []string {
	return nil
}

func (m *mockUpdateBackend) GetReach() string {
	return ""
}

func (m *mockUpdateBackend) IsRecommended() bool {
	return false
}

func (m *mockUpdateBackend) IsScriptless() bool {
	return false
}

func (m *mockUpdateBackend) IsStable() bool {
	return true
}

func (m *mockUpdateBackend) SupportsOffline() bool {
	return false
}

func (m *mockUpdateBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var versions []backend.VersionInfo
	for _, v := range m.versions {
		versions = append(versions, *v)
	}
	return versions, nil
}

func (m *mockUpdateBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	if m.resolveErr != nil {
		return nil, m.resolveErr
	}
	if versionRequest == "latest" {
		// Return the highest version (simplified)
		for _, v := range m.versions {
			return v, nil
		}
	}
	if v, ok := m.versions[versionRequest]; ok {
		return v, nil
	}
	return nil, errors.New("version not found")
}

func (m *mockUpdateBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	if v, ok := m.versions[version]; ok {
		return v, nil
	}
	return nil, errors.New("version not found")
}

func (m *mockUpdateBackend) SupportsChecksum() bool {
	return m.supportsChecksum
}

func (m *mockUpdateBackend) SupportsGPG() bool {
	return m.supportsGPG
}

type mockProvider struct {
	name           string
	installErr     error
	postInstallErr error
	uninstallErr   error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Install(ctx context.Context, tool string, installPath, artifactPath, version string) error {
	return m.installErr
}

func (m *mockProvider) PostInstall(ctx context.Context, tool string, installPath, version string) error {
	return m.postInstallErr
}

func (m *mockProvider) Uninstall(ctx context.Context, tool string, installPath, version string) error {
	return m.uninstallErr
}

func (m *mockProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return "", nil
}

func (m *mockProvider) GenerateShims(tool string, installPath, version string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (m *mockProvider) ListExecutables(tool string, installPath, version string) ([]string, error) {
	return []string{}, nil
}

func (m *mockProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{}, nil
}

func (m *mockProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{}, nil
}

type mockProviderRegistry struct {
	providers map[string]provider.Provider
}

func (m *mockProviderRegistry) Register(p provider.Provider) error {
	if m.providers == nil {
		m.providers = make(map[string]provider.Provider)
	}
	m.providers[p.Name()] = p
	return nil
}

func (m *mockProviderRegistry) Get(tool string) provider.Provider {
	if p, ok := m.providers[tool]; ok {
		return p
	}
	return &mockProvider{name: "generic"}
}

type mockDownloader struct {
	downloadErr       error
	verifyChecksumErr error
}

func (m *mockDownloader) Download(ctx context.Context, url, destination string, opts download.DownloadOptions) error {
	return m.downloadErr
}

func (m *mockDownloader) VerifyChecksum(ctx context.Context, file, expectedChecksum string) error {
	return m.verifyChecksumErr
}

type mockDownloadManager struct {
	downloader download.Downloader
}

func (m *mockDownloadManager) Register(protocol string, d download.Downloader) error {
	return nil
}

func (m *mockDownloadManager) Get(protocol string) (download.Downloader, error) {
	if m.downloader != nil {
		return m.downloader, nil
	}
	return nil, errors.New("downloader not found")
}

type mockInstallationRepo struct {
	installations map[string]*repository.Installation
	createErr     error
	findErr       error
	deleteErr     error
}

func (m *mockInstallationRepo) Create(ctx context.Context, installation *repository.Installation) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.installations == nil {
		m.installations = make(map[string]*repository.Installation)
	}
	key := installation.Tool + "-" + installation.Version
	m.installations[key] = installation
	return nil
}

func (m *mockInstallationRepo) Upsert(ctx context.Context, installation *repository.Installation) error {
	return m.Create(ctx, installation)
}

func (m *mockInstallationRepo) FindByToolAndVersion(ctx context.Context, tool string, version string) (*repository.Installation, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	// If version is empty, return any version of the tool
	if version == "" {
		for _, inst := range m.installations {
			if inst.Tool == tool {
				return inst, nil
			}
		}
	}
	key := tool + "-" + version
	if inst, ok := m.installations[key]; ok {
		return inst, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockInstallationRepo) GetByToolAndVersion(ctx context.Context, tool string, version string) (*repository.Installation, error) {
	return m.FindByToolAndVersion(ctx, tool, version)
}

func (m *mockInstallationRepo) List(ctx context.Context) ([]*repository.Installation, error) {
	var installations []*repository.Installation
	for _, inst := range m.installations {
		installations = append(installations, inst)
	}
	return installations, nil
}

func (m *mockInstallationRepo) Delete(ctx context.Context, tool string, version string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	key := tool + "-" + version
	delete(m.installations, key)
	return nil
}

type mockAuditRepo struct {
	entries []*repository.AuditEntry
	logErr  error
}

func (m *mockAuditRepo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	if m.logErr != nil {
		return m.logErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditRepo) Query(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	return m.entries, nil
}

type mockTransaction struct {
	installationRepo repository.InstallationRepository
	auditRepo        repository.AuditRepository
	commitErr        error
	rollbackErr      error
	committed        bool
	rolledBack       bool
}

func (m *mockTransaction) Commit() error {
	if m.commitErr != nil {
		return m.commitErr
	}
	m.committed = true
	return nil
}

func (m *mockTransaction) Rollback() error {
	if m.rollbackErr != nil {
		return m.rollbackErr
	}
	m.rolledBack = true
	return nil
}

func (m *mockTransaction) InstallationRepo() repository.InstallationRepository {
	return m.installationRepo
}

func (m *mockTransaction) CacheRepo() repository.CacheRepository {
	return nil
}

func (m *mockTransaction) AuditRepo() repository.AuditRepository {
	return m.auditRepo
}

func (m *mockTransaction) IndexRepo() repository.IndexRepository {
	return nil
}

type mockTransactionManager struct {
	tx       *mockTransaction
	beginErr error
}

func (m *mockTransactionManager) Begin(ctx context.Context) (transaction.Transaction, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return m.tx, nil
}

// Test CheckForUpdates

func TestUpdateManager_CheckForUpdates(t *testing.T) {
	tests := []struct {
		name          string
		installations []*repository.Installation
		backends      map[string]backend.Backend
		config        *config.Config
		wantUpdates   int
		wantErr       bool
	}{
		{
			name: "single tool with update available",
			installations: []*repository.Installation{
				{
					Tool:    "node",
					Version: "18.0.0",
					Backend: "github",
				},
			},
			backends: map[string]backend.Backend{
				"github": &mockUpdateBackend{
					name: "github",
					versions: map[string]*backend.VersionInfo{
						"latest": {Version: "20.0.0"},
					},
				},
			},
			wantUpdates: 1,
			wantErr:     false,
		},
		{
			name: "tool already at latest version",
			installations: []*repository.Installation{
				{
					Tool:    "node",
					Version: "20.0.0",
					Backend: "github",
				},
			},
			backends: map[string]backend.Backend{
				"github": &mockUpdateBackend{
					name: "github",
					versions: map[string]*backend.VersionInfo{
						"latest": {Version: "20.0.0"},
					},
				},
			},
			wantUpdates: 1,
			wantErr:     false,
		},
		{
			name: "respect version constraints from config",
			installations: []*repository.Installation{
				{
					Tool:    "node",
					Version: "18.0.0",
					Backend: "github",
				},
			},
			backends: map[string]backend.Backend{
				"github": &mockUpdateBackend{
					name: "github",
					versions: map[string]*backend.VersionInfo{
						"latest": {Version: "20.0.0"},
						"18.5.0": {Version: "18.5.0"},
					},
				},
			},
			config: &config.Config{
				Tools: map[string]config.ToolConfig{
					"node": {Version: "18.5.0"},
				},
			},
			wantUpdates: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installRepo := &mockInstallationRepo{
				installations: make(map[string]*repository.Installation),
			}
			for _, inst := range tt.installations {
				key := inst.Tool + "-" + inst.Version
				installRepo.installations[key] = inst
			}

			backendRegistry := backend.NewRegistry()
			for name, b := range tt.backends {
				backendRegistry.Register(b)
				_ = name // avoid unused variable
			}

			providerRegistry := provider.NewRegistry()
			downloadManager := download.NewManager()

			um := NewUpdateManager(
				backendRegistry,
				providerRegistry,
				downloadManager,
				installRepo,
				&mockAuditRepo{},
				&mockTransactionManager{},
				tt.config,
			)

			updates, err := um.CheckForUpdates(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, updates, tt.wantUpdates)
		})
	}
}

// Test UpdateTool

func TestUpdateManager_UpdateTool(t *testing.T) {
	tests := []struct {
		name          string
		tool          string
		targetVersion string
		installation  *repository.Installation
		backend       backend.Backend
		provider      provider.Provider
		downloader    download.Downloader
		wantSuccess   bool
		wantErr       bool
	}{
		{
			name:          "successful update",
			tool:          "node",
			targetVersion: "20.0.0",
			installation: &repository.Installation{
				Tool:        "node",
				Version:     "18.0.0",
				Backend:     "github",
				InstallPath: "/opt/unirtm/tools/node/18.0.0",
			},
			backend: &mockUpdateBackend{
				name: "github",
				versions: map[string]*backend.VersionInfo{
					"20.0.0": {
						Version:     "20.0.0",
						DownloadURL: "https://example.com/node-20.0.0.tar.gz",
						Checksum:    "abc123",
					},
				},
			},
			provider:    &mockProvider{name: "node"},
			downloader:  &mockDownloader{},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:          "already at target version",
			tool:          "node",
			targetVersion: "18.0.0",
			installation: &repository.Installation{
				Tool:        "node",
				Version:     "18.0.0",
				Backend:     "github",
				InstallPath: "/opt/unirtm/tools/node/18.0.0",
			},
			backend: &mockUpdateBackend{
				name: "github",
			},
			provider:    &mockProvider{name: "node"},
			downloader:  &mockDownloader{},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name:          "download failure triggers rollback",
			tool:          "node",
			targetVersion: "20.0.0",
			installation: &repository.Installation{
				Tool:        "node",
				Version:     "18.0.0",
				Backend:     "github",
				InstallPath: "/opt/unirtm/tools/node/18.0.0",
			},
			backend: &mockUpdateBackend{
				name: "github",
				versions: map[string]*backend.VersionInfo{
					"20.0.0": {
						Version:     "20.0.0",
						DownloadURL: "https://example.com/node-20.0.0.tar.gz",
					},
				},
			},
			provider:    &mockProvider{name: "node"},
			downloader:  &mockDownloader{downloadErr: errors.New("network error")},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name:          "checksum verification failure",
			tool:          "node",
			targetVersion: "20.0.0",
			installation: &repository.Installation{
				Tool:        "node",
				Version:     "18.0.0",
				Backend:     "github",
				InstallPath: "/opt/unirtm/tools/node/18.0.0",
			},
			backend: &mockUpdateBackend{
				name: "github",
				versions: map[string]*backend.VersionInfo{
					"20.0.0": {
						Version:     "20.0.0",
						DownloadURL: "https://example.com/node-20.0.0.tar.gz",
						Checksum:    "abc123",
					},
				},
			},
			provider:    &mockProvider{name: "node"},
			downloader:  &mockDownloader{verifyChecksumErr: errors.New("checksum mismatch")},
			wantSuccess: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installRepo := &mockInstallationRepo{
				installations: make(map[string]*repository.Installation),
			}
			key := tt.installation.Tool + "-" + tt.installation.Version
			installRepo.installations[key] = tt.installation

			backendRegistry := backend.NewRegistry()
			backendRegistry.Register(tt.backend)

			providerRegistry := provider.NewRegistry()
			providerRegistry.Register(tt.tool, tt.provider)

			downloadManager := download.NewManager()
			downloadManager.Register("https", tt.downloader)

			auditRepo := &mockAuditRepo{}

			txManager := &mockTransactionManager{
				tx: &mockTransaction{
					installationRepo: installRepo,
					auditRepo:        auditRepo,
				},
			}

			um := NewUpdateManager(
				backendRegistry,
				providerRegistry,
				downloadManager,
				installRepo,
				auditRepo,
				txManager,
				nil,
			)

			result, err := um.UpdateTool(context.Background(), tt.tool, tt.targetVersion)

			if tt.wantErr {
				require.Error(t, err)
				if result != nil {
					assert.False(t, result.Success)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantSuccess, result.Success)
			assert.Equal(t, tt.tool, result.Tool)
			assert.Equal(t, tt.targetVersion, result.NewVersion)
		})
	}
}

// Test PreviewUpdates

func TestUpdateManager_PreviewUpdates(t *testing.T) {
	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"node-18.0.0": {
				Tool:    "node",
				Version: "18.0.0",
				Backend: "github",
			},
			"python-3.10.0": {
				Tool:    "python",
				Version: "3.10.0",
				Backend: "github",
			},
		},
	}

	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "github",
		versions: map[string]*backend.VersionInfo{
			"latest": {Version: "20.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	um := NewUpdateManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		&mockAuditRepo{},
		&mockTransactionManager{},
		nil,
	)

	preview, err := um.PreviewUpdates(context.Background())

	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Equal(t, 2, preview.TotalUpdates)
	assert.Len(t, preview.Updates, 2)
	assert.Greater(t, preview.EstimatedTime, time.Duration(0))
}

// Test EnableAutomaticUpdates

func TestUpdateManager_EnableAutomaticUpdates(t *testing.T) {
	auditRepo := &mockAuditRepo{}

	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	um := NewUpdateManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		&mockInstallationRepo{},
		auditRepo,
		&mockTransactionManager{},
		nil,
	)

	err := um.EnableAutomaticUpdates(context.Background(), "0 2 * * *")

	require.NoError(t, err)
	assert.Len(t, auditRepo.entries, 1)
	assert.Equal(t, "enable_automatic_updates", auditRepo.entries[0].Operation)
	assert.Equal(t, "success", auditRepo.entries[0].Status)
}

// Test DisableAutomaticUpdates

func TestUpdateManager_DisableAutomaticUpdates(t *testing.T) {
	auditRepo := &mockAuditRepo{}

	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	um := NewUpdateManager(
		backendRegistry,
		providerRegistry,
		downloadManager,
		&mockInstallationRepo{},
		auditRepo,
		&mockTransactionManager{},
		nil,
	)

	err := um.DisableAutomaticUpdates(context.Background())

	require.NoError(t, err)
	assert.Len(t, auditRepo.entries, 1)
	assert.Equal(t, "disable_automatic_updates", auditRepo.entries[0].Operation)
	assert.Equal(t, "success", auditRepo.entries[0].Status)
}
