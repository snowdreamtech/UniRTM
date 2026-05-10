// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GitlabBackend implements the Backend interface for GitLab releases.
type GitlabBackend struct {
	client  *http.Client
	baseURL string
}

// NewGitlabBackend creates a new GitLab backend.
func NewGitlabBackend() *GitlabBackend {
	baseURL := os.Getenv("GITLAB_API_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	return &GitlabBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (b *GitlabBackend) Name() string {
	return "gitlab"
}

type gitlabRelease struct {
	TagName string `json:"tag_name"`
	Assets  struct {
		Links []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"links"`
	} `json:"assets"`
}

func (b *GitlabBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// tool format: "owner/repo"
	encodedRepo := url.PathEscape(tool)
	apiURL := fmt.Sprintf("%s/projects/%s/releases", b.baseURL, encodedRepo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewBackendError(b.Name(), tool, "project not found", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var releases []gitlabRelease
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

func (b *GitlabBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *GitlabBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	encodedRepo := url.PathEscape(tool)
	// try to fetch the specific release by tag name
	// Since tags could be 'v1.0.0' or '1.0.0', we might need to check both, but standard is just trusting versionRequest
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}
	apiURL := fmt.Sprintf("%s/projects/%s/releases/%s", b.baseURL, encodedRepo, url.PathEscape(tag))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// try without 'v'
		tag = version
		apiURL = fmt.Sprintf("%s/projects/%s/releases/%s", b.baseURL, encodedRepo, url.PathEscape(tag))
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			req.Header.Set("PRIVATE-TOKEN", token)
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

	var release gitlabRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	// Simple heuristic to find the right asset URL
	var downloadURL string
	for _, asset := range release.Assets.Links {
		lowerName := strings.ToLower(asset.Name)
		if strings.Contains(lowerName, platform.OS) && strings.Contains(lowerName, platform.Arch) {
			downloadURL = asset.URL
			break
		}
	}

	if downloadURL == "" && len(release.Assets.Links) > 0 {
		// fallback to the first asset if OS/Arch not explicitly found
		downloadURL = release.Assets.Links[0].URL
	}

	return &VersionInfo{
		Version:     version,
		DownloadURL: downloadURL,
		Platform:    platform,
	}, nil
}

func (b *GitlabBackend) SupportsChecksum() bool {
	return false
}

func (b *GitlabBackend) SupportsGPG() bool {
	return false
}
