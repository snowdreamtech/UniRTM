// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// CargoBackend implements the Backend interface for Cargo packages.
type CargoBackend struct {
	client *http.Client
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
		Num string `json:"num"`
	} `json:"versions"`
}

func (b *CargoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var versions []VersionInfo
	for _, v := range registry.Versions {
		versions = append(versions, VersionInfo{
			Version:  v.Num,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *CargoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", tool)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "unirtm (https://github.com/snowdreamtech/unirtm)")

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, "crate not found", nil)
		}

		var registry cargoRegistryResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
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

func (b *CargoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
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
