// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"strings"
)

// ContainerBackend implements the Backend interface for container tools.
// It handles tools registered with prefixes like "docker", "podman", etc.
type ContainerBackend struct {
	name string
}

// NewContainerBackend creates a new Container backend.
func NewContainerBackend(name string) *ContainerBackend {
	return &ContainerBackend{name: name}
}

// Name returns the backend identifier.
func (b *ContainerBackend) Name() string {
	return b.name
}

// Dependencies returns any external dependencies required by this backend.
func (b *ContainerBackend) Dependencies() []string {
	// Provider will check for podman/docker/nerdctl.
	// Backend itself doesn't have hard dependencies.
	return nil
}

// ListVersions returns available versions for a container tool.
func (b *ContainerBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Fetching tags from a generic registry API requires robust auth and varied API handling.
	// For now, we return "latest" to allow basic resolution if unspecified.
	return []VersionInfo{
		{Version: "latest"},
	}, nil
}

// ResolveVersion resolves a version request.
func (b *ContainerBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "" {
		versionRequest = "latest"
	}

	return b.GetDownloadInfo(ctx, tool, versionRequest, platform)
}

// GetDownloadInfo returns metadata needed by the ContainerProvider.
func (b *ContainerBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// The `tool` parameter contains the registry and image path, e.g. "ghcr.io/aquasec/trivy".
	// The prefix like "docker:" has already been stripped by UniRTM config parsing.
	
	// If the version is a digest (e.g., sha256:abcd...), treat it as Checksum and use it in the URL construction if needed.
	checksum := ""
	if strings.HasPrefix(version, "sha256:") {
		checksum = version
	}

	return &VersionInfo{
		Version:     version,
		// DownloadURL is deliberately empty because this backend does not download an artifact over HTTP.
		// Instead, the provider handles `docker pull`.
		DownloadURL: "",
		Checksum:    checksum,
		Platform:    platform,
		Metadata: map[string]string{
			"backend": "container",
			"image":   tool,
			"tag":     version,
		},
	}, nil
}

func (b *ContainerBackend) SupportsChecksum() bool {
	// Container backend natively supports digests, which act as checksums.
	return true
}

func (b *ContainerBackend) SupportsGPG() bool {
	return false
}

func (b *ContainerBackend) AttestationType() string {
	return ""
}

func (b *ContainerBackend) IsRecommended() bool {
	return true
}

func (b *ContainerBackend) IsScriptless() bool {
	return false // We generate a wrapper script
}

func (b *ContainerBackend) GetReach() string {
	return "Large"
}

func (b *ContainerBackend) IsStable() bool {
	return true
}

func (b *ContainerBackend) SupportsOffline() bool {
	return false
}
