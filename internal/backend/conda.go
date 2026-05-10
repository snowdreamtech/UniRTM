// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
)

// CondaBackend implements the Backend interface for Conda packages.
type CondaBackend struct {
}

// NewCondaBackend creates a new conda backend.
func NewCondaBackend() *CondaBackend {
	return &CondaBackend{}
}

func (b *CondaBackend) Name() string {
	return "conda"
}

func (b *CondaBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Conda repositories are massive and difficult to parse directly via HTTP without the CLI.
	// For simplicity, we rely on the user to provide explicit versions, or we defer to conda search.
	return nil, NewBackendError(b.Name(), tool, "listing versions is not supported via REST for conda", nil)
}

func (b *CondaBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		// Conda install will automatically resolve the latest version if no version is provided.
		// However, in UniRTM we need a concrete string. Let's let the provider handle "latest" natively,
		// or we can just pass it through.
		return &VersionInfo{
			Version:  "latest",
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CondaBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CondaBackend) SupportsChecksum() bool {
	return false
}

func (b *CondaBackend) SupportsGPG() bool {
	return false
}
