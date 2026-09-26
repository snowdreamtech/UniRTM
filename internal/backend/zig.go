// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

type ZigBackend struct {
	client        *http.Client
	requestGroup  singleflight.Group
	versionsCache sync.Map // map[string][]string (tool -> version strings)
}

func NewZigBackend() *ZigBackend {
	return &ZigBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *ZigBackend) Name() string {
	return "zig"
}

func (b *ZigBackend) Dependencies() []string {
	return nil
}

type zigDownloadResponse map[string]interface{}

func (b *ZigBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
	}

	if val, ok := b.versionsCache.Load("zig_versions"); ok {
		versionStrs := val.([]string)
		versions := make([]VersionInfo, len(versionStrs))
		for i, v := range versionStrs {
			versions[i] = VersionInfo{
				Version:  v,
				Platform: platform,
			}
		}
		return versions, nil
	}

	result, err, _ := b.requestGroup.Do("zig:versions", func() (interface{}, error) {
		if val, ok := b.versionsCache.Load("zig_versions"); ok {
			return val, nil
		}

		// Zig compiler versions are listed at https://ziglang.org/download/index.json
		// For zig packages, it's often github releases.
		// For now we implement the compiler/core discovery.
		url := "https://ziglang.org/download/index.json"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var data zigDownloadResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for v := range data {
			versionStrs = append(versionStrs, v)
		}

		b.versionsCache.Store("zig_versions", versionStrs)
		return versionStrs, nil
	})

	if err != nil {
		return nil, err
	}

	versionStrs := result.([]string)
	versions := make([]VersionInfo, len(versionStrs))
	for i, v := range versionStrs {
		versions[i] = VersionInfo{
			Version:  v,
			Platform: platform,
		}
	}

	return versions, nil
}

func (b *ZigBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		return &VersionInfo{
			Version:  "master", // Zig calls latest 'master' or we pick from index
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *ZigBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *ZigBackend) SupportsChecksum() bool {
	return true
}

func (b *ZigBackend) SupportsGPG() bool {
	return false
}

func (b *ZigBackend) AttestationType() string {
	return ""
}

func (b *ZigBackend) IsRecommended() bool {
	return true
}

func (b *ZigBackend) IsScriptless() bool {
	return true
}

func (b *ZigBackend) GetReach() string {
	return "Medium"
}

func (b *ZigBackend) IsStable() bool {
	return true
}

func (b *ZigBackend) SupportsOffline() bool {
	return true
}
