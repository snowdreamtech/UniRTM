// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
)

// SpmBackend implements the Backend interface for Swift Package Manager.
type SpmBackend struct {
}

// NewSpmBackend creates a new spm backend.
func NewSpmBackend() *SpmBackend {
	return &SpmBackend{}
}

func (b *SpmBackend) Name() string {
	return "spm"
}

func (b *SpmBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// For SPM, tool is usually a git repo. Fetching versions requires git ls-remote.
	// For now, we rely on explicit versions passed by the user.
	return nil, NewBackendError(b.Name(), tool, "spm version listing requires git ls-remote, which is not implemented in the backend API natively", nil)
}

func (b *SpmBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *SpmBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *SpmBackend) SupportsChecksum() bool {
	return true
}

func (b *SpmBackend) SupportsGPG() bool {
	return false
}

func (b *SpmBackend) AttestationType() string {
	return ""
}

func (b *SpmBackend) IsRecommended() bool {
	return true
}

func (b *SpmBackend) IsScriptless() bool {
	return true
}

func (b *SpmBackend) GetReach() string {
	return "Large"
}

func (b *SpmBackend) IsStable() bool {
	return true
}

func (b *SpmBackend) SupportsOffline() bool {
	return true
}
