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
			Timeout: 15 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Name returns the backend identifier.
func (b *GitlabBackend) Name() string {
	return "gitlab"
}

type gitlabRelease struct {
	TagName string `json:"tag_name"`
	Assets  struct {
		Links []gitlabAsset `json:"links"`
	} `json:"assets"`
}

type gitlabAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ListVersions returns all available versions from GitLab releases.
func (b *GitlabBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// tool format: "owner/repo" or "gitlab:owner/repo"
	tool = strings.TrimPrefix(tool, "gitlab:")
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
		commonAssets := b.toCommonAssets(release.Assets.Links)
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
		Version:  strings.TrimPrefix(versionRequest, "v"),
		Platform: platform,
	}, nil
}

// GetDownloadInfo retrieves download information for a specific version.
func (b *GitlabBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// tool format: "owner/repo" or "gitlab:owner/repo"
	tool = strings.TrimPrefix(tool, "gitlab:")
	encodedRepo := url.PathEscape(tool)
	
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}
	
	release, err := b.fetchRelease(ctx, encodedRepo, tag)
	if err != nil {
		// try without 'v'
		tag = version
		release, err = b.fetchRelease(ctx, encodedRepo, tag)
		if err != nil {
			return nil, err
		}
	}

	commonAssets := b.toCommonAssets(release.Assets.Links)
	bestAsset, _ := FindBestAsset(commonAssets, platform, tool)
	if bestAsset == nil {
		return nil, NewBackendError(b.Name(), tool, "no matching asset found", nil)
	}

	checksum := FindChecksumForAsset(ctx, b.client, commonAssets, bestAsset)
	gpgSigURL := FindGPGSignatureForAsset(commonAssets, bestAsset)

	return &VersionInfo{
		Version:     version,
		DownloadURL: bestAsset.URL,
		Checksum:    checksum,
		Platform:    platform,
		Metadata: map[string]string{
			"gpg_signature_url": gpgSigURL,
		},
	}, nil
}

func (b *GitlabBackend) fetchRelease(ctx context.Context, encodedRepo, tag string) (*gitlabRelease, error) {
	apiURL := fmt.Sprintf("%s/projects/%s/releases/%s", b.baseURL, encodedRepo, url.PathEscape(tag))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var release gitlabRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func (b *GitlabBackend) toCommonAssets(assets []gitlabAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.URL}
	}
	return res
}

// SupportsChecksum indicates whether this backend provides checksums.
func (b *GitlabBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (b *GitlabBackend) SupportsGPG() bool {
	return true
}

// AttestationType returns the type of attestation verification supported.
func (b *GitlabBackend) AttestationType() string {
	return "SLSA"
}
