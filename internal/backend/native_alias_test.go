// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
)

type mockAliasNativeHandler struct{}

func (m *mockAliasNativeHandler) Name() string {
	return "mock-alias"
}

func (m *mockAliasNativeHandler) ResolveVersions(ctx context.Context, baseURL string) ([]native.VersionInfo, error) {
	return []native.VersionInfo{
		{
			Version: "22.1.0",
			IsLTS:   false,
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", URL: "https://example.com/node-v22.1.0-linux-x64.tar.gz", Filename: "node-v22.1.0-linux-x64.tar.gz"},
			},
		},
		{
			Version: "20.18.0",
			IsLTS:   true,
			LTSName: "Iron",
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", URL: "https://example.com/node-v20.18.0-linux-x64.tar.gz", Filename: "node-v20.18.0-linux-x64.tar.gz"},
			},
		},
		{
			Version: "18.20.4",
			IsLTS:   true,
			LTSName: "Hydrogen",
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", URL: "https://example.com/node-v18.20.4-linux-x64.tar.gz", Filename: "node-v18.20.4-linux-x64.tar.gz"},
			},
		},
	}, nil
}

func TestNativeBackend_DynamicLTSAndPrefixResolution(t *testing.T) {
	b := NewNativeBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	b.recipes["node-mock"] = native.Recipe{
		BaseURL: "https://example.com",
		Handler: &mockAliasNativeHandler{},
	}

	// 1. Test "latest" -> should resolve to 22.1.0
	info, err := b.ResolveVersion(ctx, "node-mock", "latest", platform)
	require.NoError(t, err)
	assert.Equal(t, "22.1.0", info.Version)

	// 2. Test "lts" -> should resolve to 20.18.0 (highest LTS)
	info, err = b.ResolveVersion(ctx, "node-mock", "lts", platform)
	require.NoError(t, err)
	assert.Equal(t, "20.18.0", info.Version)

	// 3. Test "lts/hydrogen" -> should resolve to 18.20.4
	info, err = b.ResolveVersion(ctx, "node-mock", "lts/hydrogen", platform)
	require.NoError(t, err)
	assert.Equal(t, "18.20.4", info.Version)

	// 4. Test prefix "20" -> should resolve to 20.18.0
	info, err = b.ResolveVersion(ctx, "node-mock", "20", platform)
	require.NoError(t, err)
	assert.Equal(t, "20.18.0", info.Version)

	// 5. Test prefix "20.x" -> should resolve to 20.18.0
	info, err = b.ResolveVersion(ctx, "node-mock", "20.x", platform)
	require.NoError(t, err)
	assert.Equal(t, "20.18.0", info.Version)

	// 6. Test leading 'v' normalization "v20.18.0" -> 20.18.0
	info, err = b.ResolveVersion(ctx, "node-mock", "v20.18.0", platform)
	require.NoError(t, err)
	assert.Equal(t, "20.18.0", info.Version)
}
