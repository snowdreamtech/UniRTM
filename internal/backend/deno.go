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

var (
	denoFlight        singleflight.Group
	denoVersionsCache sync.Map // map[string]*denoVersionsResponse
)

// ClearDenoCache clears both in-memory and disk cache for Deno modules. Mainly used for testing.
func ClearDenoCache() {
	denoVersionsCache.Range(func(key, _ interface{}) bool {
		denoVersionsCache.Delete(key)
		return true
	})
	ClearEcosystemMetadataDiskCache("deno")
}

type DenoBackend struct {
	client        *http.Client
	requestGroup  *singleflight.Group
	versionsCache *sync.Map
}

func NewDenoBackend() *DenoBackend {
	return &DenoBackend{
		client:        pkgHttp.NewClientWithTimeout(10 * time.Second),
		requestGroup:  &denoFlight,
		versionsCache: &denoVersionsCache,
	}
}

func (b *DenoBackend) Name() string {
	return "deno"
}

func (b *DenoBackend) Dependencies() []string {
	return nil
}

type denoVersionsResponse struct {
	Latest   string   `json:"latest"`
	Versions []string `json:"versions"`
}

func (b *DenoBackend) fetchMeta(ctx context.Context, tool string) (*denoVersionsResponse, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
	}

	if val, ok := b.versionsCache.Load(tool); ok {
		return val.(*denoVersionsResponse), nil
	}

	res, err, _ := b.requestGroup.Do("deno:"+tool, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(tool); ok {
			return val.(*denoVersionsResponse), nil
		}

		var diskData denoVersionsResponse
		if readEcosystemMetadataDiskCache("deno", tool, &diskData, 10*time.Minute) {
			b.versionsCache.Store(tool, &diskData)
			return &diskData, nil
		}

		url := fmt.Sprintf("https://cdn.deno.land/%s/meta/versions.json", tool)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, NewBackendError(b.Name(), tool, "module not found on deno.land", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var data denoVersionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		writeEcosystemMetadataDiskCache("deno", tool, data)
		b.versionsCache.Store(tool, &data)
		return &data, nil
	})

	if err != nil {
		return nil, err
	}
	return res.(*denoVersionsResponse), nil
}

func (b *DenoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	data, err := b.fetchMeta(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, v := range data.Versions {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *DenoBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		data, err := b.fetchMeta(ctx, tool)
		if err != nil {
			return nil, err
		}
		return &VersionInfo{
			Version:  data.Latest,
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *DenoBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *DenoBackend) SupportsChecksum() bool {
	return true
}

func (b *DenoBackend) SupportsGPG() bool {
	return false
}

func (b *DenoBackend) AttestationType() string {
	return ""
}

func (b *DenoBackend) IsRecommended() bool {
	return true
}

func (b *DenoBackend) IsScriptless() bool {
	return true
}

func (b *DenoBackend) GetReach() string {
	return "Medium"
}

func (b *DenoBackend) IsStable() bool {
	return true
}

func (b *DenoBackend) SupportsOffline() bool {
	return true
}
