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
		baseURL = "https://codeberg.org/api/v1" // Defaulting to Codeberg as the most popular Forgejo instance
	}
	return &ForgejoBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (b *ForgejoBackend) Name() string {
	return "forgejo"
}

type forgejoRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (b *ForgejoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// tool format: "owner/repo"
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewBackendError(b.Name(), tool, "repository not found", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var releases []forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, release := range releases {
		v := strings.TrimPrefix(release.TagName, "v")
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

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
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *ForgejoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// fetch the specific release by tag name
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}
	apiURL := fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)

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
		// Try without 'v'
		tag = version
		apiURL = fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if token := os.Getenv("FORGEJO_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		resp2, err2 := b.client.Do(req)
		if err2 != nil || resp2.StatusCode != http.StatusOK {
			if resp2 != nil {
				resp2.Body.Close()
			}
			return nil, NewBackendError(b.Name(), tool, "release not found", nil)
		}
		resp = resp2
		defer resp.Body.Close()
	}

	var release forgejoRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var downloadURL string
	for _, asset := range release.Assets {
		lowerName := strings.ToLower(asset.Name)
		if strings.Contains(lowerName, platform.OS) && strings.Contains(lowerName, platform.Arch) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" && len(release.Assets) > 0 {
		downloadURL = release.Assets[0].BrowserDownloadURL
	}

	return &VersionInfo{
		Version:     version,
		DownloadURL: downloadURL,
		Platform:    platform,
	}, nil
}

func (b *ForgejoBackend) SupportsChecksum() bool {
	return false
}

func (b *ForgejoBackend) SupportsGPG() bool {
	return false
}
