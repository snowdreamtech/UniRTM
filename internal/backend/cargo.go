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
	cargoCache  sync.Map
	cargoFlight singleflight.Group
)

// ClearCargoMetadataCache clears in-memory caches for Cargo crate metadata.
func ClearCargoMetadataCache() {
	cargoCache.Range(func(key, value any) bool {
		cargoCache.Delete(key)
		return true
	})
}

// CargoBackend implements the Backend interface for Cargo packages.
type CargoBackend struct {
	client *http.Client
	flight singleflight.Group
	cache  sync.Map
}

// NewCargoBackend creates a new Cargo backend.
func NewCargoBackend() *CargoBackend {
	return &CargoBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *CargoBackend) Name() string {
	return "cargo"
}

func (b *CargoBackend) Dependencies() []string {
	return []string{"rust"}
}

type cargoRegistryResponse struct {
	Crate struct {
		MaxVersion string `json:"max_version"`
	} `json:"crate"`
	Versions []struct {
		Num       string `json:"num"`
		CreatedAt string `json:"created_at"`
	} `json:"versions"`
}

func (b *CargoBackend) fetchRegistry(ctx context.Context, tool string) (*cargoRegistryResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if val, ok := cargoCache.Load(tool); ok {
		if reg, ok := val.(*cargoRegistryResponse); ok {
			return reg, nil
		}
	}

	val, err, _ := cargoFlight.Do(tool, func() (interface{}, error) {
		url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", tool)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		// crates.io requires a user-agent
		req.Header.Set("User-Agent", "unirtm (https://github.com/snowdreamtech/unirtm)")

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, NewBackendError(b.Name(), tool, "crate not found", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var registry cargoRegistryResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		cargoCache.Store(tool, &registry)
		return &registry, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*cargoRegistryResponse), nil
}


func (b *CargoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	registry, err := b.fetchRegistry(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, v := range registry.Versions {
		var publishedAt time.Time
		if v.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, v.CreatedAt); err == nil {
				publishedAt = t
			}
		}
		versions = append(versions, VersionInfo{
			Version:     v.Num,
			Platform:    platform,
			PublishedAt: publishedAt,
		})
	}

	return versions, nil
}

func (b *CargoBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
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
			Version:  registry.Crate.MaxVersion,
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CargoBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CargoBackend) SupportsChecksum() bool {
	return true
}

func (b *CargoBackend) SupportsGPG() bool {
	return false
}

func (b *CargoBackend) AttestationType() string {
	return ""
}

func (b *CargoBackend) IsRecommended() bool {
	return true
}

func (b *CargoBackend) IsScriptless() bool {
	return true
}

func (b *CargoBackend) GetReach() string {
	return "Huge"
}

func (b *CargoBackend) IsStable() bool {
	return true
}

func (b *CargoBackend) SupportsOffline() bool {
	return true
}
