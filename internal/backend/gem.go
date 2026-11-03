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
	gemFlight        singleflight.Group
	gemVersionsCache sync.Map // map[string][]string (tool -> version numbers)
	gemLatestCache   sync.Map // map[string]string (tool -> latest version)
)

// ClearGemCache clears the in-memory cache for RubyGems. Mainly used for testing.
func ClearGemCache() {
	gemVersionsCache.Range(func(key, _ interface{}) bool {
		gemVersionsCache.Delete(key)
		return true
	})
	gemLatestCache.Range(func(key, _ interface{}) bool {
		gemLatestCache.Delete(key)
		return true
	})
}

// GemBackend implements the Backend interface for RubyGems.
type GemBackend struct {
	client        *http.Client
	requestGroup  *singleflight.Group
	versionsCache *sync.Map
	latestCache   *sync.Map
}

// NewGemBackend creates a new gem backend.
func NewGemBackend() *GemBackend {
	return &GemBackend{
		client:        pkgHttp.NewClientWithTimeout(10 * time.Second),
		requestGroup:  &gemFlight,
		versionsCache: &gemVersionsCache,
		latestCache:   &gemLatestCache,
	}
}

func (b *GemBackend) Name() string {
	return "gem"
}

func (b *GemBackend) Dependencies() []string {
	return []string{"ruby"}
}

type gemVersion struct {
	Number string `json:"number"`
}

func (b *GemBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
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

	if b.client.Transport == nil {
		var versionStrs []string
		if readEcosystemMetadataDiskCache("gem", tool, &versionStrs, 10*time.Minute) {
			b.versionsCache.Store(tool, versionStrs)
			versions := make([]VersionInfo, len(versionStrs))
			for i, v := range versionStrs {
				versions[i] = VersionInfo{
					Version:  v,
					Platform: platform,
				}
			}
			return versions, nil
		}
	}

	result, err, _ := b.requestGroup.Do("versions:"+tool, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(tool); ok {
			return val, nil
		}

		url := fmt.Sprintf("https://rubygems.org/api/v1/versions/%s.json", tool)

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
			return nil, NewBackendError(b.Name(), tool, "gem not found", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var gemVersions []gemVersion
		if err := json.NewDecoder(resp.Body).Decode(&gemVersions); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for _, v := range gemVersions {
			versionStrs = append(versionStrs, v.Number)
		}

		b.versionsCache.Store(tool, versionStrs)
		if b.client.Transport == nil {
			writeEcosystemMetadataDiskCache("gem", tool, versionStrs)
		}
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

func (b *GemBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		if val, ok := b.latestCache.Load(tool); ok {
			return &VersionInfo{
				Version:  val.(string),
				Platform: platform,
			}, nil
		}

		result, err, _ := b.requestGroup.Do("latest:"+tool, func() (interface{}, error) {
			if val, ok := b.latestCache.Load(tool); ok {
				return val, nil
			}

			url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", tool)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
			if err != nil {
				return nil, err
			}

			resp, err := b.client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, NewBackendError(b.Name(), tool, "latest version not found", nil)
			}

			var latest struct {
				Version string `json:"version"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
				return nil, err
			}

			b.latestCache.Store(tool, latest.Version)
			return latest.Version, nil
		})

		if err != nil {
			return nil, err
		}

		return &VersionInfo{
			Version:  result.(string),
			Platform: platform,
		}, nil
	}

	// Assume explicit version is correct; provider will fail if it doesn't exist
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *GemBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	// Provider handles actual downloading via gem cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *GemBackend) SupportsChecksum() bool {
	return true
}

func (b *GemBackend) SupportsGPG() bool {
	return false
}

func (b *GemBackend) AttestationType() string {
	return ""
}

func (b *GemBackend) IsRecommended() bool {
	return true
}

func (b *GemBackend) IsScriptless() bool {
	return true
}

func (b *GemBackend) GetReach() string {
	return "Large"
}

func (b *GemBackend) IsStable() bool {
	return true
}

func (b *GemBackend) SupportsOffline() bool {
	return true
}
