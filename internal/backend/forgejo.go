// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// ForgejoBackend implements the Backend interface for Forgejo/Gitea releases.
type ForgejoBackend struct {
	client  *http.Client
	baseURL string
}

// NewForgejoBackend creates a new Forgejo backend.
func NewForgejoBackend() *ForgejoBackend {
	baseURL := os.Getenv("FORGEJO_API_URL")
	if baseURL == "" {
		// Default to some known Forgejo instance or just empty if custom is required
		baseURL = "https://codeberg.org/api/v1"
	}
	return &ForgejoBackend{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Name returns the backend identifier.
func (b *ForgejoBackend) Name() string {
	return "forgejo"
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

// ListVersions returns all available versions from Forgejo releases.
func (b *ForgejoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")
	apiURL := fmt.Sprintf("%s/repos/%s/releases", b.baseURL, tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	if token := os.Getenv("FORGEJO_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var releases []forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, release := range releases {
		commonAssets := b.toCommonAssets(release.Assets)
		bestAsset, _ := FindBestAsset(commonAssets, platform, tool)
		if bestAsset == nil {
			continue
		}

		v := strings.TrimPrefix(release.TagName, "v")
		versions = append(versions, VersionInfo{
			Version:     v,
			DownloadURL: bestAsset.URL,
			Platform:    platform,
		})
	}

	return versions, nil
}

// ResolveVersion resolves a version request to a concrete version.
func (b *ForgejoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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
		Version:  strings.TrimPrefix(versionRequest, "v"),
		Platform: platform,
	}, nil
}

// GetDownloadInfo retrieves download information for a specific version.
func (b *ForgejoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")
	
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}

	apiURL := fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if token := os.Getenv("FORGEJO_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// try without 'v'
		tag = version
		apiURL = fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if token := os.Getenv("FORGEJO_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		resp, err = b.client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, "release not found", nil)
		}
		defer resp.Body.Close()
	}

	var release forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	commonAssets := b.toCommonAssets(release.Assets)
	bestAsset, _ := FindBestAsset(commonAssets, platform, tool)
	if bestAsset == nil {
		return nil, NewBackendError(b.Name(), tool, "no matching asset found", nil)
	}

	checksum := FindChecksumForAsset(ctx, b.client, commonAssets, bestAsset)

	return &VersionInfo{
		Version:     version,
		DownloadURL: bestAsset.URL,
		Checksum:    checksum,
		Platform:    platform,
	}, nil
}

func (b *ForgejoBackend) toCommonAssets(assets []forgejoAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.URL, Size: a.Size}
	}
	return res
}

// SupportsChecksum indicates whether this backend provides checksums.
func (b *ForgejoBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (b *ForgejoBackend) SupportsGPG() bool {
	return false
}

// AttestationType returns the type of attestation verification supported.
func (b *ForgejoBackend) AttestationType() string {
	return "SLSA"
}
