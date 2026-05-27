package backend

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
	"github.com/stretchr/testify/assert"
)

func TestNativeBackend_MockRecipe(t *testing.T) {
	b := NewNativeBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Inject a mock recipe
	b.recipes["mock-tool"] = native.Recipe{
		BaseURL: "mock-url",
		Aliases: map[string]string{"latest": "2.0"},
		Handler: &mockNativeHandler{},
	}

	// Test ListVersions
	versions, err := b.ListVersions(ctx, "mock-tool", platform)
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "2.0", versions[0].Version)
	assert.Equal(t, "https://example.com/mock-2.0-linux-amd64.tar.gz", versions[0].DownloadURL)

	// Test ResolveVersion (latest alias)
	info, err := b.ResolveVersion(ctx, "mock-tool", "latest", platform)
	assert.NoError(t, err)
	assert.Equal(t, "2.0", info.Version)

	// Test ResolveVersion (specific version)
	info, err = b.ResolveVersion(ctx, "mock-tool", "1.0", platform)
	assert.NoError(t, err)
	assert.Equal(t, "1.0", info.Version)

	// Test GetDownloadInfo (no matching asset fallback score)
	info, err = b.GetDownloadInfo(ctx, "mock-tool", "1.0", platform)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/mock-1.0-fallback.tar.gz", info.DownloadURL)
}

type mockNativeHandler struct{}

func (m *mockNativeHandler) Name() string {
	return "mock"
}

func (m *mockNativeHandler) ResolveVersions(ctx context.Context, baseURL string) ([]native.VersionInfo, error) {
	return []native.VersionInfo{
		{
			Version: "2.0",
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", URL: "https://example.com/mock-2.0-linux-amd64.tar.gz"},
			},
		},
		{
			Version: "1.0",
			Assets: []native.Asset{
				{Filename: "mock-1.0-linux-amd64.tar.gz", URL: "https://example.com/mock-1.0-fallback.tar.gz"},
			},
		},
	}, nil
}
