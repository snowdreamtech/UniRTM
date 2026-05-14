// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"strings"
	"time"
)

// ForgejoBackend implements the Backend interface using GenericReleaseManager.
type ForgejoBackend struct {
	client  *http.Client
	baseURL string
}

// NewForgejoBackend creates a new Forgejo backend.
func NewForgejoBackend() *ForgejoBackend {
	baseURL := env.Get("FORGEJO_API_URL")
	if baseURL == "" {
		baseURL = "https://codeberg.org/api/v1"
	}
	return &ForgejoBackend{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (b *ForgejoBackend) Name() string            { return "forgejo" }

func (b *ForgejoBackend) Dependencies() []string {
	return nil
}

func (b *ForgejoBackend) GetClient() *http.Client { return b.client }
func (b *ForgejoBackend) GetAttestationType() string {
	return "SLSA"
}
func (b *ForgejoBackend) SupportsChecksum() bool { return true }
func (b *ForgejoBackend) SupportsGPG() bool      { return true }
func (b *ForgejoBackend) AttestationType() string {
	return b.GetAttestationType()
}

func (b *ForgejoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	releases, err := b.FetchReleases(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, release := range releases {
		bestAsset, _ := FindBestAsset(release.Assets, platform, tool)
		if bestAsset == nil {
			continue
		}

		v := strings.TrimPrefix(release.Tag, "v")
		versions = append(versions, VersionInfo{
			Version:     v,
			DownloadURL: bestAsset.URL,
			Platform:    platform,
		})
	}
	return versions, nil
}

func (b *ForgejoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return GenericResolveVersion(ctx, b, tool, versionRequest, platform)
}

func (b *ForgejoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfo(ctx, b, tool, version, platform)
}

// FetchReleases implements HostingProvider.
func (b *ForgejoBackend) FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")
	apiURL := fmt.Sprintf("%s/repos/%s/releases", b.baseURL, tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	if token := env.Get("FORGEJO_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Forgejo API status %d", resp.StatusCode)
	}

	var releases []forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	res := make([]CommonRelease, len(releases))
	for i, r := range releases {
		res[i] = CommonRelease{
			Tag:    r.TagName,
			Assets: b.toCommonAssets(r.Assets),
		}
	}
	return res, nil
}

// FetchReleaseByTag implements HostingProvider.
func (b *ForgejoBackend) FetchReleaseByTag(ctx context.Context, tool string, tag string) (*CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")
	apiURL := fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if token := env.Get("FORGEJO_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var r forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &CommonRelease{
		Tag:    r.TagName,
		Assets: b.toCommonAssets(r.Assets),
	}, nil
}

func (b *ForgejoBackend) toCommonAssets(assets []forgejoAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.URL, Size: a.Size}
	}
	return res
}

type forgejoRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []forgejoAsset `json:"assets"`
}

type forgejoAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

func (b *ForgejoBackend) IsRecommended() bool {
	return true
}

func (b *ForgejoBackend) IsScriptless() bool {
	return true
}

func (b *ForgejoBackend) GetReach() string {
	return "Large"
}

func (b *ForgejoBackend) IsStable() bool {
	return true
}

func (b *ForgejoBackend) SupportsOffline() bool {
	return false
}
