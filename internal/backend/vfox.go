// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os/exec"
	"strings"
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

func (b *VfoxBackend) Dependencies() []string {
	return nil
}
func (b *VfoxBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Since we don't have a Lua VM to run vfox plugins, we shell out to vfox if installed.
	cmd := exec.CommandContext(ctx, "vfox", "list", "all", tool)
	out, err := cmd.Output()
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "vfox list all failed (ensure vfox is installed)", err)
	}

	var versions []VersionInfo
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		v := strings.TrimSpace(line)
		if v == "" || strings.Contains(v, "Available versions") {
			continue
		}
		// vfox output can be messy, we try to grab the first word which is usually the version
		parts := strings.Fields(v)
		if len(parts) > 0 {
			versions = append(versions, VersionInfo{
				Version:  parts[0],
				Platform: platform,
			})
		}
	}

	return versions, nil
}

func (b *VfoxBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		// vfox list all usually returns newest last or we can pick last
		return &versions[len(versions)-1], nil
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

func (b *VfoxBackend) IsRecommended() bool {
	return false
}

func (b *VfoxBackend) IsScriptless() bool {
	return false
}

func (b *VfoxBackend) GetReach() string {
	return "Huge"
}

func (b *VfoxBackend) IsStable() bool {
	return false
}

func (b *VfoxBackend) SupportsOffline() bool {
	return false
}
