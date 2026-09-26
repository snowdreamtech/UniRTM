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

	"golang.org/x/sync/singleflight"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

var (
	pypiCache  sync.Map
	pypiFlight singleflight.Group
)

// ClearPypiMetadataCache clears in-memory caches for PyPI package metadata.
func ClearPypiMetadataCache() {
	pypiCache.Range(func(key, value any) bool {
		pypiCache.Delete(key)
		return true
	})
}

// PypiBackend implements the Backend interface for PyPI packages.
type PypiBackend struct {
	client *http.Client
	flight singleflight.Group
	cache  sync.Map
}

// NewPypiBackend creates a new PyPI backend.
func NewPypiBackend() *PypiBackend {
	return &PypiBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *PypiBackend) Name() string {
	return "pypi"
}

func (b *PypiBackend) Dependencies() []string {
	return nil
}

type pypiRegistryResponse struct {
	Releases map[string]interface{} `json:"releases"`
	Info     struct {
		Version string `json:"version"`
	} `json:"info"`
}

func (b *PypiBackend) fetchRegistry(ctx context.Context, tool string) (*pypiRegistryResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if val, ok := pypiCache.Load(tool); ok {
		if reg, ok := val.(*pypiRegistryResponse); ok {
			return reg, nil
		}
	}

	if b.client.Transport == nil {
		var reg pypiRegistryResponse
		if readEcosystemMetadataDiskCache("pypi", tool, &reg, 10*time.Minute) {
			pypiCache.Store(tool, &reg)
			return &reg, nil
		}
	}

	val, err, _ := pypiFlight.Do(tool, func() (interface{}, error) {
		if val, ok := pypiCache.Load(tool); ok {
			if reg, ok := val.(*pypiRegistryResponse); ok {
				return reg, nil
			}
		}

		url := fmt.Sprintf("https://pypi.org/pypi/%s/json", tool)

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
			return nil, NewBackendError(b.Name(), tool, "package not found", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var registry pypiRegistryResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		pypiCache.Store(tool, &registry)
		if b.client.Transport == nil {
			writeEcosystemMetadataDiskCache("pypi", tool, &registry)
		}
		return &registry, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*pypiRegistryResponse), nil
}


func (b *PypiBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	registry, err := b.fetchRegistry(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for v := range registry.Releases {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *PypiBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		registry, err := b.fetchRegistry(ctx, tool)
		if err != nil {
			return nil, err
		}

		return &VersionInfo{
			Version:  registry.Info.Version,
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *PypiBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *PypiBackend) SupportsChecksum() bool {
	return true
}

func (b *PypiBackend) SupportsGPG() bool {
	return false
}

func (b *PypiBackend) AttestationType() string {
	return ""
}

func (b *PypiBackend) IsRecommended() bool {
	return true
}

func (b *PypiBackend) IsScriptless() bool {
	return true
}

func (b *PypiBackend) GetReach() string {
	return "Huge"
}

func (b *PypiBackend) IsStable() bool {
	return true
}

func (b *PypiBackend) SupportsOffline() bool {
	return true
}
