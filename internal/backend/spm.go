// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os/exec"
	"strings"
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
	// For SPM, tool is usually a git repo URL.
	// We use git ls-remote to fetch tags.
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", tool)
	out, err := cmd.Output()
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "git ls-remote failed", err)
	}

	var versions []VersionInfo
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		if strings.HasPrefix(ref, "refs/tags/") {
			v := strings.TrimPrefix(ref, "refs/tags/")
			v = strings.TrimSuffix(v, "^{}") // Remove peeled tag suffix
			versions = append(versions, VersionInfo{
				Version:  v,
				Platform: platform,
			})
		}
	}

	return versions, nil
}

func (b *SpmBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		// Rough sort by string (can be improved with semver)
		return &versions[len(versions)-1], nil
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
