// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
)

// VfoxBackend implements the Backend interface for vfox plugins.
type VfoxBackend struct {
}

// NewVfoxBackend creates a new vfox backend.
func NewVfoxBackend() *VfoxBackend {
	return &VfoxBackend{}
}

func (b *VfoxBackend) Name() string {
	return "vfox"
}

func (b *VfoxBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Rely on explicit requests or local vfox CLI output for full listing.
	return nil, NewBackendError(b.Name(), tool, "vfox version listing via API is not natively supported without a Lua VM or CLI wrapper", nil)
}

func (b *VfoxBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
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

func (b *VfoxBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *VfoxBackend) SupportsChecksum() bool {
	return false
}

func (b *VfoxBackend) SupportsGPG() bool {
	return false
}

func (b *VfoxBackend) AttestationType() string {
	return ""
}
