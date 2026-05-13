// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
)

// PipxBackend implements the Backend interface for pipx packages.
// It leverages the PyPI backend for version resolution as pipx packages are hosted on PyPI.
type PipxBackend struct {
	pypi *PypiBackend
}

// NewPipxBackend creates a new pipx backend.
func NewPipxBackend() *PipxBackend {
	return &PipxBackend{
		pypi: NewPypiBackend(),
	}
}

func (b *PipxBackend) Name() string {
	return "pipx"
}

func (b *PipxBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	return b.pypi.ListVersions(ctx, tool, platform)
}

func (b *PipxBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return b.pypi.ResolveVersion(ctx, tool, versionRequest, platform)
}

func (b *PipxBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// Provider handles actual installation via pipx cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *PipxBackend) SupportsChecksum() bool {
	return true
}

func (b *PipxBackend) SupportsGPG() bool {
	return false
}

func (b *PipxBackend) AttestationType() string {
	return ""
}

func (b *PipxBackend) IsRecommended() bool {
	return true
}

func (b *PipxBackend) IsScriptless() bool {
	return true
}

func (b *PipxBackend) GetReach() string {
	return "Large"
}

func (b *PipxBackend) IsStable() bool {
	return true
}

func (b *PipxBackend) SupportsOffline() bool {
	return false
}
