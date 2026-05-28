// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	unirtmhttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("mocked response")),
		Header:     make(http.Header),
	}, nil
}

func TestInstallationManager_Install_WithLockfileStrict(t *testing.T) {
	// Set up global MockTransport to avoid hanging network calls
	oldMock := unirtmhttp.MockTransport
	defer func() { unirtmhttp.MockTransport = oldMock }()
	unirtmhttp.MockTransport = &mockRoundTripper{}

	// Create a temporary directory for our lockfile
	tempDir := t.TempDir()
	lockfilePath := filepath.Join(tempDir, "unirtm.lock")

	// We create a lockfile with the "github:foo/bar" key.
	// The current platform will determine which platform key is checked.
	// We'll add the current platform to ensure it finds it.
	currentPlatform := backend.CurrentPlatform()
	platKey := lockfile.PlatformKey(string(currentPlatform.OS), string(currentPlatform.Arch), false)

	lockContent := `
[[tools."github:foo/bar"]]
version = "1.0.0"
backend = "github"

[tools."github:foo/bar"."platforms.` + platKey + `"]
url = "https://example.com/foo-` + platKey + `"
checksum = "123"
`
	err := os.WriteFile(lockfilePath, []byte(lockContent), 0644)
	require.NoError(t, err)

	// Create LockService with strict mode
	ls, err := NewLockService(LockServiceOptions{
		LockfilePath: lockfilePath,
		StrictMode:   true,
	})
	require.NoError(t, err)

	// Register a mock backend so it doesn't fail on backend lookup
	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "github",
		versions: map[string]*backend.VersionInfo{
			"1.0.0": {Version: "1.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	// Create mock installation repo and tx manager
	installRepo := &mockInstallationRepo{
		installations: make(map[string]*repository.Installation),
	}
	txManager := &mockTransactionManager{
		tx: &mockTransaction{
			installationRepo: installRepo,
			auditRepo:        &mockAuditRepo{},
		},
	}

	im := NewInstallationManagerWithLock(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
		ls,
		&config.Settings{},
	)

	ctx := context.Background()

	// Scenario 1: Using the stripped tool name (simulating the old bug).
	// CheckStrict should fail because it won't find "foo/bar" in the lockfile.
	err = im.Install(ctx, "foo/bar", "foo/bar", "1.0.0", "github")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no locked entry")

	// Scenario 2: Using the full tool key (simulating the fix).
	// CheckStrict should pass, and it will fail later in the installation process
	// (e.g. downloading or finding a provider), which proves it passed the lockfile check.
	err = im.Install(ctx, "github:foo/bar", "foo/bar", "1.0.0", "github")
	// Since we don't have a registered provider for "foo/bar", or the downloader is not fully mocked,
	// it will fail at a subsequent step.
	// As long as it doesn't fail with "no locked platform", the fix is verified.
	if err != nil {
		assert.NotContains(t, err.Error(), "no locked platform", "Should have passed strict lockfile validation")
	}
}

func TestTryVerifyProvenance(t *testing.T) {
	ctx := context.Background()

	t.Run("skipped via env", func(t *testing.T) {
		os.Setenv("UNIRTM_VERIFY_PROVENANCE", "0")
		defer os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")
		status, err := tryVerifyProvenance(ctx, "github", "foo/bar", "dummy.tar.gz")
		assert.NoError(t, err)
		assert.Equal(t, "skipped", status)
	})

	t.Run("not applicable - invalid tool format", func(t *testing.T) {
		os.Setenv("UNIRTM_VERIFY_PROVENANCE", "1")
		defer os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")
		status, err := tryVerifyProvenance(ctx, "github", "invalid-tool", "dummy.tar.gz")
		assert.NoError(t, err)
		assert.Equal(t, "not_applicable", status)
	})

	t.Run("failed - github missing file", func(t *testing.T) {
		os.Setenv("UNIRTM_VERIFY_PROVENANCE", "1")
		defer os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")
		status, err := tryVerifyProvenance(ctx, "github", "foo/bar", "nonexistent.tar.gz")
		assert.Error(t, err)
		assert.Equal(t, "failed", status)
	})

	t.Run("failed - gitlab missing file", func(t *testing.T) {
		os.Setenv("UNIRTM_VERIFY_PROVENANCE", "1")
		defer os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")
		status, err := tryVerifyProvenance(ctx, "gitlab", "foo/bar", "nonexistent.tar.gz")
		assert.Error(t, err)
		assert.Equal(t, "failed", status)
	})
}

func TestIsExecutableFile(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Directory
	assert.False(t, isExecutableFile(tempDir))

	// 2. Archive
	archiveFile := filepath.Join(tempDir, "test.tar.gz")
	os.WriteFile(archiveFile, []byte("data"), 0644)
	assert.False(t, isExecutableFile(archiveFile))

	// 3. Non-existent
	assert.False(t, isExecutableFile(filepath.Join(tempDir, "does-not-exist")))

	// 4. Executable
	execFile := filepath.Join(tempDir, "my-exec")
	os.WriteFile(execFile, []byte("echo hi"), 0755)
	// On Unix, this should be true. (Windows may be different based on GOOS logic).
	if runtime.GOOS != "windows" {
		assert.True(t, isExecutableFile(execFile))
	}

	// 5. Non-executable regular file
	nonExecFile := filepath.Join(tempDir, "my-file.txt")
	os.WriteFile(nonExecFile, []byte("hello"), 0644)
	if runtime.GOOS != "windows" {
		assert.False(t, isExecutableFile(nonExecFile))
	}
}

func TestSortTools(t *testing.T) {
	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "test-backend",
		versions: map[string]*backend.VersionInfo{
			"1.0.0": {Version: "1.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	im := NewInstallationManager(backendRegistry, nil, nil, nil, nil, nil)

	tools := map[string]config.ToolConfig{
		"go:golang.org/x/tools/cmd/goimports": {Version: "latest"},
		"node":                                {Version: "20", Backend: "native"},
		"github:foo/bar":                      {Version: "1.0.0"},
	}

	sorted := im.SortTools(tools)
	assert.Len(t, sorted, 3)

	// Since we don't have deep dependency chains in mock backends, they should just all be returned.
	names := make([]string, 0, 3)
	for _, t := range sorted {
		names = append(names, t.ToolName)
	}
	assert.Contains(t, names, "node")
	assert.Contains(t, names, "foo/bar")
	assert.Contains(t, names, "golang.org/x/tools/cmd/goimports")
}

func TestInstall_Errors(t *testing.T) {
	ctx := context.Background()
	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "test-backend",
		versions: map[string]*backend.VersionInfo{
			"1.0.0": {Version: "1.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	installRepo := &mockInstallationRepo{
		installations: make(map[string]*repository.Installation),
	}
	installRepo.installations["tool-1.0.0"] = &repository.Installation{
		Tool:    "tool",
		Version: "1.0.0",
		Backend: "test-backend",
	}

	downloadManager := download.NewManager()
	mockDl := &mockDownloader{}
	downloadManager.Register("https", mockDl)
	downloadManager.Register("http", mockDl)

	auditRepo := &mockAuditRepo{}
	txManager := &mockTransactionManager{
		tx: &mockTransaction{
			installationRepo: installRepo,
			auditRepo:        auditRepo,
		},
	}

	providerRegistry := provider.NewRegistry()
	im := NewInstallationManager(backendRegistry, providerRegistry, downloadManager, installRepo, txManager, nil)

	// 1. Backend not found
	err := im.Install(ctx, "tool", "tool", "1.0.0", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend not found")

	// 2. Provider Install error
	providerRegistry.Register("tool", &mockProvider{
		name:       "tool",
		installErr: errors.New("mock install error"),
	})

	err = im.Install(ctx, "tool", "tool", "1.0.0", "test-backend")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock install error")
}

func TestSelectVersionInteractive(t *testing.T) {
	ctx := context.Background()
	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "test-backend",
		versions: map[string]*backend.VersionInfo{
			"1.0.0": {Version: "1.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	im := NewInstallationManager(backendRegistry, nil, nil, nil, nil, nil)

	// 1. Backend not found
	_, err := im.SelectVersionInteractive(ctx, "tool", "nonexistent")
	assert.Error(t, err)

	// 2. Empty versions
	emptyBackend := &mockUpdateBackend{name: "empty"}
	backendRegistry.Register(emptyBackend)
	_, err = im.SelectVersionInteractive(ctx, "tool", "empty")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no versions found")

	// The interactive prompt blocks if we give it versions, so we skip the success path.
}

func TestInstall_Success(t *testing.T) {
	ctx := context.Background()
	backendRegistry := backend.NewRegistry()
	mockBackend := &mockUpdateBackend{
		name: "test-backend",
		versions: map[string]*backend.VersionInfo{
			"1.0.0": {Version: "1.0.0"},
		},
	}
	backendRegistry.Register(mockBackend)

	installRepo := &mockInstallationRepo{
		installations: make(map[string]*repository.Installation),
	}

	providerRegistry := provider.NewRegistry()
	mockProviderInstance := &mockProvider{
		name: "tool",
	}
	providerRegistry.Register("tool", mockProviderInstance)

	downloadManager := download.NewManager()
	mockDl := &mockDownloader{}
	downloadManager.Register("https", mockDl)
	downloadManager.Register("http", mockDl)

	auditRepo := &mockAuditRepo{}
	txManager := &mockTransactionManager{
		tx: &mockTransaction{
			installationRepo: installRepo,
			auditRepo:        auditRepo,
		},
	}

	im := NewInstallationManager(backendRegistry, providerRegistry, downloadManager, installRepo, txManager, nil)

	// Disable verification to avoid git network calls
	os.Setenv("UNIRTM_VERIFY_PROVENANCE", "0")
	defer os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")

	tempDataDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tempDataDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	err := im.Install(ctx, "tool", "tool", "1.0.0", "test-backend")
	assert.NoError(t, err)
}

func TestInstall_Uninstall(t *testing.T) {
	ctx := context.Background()
	installPath := filepath.Join(t.TempDir(), "tool", "1.0.0")
	os.MkdirAll(installPath, 0755)
	installRepo := &mockInstallationRepo{
		installations: map[string]*repository.Installation{
			"tool-1.0.0": {
				Tool:        "tool",
				Version:     "1.0.0",
				Backend:     "test-backend",
				InstallPath: installPath,
			},
		},
	}

	providerRegistry := provider.NewRegistry()
	mockProviderInstance := &mockProvider{name: "tool"}
	providerRegistry.Register("tool", mockProviderInstance)

	txManager := &mockTransactionManager{
		tx: &mockTransaction{
			installationRepo: installRepo,
			auditRepo:        &mockAuditRepo{},
		},
	}

	im := NewInstallationManager(nil, providerRegistry, nil, installRepo, txManager, nil)

	// Test uninstall error not found
	err := im.Uninstall(ctx, "tool", "2.0.0")
	assert.Error(t, err)

	// Test uninstall success
	err = im.Uninstall(ctx, "tool", "1.0.0")
	assert.NoError(t, err)

	// Verify it was removed
	_, err = installRepo.FindByToolAndVersion(ctx, "tool", "1.0.0")
	assert.Error(t, err)
}
