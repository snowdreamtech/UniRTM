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
	cabalFlight        singleflight.Group
	cabalVersionsCache sync.Map // map[string][]string (tool -> version strings)
)

// ClearCabalCache clears the in-memory cache for Cabal packages. Mainly used for testing.
func ClearCabalCache() {
	cabalVersionsCache.Range(func(key, _ interface{}) bool {
		cabalVersionsCache.Delete(key)
		return true
	})
}

type CabalBackend struct {
	client        *http.Client
	requestGroup  *singleflight.Group
	versionsCache *sync.Map
}

func NewCabalBackend() *CabalBackend {
	return &CabalBackend{
		client:        pkgHttp.NewClientWithTimeout(10 * time.Second),
		requestGroup:  &cabalFlight,
		versionsCache: &cabalVersionsCache,
	}
}

func (b *CabalBackend) Name() string {
	return "cabal"
}

func (b *CabalBackend) Dependencies() []string {
	return []string{"haskell"}
}

type hackageResponse []struct {
	Version string `json:"version"`
}

func (b *CabalBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
	}

	if val, ok := b.versionsCache.Load(tool); ok {
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

	result, err, _ := b.requestGroup.Do("cabal:"+tool, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(tool); ok {
			return val, nil
		}

		url := fmt.Sprintf("https://hackage.haskell.org/package/%s.json", tool)

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
			return nil, NewBackendError(b.Name(), tool, "package not found on hackage", nil)
		}

		var data hackageResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for _, v := range data {
			versionStrs = append(versionStrs, v.Version)
		}

		b.versionsCache.Store(tool, versionStrs)
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

func (b *CabalBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		return &versions[len(versions)-1], nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CabalBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CabalBackend) SupportsChecksum() bool {
	return true
}

func (b *CabalBackend) SupportsGPG() bool {
	return false
}

func (b *CabalBackend) AttestationType() string {
	return ""
}

func (b *CabalBackend) IsRecommended() bool {
	return true
}

func (b *CabalBackend) IsScriptless() bool {
	return true
}

func (b *CabalBackend) GetReach() string {
	return "Large"
}

func (b *CabalBackend) IsStable() bool {
	return true
}

func (b *CabalBackend) SupportsOffline() bool {
	return true
}
