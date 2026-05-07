// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackend is a mock implementation of the Backend interface for testing
type mockBackend struct {
	name                string
	listVersionsFunc    func(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error)
	resolveVersionFunc  func(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error)
	getDownloadInfoFunc func(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error)
	supportsChecksum    bool
	supportsGPG         bool
}

func (m *mockBackend) Name() string {
	return m.name
}

func (m *mockBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	if m.listVersionsFunc != nil {
		return m.listVersionsFunc(ctx, tool, platform)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	if m.resolveVersionFunc != nil {
		return m.resolveVersionFunc(ctx, tool, versionRequest, platform)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	if m.getDownloadInfoFunc != nil {
		return m.getDownloadInfoFunc(ctx, tool, version, platform)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBackend) SupportsChecksum() bool {
	return m.supportsChecksum
}

func (m *mockBackend) SupportsGPG() bool {
	return m.supportsGPG
}

func TestNewVersionManager(t *testing.T) {
	backends := map[string]backend.Backend{
		"github": &mockBackend{name: "github"},
		"aqua":   &mockBackend{name: "aqua"},
	}

	vm := NewVersionManager(backends)

	require.NotNil(t, vm)
	assert.Equal(t, backends, vm.backends)
}

func TestVersionManager_ParseVersionConstraint(t *testing.T) {
	vm := NewVersionManager(nil)

	tests := []struct {
		name        string
		versionStr  string
		wantType    VersionType
		wantErr     bool
		errContains string
	}{
		{
			name:       "exact version",
			versionStr: "1.20.0",
			wantType:   VersionTypeExact,
			wantErr:    false,
		},
		{
			name:       "exact version with v prefix",
			versionStr: "v1.20.0",
			wantType:   VersionTypeExact,
			wantErr:    false,
		},
		{
			name:       "range with >=",
			versionStr: ">=1.20.0",
			wantType:   VersionTypeRange,
			wantErr:    false,
		},
		{
			name:       "caret range",
			versionStr: "^3.11.0",
			wantType:   VersionTypeRange,
			wantErr:    false,
		},
		{
			name:       "tilde range",
			versionStr: "~2.7.0",
			wantType:   VersionTypeRange,
			wantErr:    false,
		},
		{
			name:       "latest alias",
			versionStr: "latest",
			wantType:   VersionTypeAlias,
			wantErr:    false,
		},
		{
			name:       "lts alias",
			versionStr: "lts",
			wantType:   VersionTypeAlias,
			wantErr:    false,
		},
		{
			name:       "stable alias",
			versionStr: "stable",
			wantType:   VersionTypeAlias,
			wantErr:    false,
		},
		{
			name:        "empty version - explicit requirement",
			versionStr:  "",
			wantErr:     true,
			errContains: "version specification is required",
		},
		{
			name:        "invalid version",
			versionStr:  "invalid",
			wantErr:     true,
			errContains: "invalid version string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := vm.ParseVersionConstraint(tt.versionStr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, version)
			assert.Equal(t, tt.wantType, version.Type)
		})
	}
}

func TestVersionManager_ResolveVersion_ExactVersion(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		getDownloadInfoFunc: func(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
			return &backend.VersionInfo{
				Version:     version,
				DownloadURL: "https://example.com/tool-" + version + ".tar.gz",
				Checksum:    "abc123",
				Platform:    platform,
			}, nil
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "1.20.0", platform)

	require.NoError(t, err)
	require.NotNil(t, versionInfo)
	assert.Equal(t, "1.20.0", versionInfo.Version)
	assert.Equal(t, "https://example.com/tool-1.20.0.tar.gz", versionInfo.DownloadURL)
	assert.Equal(t, "abc123", versionInfo.Checksum)
}

func TestVersionManager_ResolveVersion_Alias(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		resolveVersionFunc: func(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
			// Simulate resolving "latest" to a concrete version
			if versionRequest == "latest" {
				return &backend.VersionInfo{
					Version:     "2.1.0",
					DownloadURL: "https://example.com/tool-2.1.0.tar.gz",
					Checksum:    "def456",
					Platform:    platform,
				}, nil
			}
			return nil, errors.New("unsupported version request")
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "latest", platform)

	require.NoError(t, err)
	require.NotNil(t, versionInfo)
	assert.Equal(t, "2.1.0", versionInfo.Version)
	assert.Equal(t, "https://example.com/tool-2.1.0.tar.gz", versionInfo.DownloadURL)
}

func TestVersionManager_ResolveVersion_Range(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		resolveVersionFunc: func(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
			// Simulate resolving ">=1.20.0" to the latest matching version
			if versionRequest == ">=1.20.0" {
				return &backend.VersionInfo{
					Version:     "1.22.5",
					DownloadURL: "https://example.com/tool-1.22.5.tar.gz",
					Checksum:    "ghi789",
					Platform:    platform,
				}, nil
			}
			return nil, errors.New("unsupported version request")
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", ">=1.20.0", platform)

	require.NoError(t, err)
	require.NotNil(t, versionInfo)
	assert.Equal(t, "1.22.5", versionInfo.Version)
}

func TestVersionManager_ResolveVersion_EmptyVersion(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	vm := NewVersionManager(map[string]backend.Backend{
		"github": &mockBackend{name: "github"},
	})

	_, err := vm.ResolveVersion(ctx, "github", "node", "", platform)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "explicit version specification required")
	assert.Contains(t, err.Error(), "node")
}

func TestVersionManager_ResolveVersion_BackendNotFound(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	vm := NewVersionManager(map[string]backend.Backend{
		"github": &mockBackend{name: "github"},
	})

	_, err := vm.ResolveVersion(ctx, "nonexistent", "node", "1.20.0", platform)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
}

func TestVersionManager_ResolveVersion_InvalidVersion(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	vm := NewVersionManager(map[string]backend.Backend{
		"github": &mockBackend{name: "github"},
	})

	_, err := vm.ResolveVersion(ctx, "github", "node", "invalid-version", platform)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse version constraint")
}

func TestVersionManager_ValidateVersionConstraint(t *testing.T) {
	vm := NewVersionManager(nil)

	tests := []struct {
		name        string
		versionStr  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid exact version",
			versionStr: "1.20.0",
			wantErr:    false,
		},
		{
			name:       "valid range",
			versionStr: ">=1.20.0",
			wantErr:    false,
		},
		{
			name:       "valid caret",
			versionStr: "^3.11.0",
			wantErr:    false,
		},
		{
			name:       "valid tilde",
			versionStr: "~2.7.0",
			wantErr:    false,
		},
		{
			name:       "valid alias",
			versionStr: "latest",
			wantErr:    false,
		},
		{
			name:        "empty version",
			versionStr:  "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "invalid version",
			versionStr:  "not-a-version",
			wantErr:     true,
			errContains: "invalid version constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateVersionConstraint(tt.versionStr)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestVersionManager_ListAvailableVersions(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	expectedVersions := []backend.VersionInfo{
		{Version: "2.1.0", DownloadURL: "https://example.com/tool-2.1.0.tar.gz"},
		{Version: "2.0.0", DownloadURL: "https://example.com/tool-2.0.0.tar.gz"},
		{Version: "1.20.0", DownloadURL: "https://example.com/tool-1.20.0.tar.gz"},
	}

	mockBackend := &mockBackend{
		name: "github",
		listVersionsFunc: func(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
			return expectedVersions, nil
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versions, err := vm.ListAvailableVersions(ctx, "github", "node", platform)

	require.NoError(t, err)
	require.Len(t, versions, 3)
	assert.Equal(t, expectedVersions, versions)
}

func TestVersionManager_ListAvailableVersions_BackendNotFound(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	vm := NewVersionManager(map[string]backend.Backend{
		"github": &mockBackend{name: "github"},
	})

	_, err := vm.ListAvailableVersions(ctx, "nonexistent", "node", platform)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
}

func TestVersionManager_SupportsChecksum(t *testing.T) {
	backends := map[string]backend.Backend{
		"github": &mockBackend{name: "github", supportsChecksum: true},
		"http":   &mockBackend{name: "http", supportsChecksum: false},
	}

	vm := NewVersionManager(backends)

	t.Run("backend supports checksum", func(t *testing.T) {
		supports, err := vm.SupportsChecksum("github")
		require.NoError(t, err)
		assert.True(t, supports)
	})

	t.Run("backend does not support checksum", func(t *testing.T) {
		supports, err := vm.SupportsChecksum("http")
		require.NoError(t, err)
		assert.False(t, supports)
	})

	t.Run("backend not found", func(t *testing.T) {
		_, err := vm.SupportsChecksum("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
	})
}

func TestVersionManager_SupportsGPG(t *testing.T) {
	backends := map[string]backend.Backend{
		"github": &mockBackend{name: "github", supportsGPG: true},
		"http":   &mockBackend{name: "http", supportsGPG: false},
	}

	vm := NewVersionManager(backends)

	t.Run("backend supports GPG", func(t *testing.T) {
		supports, err := vm.SupportsGPG("github")
		require.NoError(t, err)
		assert.True(t, supports)
	})

	t.Run("backend does not support GPG", func(t *testing.T) {
		supports, err := vm.SupportsGPG("http")
		require.NoError(t, err)
		assert.False(t, supports)
	})

	t.Run("backend not found", func(t *testing.T) {
		_, err := vm.SupportsGPG("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
	})
}

func TestVersionManager_ResolveVersion_BackendError(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		getDownloadInfoFunc: func(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
			return nil, errors.New("backend error: API rate limit exceeded")
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	_, err := vm.ResolveVersion(ctx, "github", "node", "1.20.0", platform)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get download info")
	assert.Contains(t, err.Error(), "backend error")
}

func TestVersionManager_ResolveVersion_CaretRange(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		resolveVersionFunc: func(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
			if versionRequest == "^1.20.0" {
				return &backend.VersionInfo{
					Version:     "1.22.5",
					DownloadURL: "https://example.com/tool-1.22.5.tar.gz",
					Checksum:    "abc123",
					Platform:    platform,
				}, nil
			}
			return nil, errors.New("unsupported version request")
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "^1.20.0", platform)

	require.NoError(t, err)
	require.NotNil(t, versionInfo)
	assert.Equal(t, "1.22.5", versionInfo.Version)
}

func TestVersionManager_ResolveVersion_TildeRange(t *testing.T) {
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	mockBackend := &mockBackend{
		name: "github",
		resolveVersionFunc: func(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
			if versionRequest == "~2.7.0" {
				return &backend.VersionInfo{
					Version:     "2.7.18",
					DownloadURL: "https://example.com/tool-2.7.18.tar.gz",
					Checksum:    "def456",
					Platform:    platform,
				}, nil
			}
			return nil, errors.New("unsupported version request")
		},
	}

	backends := map[string]backend.Backend{
		"github": mockBackend,
	}

	vm := NewVersionManager(backends)

	versionInfo, err := vm.ResolveVersion(ctx, "github", "python", "~2.7.0", platform)

	require.NoError(t, err)
	require.NotNil(t, versionInfo)
	assert.Equal(t, "2.7.18", versionInfo.Version)
}
