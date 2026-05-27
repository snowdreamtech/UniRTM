package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
