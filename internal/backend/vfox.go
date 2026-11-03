// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var (
	vfoxCache  sync.Map
	vfoxFlight singleflight.Group
)

// ClearVfoxCache clears both in-memory and disk cache for vfox. Mainly used for testing.
func ClearVfoxCache() {
	vfoxCache.Range(func(key, _ interface{}) bool {
		vfoxCache.Delete(key)
		return true
	})
	ClearEcosystemMetadataDiskCache("vfox")
}

// VfoxBackend implements the Backend interface for vfox plugins.
type VfoxBackend struct{}

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
	if val, ok := vfoxCache.Load(tool); ok {
		cached := val.([]VersionInfo)
		res := make([]VersionInfo, len(cached))
		copy(res, cached)
		for i := range res {
			res[i].Platform = platform
		}
		return res, nil
	}

	res, err, _ := vfoxFlight.Do(tool, func() (interface{}, error) {
		if val, ok := vfoxCache.Load(tool); ok {
			return val.([]VersionInfo), nil
		}

		var diskVersions []string
		if readEcosystemMetadataDiskCache("vfox", tool, &diskVersions, 10*time.Minute) {
			var versions []VersionInfo
			for _, v := range diskVersions {
				versions = append(versions, VersionInfo{
					Version:  v,
					Platform: platform,
				})
			}
			vfoxCache.Store(tool, versions)
			return versions, nil
		}

		// Since we don't have a Lua VM to run vfox plugins, we shell out to vfox if installed.
		cmd := exec.CommandContext(ctx, "vfox", "list", "all", tool)
		out, err := cmd.Output()
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "vfox list all failed (ensure vfox is installed)", err)
		}

		var versions []VersionInfo
		var versionStrs []string
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			v := strings.TrimSpace(line)
			if v == "" || strings.Contains(v, "Available versions") {
				continue
			}
			// vfox output can be messy, we try to grab the first word which is usually the version
			parts := strings.Fields(v)
			if len(parts) > 0 {
				ver := parts[0]
				versions = append(versions, VersionInfo{
					Version:  ver,
					Platform: platform,
				})
				versionStrs = append(versionStrs, ver)
			}
		}

		writeEcosystemMetadataDiskCache("vfox", tool, versionStrs)
		vfoxCache.Store(tool, versions)
		return versions, nil
	})

	if err != nil {
		return nil, err
	}

	cached := res.([]VersionInfo)
	out := make([]VersionInfo, len(cached))
	copy(out, cached)
	for i := range out {
		out[i].Platform = platform
	}
	return out, nil
}

func (b *VfoxBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *VfoxBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
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
