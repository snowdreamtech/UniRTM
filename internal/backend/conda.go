// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// CondaBackend implements the Backend interface for Conda packages.
type CondaBackend struct {
	client *http.Client
}

// NewCondaBackend creates a new conda backend.
func NewCondaBackend() *CondaBackend {
	return &CondaBackend{
		client: pkgHttp.NewClientWithTimeout(15 * time.Second),
	}
}

func (b *CondaBackend) Name() string {
	return "conda"
}

func (b *CondaBackend) Dependencies() []string {
	return nil
}

type anacondaResponse struct {
	Versions []string `json:"versions"`
}

func (b *CondaBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// We use the Anaconda API to fetch package metadata
	url := fmt.Sprintf("https://api.anaconda.org/package/anaconda/%s", tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewBackendError(b.Name(), tool, "package not found in anaconda registry", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var data anacondaResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, v := range data.Versions {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	// Sort newest first
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	return versions, nil
}

func (b *CondaBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		return &versions[0], nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CondaBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CondaBackend) SupportsChecksum() bool {
	return true
}

func (b *CondaBackend) SupportsGPG() bool {
	return false
}

func (b *CondaBackend) AttestationType() string {
	return ""
}

func (b *CondaBackend) IsRecommended() bool {
	return true
}

func (b *CondaBackend) IsScriptless() bool {
	return true
}

func (b *CondaBackend) GetReach() string {
	return "Large"
}

func (b *CondaBackend) IsStable() bool {
	return true
}

func (b *CondaBackend) SupportsOffline() bool {
	return true
}
