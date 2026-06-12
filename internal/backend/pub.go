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

type PubBackend struct {
	client *http.Client
}

func NewPubBackend() *PubBackend {
	return &PubBackend{
		client: pkgHttp.NewClientWithTimeout(15 * time.Second),
	}
}

func (b *PubBackend) Name() string {
	return "pub"
}

func (b *PubBackend) Dependencies() []string {
	return []string{"dart"}
}

type pubResponse struct {
	Name     string `json:"name"`
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

func (b *PubBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://pub.dev/api/packages/%s", tool)

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
		return nil, NewBackendError(b.Name(), tool, "package not found on pub.dev", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var data pubResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, v := range data.Versions {
		versions = append(versions, VersionInfo{
			Version:  v.Version,
			Platform: platform,
		})
	}

	// Reverse to get descending order
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return versions, nil
}

func (b *PubBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
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

func (b *PubBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *PubBackend) SupportsChecksum() bool  { return false }
func (b *PubBackend) SupportsGPG() bool       { return false }
func (b *PubBackend) AttestationType() string { return "" }
func (b *PubBackend) IsRecommended() bool     { return true }
func (b *PubBackend) IsScriptless() bool      { return true }
func (b *PubBackend) GetReach() string        { return "Medium" }
func (b *PubBackend) IsStable() bool          { return true }
func (b *PubBackend) SupportsOffline() bool   { return true }
