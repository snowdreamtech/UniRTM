// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

var (
	npmCache        sync.Map
	npmFlight       singleflight.Group
	npmLatestCache  sync.Map
	npmLatestFlight singleflight.Group
)

// ClearNpmMetadataCache clears in-memory caches for NPM package metadata.
func ClearNpmMetadataCache() {
	npmCache.Range(func(key, value any) bool {
		npmCache.Delete(key)
		return true
	})
	npmLatestCache.Range(func(key, value any) bool {
		npmLatestCache.Delete(key)
		return true
	})
}

// NpmBackend implements the Backend interface for npm packages.
type NpmBackend struct {
	client       *http.Client
	flight       singleflight.Group
	latestFlight singleflight.Group
	cache        sync.Map
	latestCache  sync.Map
}

// NewNpmBackend creates a new npm backend.
func NewNpmBackend() *NpmBackend {
	return &NpmBackend{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}
}

func (b *NpmBackend) Name() string {
	return "npm"
}

func (b *NpmBackend) Dependencies() []string {
	return []string{"node"}
}

type npmRegistryResponse struct {
	DistTags map[string]string      `json:"dist-tags"`
	Versions map[string]interface{} `json:"versions"`
	Time     map[string]string      `json:"time"`
}

func (b *NpmBackend) fetchRegistry(ctx context.Context, tool string) (*npmRegistryResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if val, ok := npmCache.Load(tool); ok {
		if reg, ok := val.(*npmRegistryResponse); ok {
			return reg, nil
		}
	}

	val, err, _ := npmFlight.Do(tool, func() (interface{}, error) {

		baseURL := env.Get("NPM_REGISTRY_URL")
		if baseURL == "" {
			baseURL = env.Get("NPM_CONFIG_REGISTRY")
		}
		if baseURL == "" {
			baseURL = "https://registry.npmjs.org"
		}
		// NPM_CONFIG_REGISTRY might have a trailing slash
		baseURL = strings.TrimRight(baseURL, "/")
		url := fmt.Sprintf("%s/%s", baseURL, tool)

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

		var registry npmRegistryResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		npmCache.Store(tool, &registry)
		return &registry, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*npmRegistryResponse), nil
}

func (b *NpmBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	registry, err := b.fetchRegistry(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for v := range registry.Versions {
		var publishedAt time.Time
		if timeStr, ok := registry.Time[v]; ok && timeStr != "" {
			if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
				publishedAt = t
			}
		}
		versions = append(versions, VersionInfo{
			Version:     v,
			Platform:    platform,
			PublishedAt: publishedAt,
		})
	}

	// npm registry doesn't strictly order keys in maps, so we should ideally sort them.
	// But since this is just a listing and version manager usually sorts them semantically,
	// returning them as-is is acceptable for now.

	return versions, nil
}

func (b *NpmBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		if val, ok := npmLatestCache.Load(tool); ok {
			if ver, ok := val.(string); ok && ver != "" {
				return &VersionInfo{
					Version:  ver,
					Platform: platform,
				}, nil
			}
		}

		// Check if we already have the full registry in cache with dist-tags
		if val, ok := npmCache.Load(tool); ok {
			if reg, ok := val.(*npmRegistryResponse); ok && reg.DistTags != nil {
				if latest, ok := reg.DistTags["latest"]; ok && latest != "" {
					npmLatestCache.Store(tool, latest)
					return &VersionInfo{
						Version:  latest,
						Platform: platform,
					}, nil
				}
			}
		}

		val, err, _ := npmLatestFlight.Do(tool, func() (interface{}, error) {
			baseURL := env.Get("NPM_REGISTRY_URL")
			if baseURL == "" {
				baseURL = env.Get("NPM_CONFIG_REGISTRY")
			}
			if baseURL == "" {
				baseURL = "https://registry.npmjs.org"
			}
			baseURL = strings.TrimRight(baseURL, "/")
			url := fmt.Sprintf("%s/%s/latest", baseURL, tool)
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

			npmLatestCache.Store(tool, latest.Version)
			return latest.Version, nil
		})


		if err != nil {
			return nil, err
		}

		return &VersionInfo{
			Version:  val.(string),
			Platform: platform,
		}, nil
	}

	// Assume explicit version is correct; provider will fail if it doesn't exist
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *NpmBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	// Provider handles actual downloading via npm cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *NpmBackend) SupportsChecksum() bool {
	return true
}

func (b *NpmBackend) SupportsGPG() bool {
	return false
}

func (b *NpmBackend) AttestationType() string {
	return ""
}

func (b *NpmBackend) IsRecommended() bool {
	return true
}

func (b *NpmBackend) IsScriptless() bool {
	return true
}

func (b *NpmBackend) GetReach() string {
	return "Huge"
}

func (b *NpmBackend) IsStable() bool {
	return true
}

func (b *NpmBackend) SupportsOffline() bool {
	return true
}
