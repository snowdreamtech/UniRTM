// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// GitHubBackend implements the Backend interface for GitHub Releases.
type GitHubBackend struct {
	client *http.Client
}

// NewGitHubBackend creates a new GitHub backend.
func NewGitHubBackend() *GitHubBackend {
	return &GitHubBackend{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the backend identifier.
func (g *GitHubBackend) Name() string {
	return "github"
}

// githubRelease represents a GitHub release from the API.
type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
	CreatedAt  string        `json:"created_at"`
}

// githubAsset represents a release asset from the GitHub API.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// ListVersions returns all available versions from GitHub Releases.
func (g *GitHubBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Tool format: "owner/repo" or "github:owner/repo"
	tool = strings.TrimPrefix(tool, "github:")
	if !strings.Contains(tool, "/") {
		return nil, NewBackendError("github", tool, "invalid tool format, expected 'owner/repo'", nil)
	}

	releases, err := g.fetchReleases(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, release := range releases {
		if release.Draft {
			continue
		}

		// Find matching asset for platform but DON'T fetch checksum here
		asset := g.findMatchingAssetOnly(release.Assets, platform, tool)
		if asset == nil {
			continue
		}

		version := strings.TrimPrefix(release.TagName, "v")
		versions = append(versions, VersionInfo{
			Version:     version,
			DownloadURL: asset.BrowserDownloadURL,
			Platform:    platform,
			Metadata: map[string]string{
				"tag_name":   release.TagName,
				"name":       release.Name,
				"prerelease": fmt.Sprintf("%t", release.Prerelease),
				"created_at": release.CreatedAt,
				"asset_name": asset.Name,
				"asset_size": fmt.Sprintf("%d", asset.Size),
			},
		})
	}

	if len(versions) == 0 {
		return nil, NewBackendError("github", tool, fmt.Sprintf("no releases found for platform %s", platform.String()), nil)
	}

	return versions, nil
}

// ResolveVersion resolves a version request to a concrete version.
func (g *GitHubBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	versions, err := g.ListVersions(ctx, tool, platform)
	if err != nil {
		return nil, err
	}

	// Handle special version requests
	switch versionRequest {
	case "latest":
		// Return the first non-prerelease version
		for _, v := range versions {
			if v.Metadata["prerelease"] == "false" {
				return &v, nil
			}
		}
		// If no stable version, return the first version
		if len(versions) > 0 {
			return &versions[0], nil
		}
		return nil, NewBackendError("github", tool, "no versions available", nil)

	case "stable":
		// Return the first non-prerelease version
		for _, v := range versions {
			if v.Metadata["prerelease"] == "false" {
				return &v, nil
			}
		}
		return nil, NewBackendError("github", tool, "no stable versions available", nil)

	default:
		// Exact version match
		versionRequest = strings.TrimPrefix(versionRequest, "v")
		for _, v := range versions {
			if v.Version == versionRequest {
				return &v, nil
			}
		}
		return nil, NewBackendError("github", tool, fmt.Sprintf("version %s not found", versionRequest), nil)
	}
}

// GetDownloadInfo retrieves download information for a specific version.
func (g *GitHubBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	v, err := g.ResolveVersion(ctx, tool, version, platform)
	if err != nil {
		return nil, err
	}

	// Now that we have a concrete version, try to fetch its checksum if not already present
	if v.Checksum == "" {
		releases, err := g.fetchReleases(ctx, strings.TrimPrefix(tool, "github:"))
		if err == nil {
			for _, r := range releases {
				if strings.TrimPrefix(r.TagName, "v") == v.Version {
					commonAssets := g.toCommonAssets(r.Assets)
					bestAsset, _ := FindBestAsset(commonAssets, platform, tool)
					if bestAsset != nil {
						v.Checksum = FindChecksumForAsset(ctx, g.client, commonAssets, bestAsset)
						if v.Metadata == nil {
							v.Metadata = make(map[string]string)
						}
						v.Metadata["gpg_signature_url"] = FindGPGSignatureForAsset(commonAssets, bestAsset)
					}
					break
				}
			}
		}
	}

	return v, nil
}

func (g *GitHubBackend) toCommonAssets(assets []githubAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.BrowserDownloadURL, Size: a.Size}
	}
	return res
}

// findMatchingAssetOnly uses the common scoring system to find the best asset.
func (g *GitHubBackend) findMatchingAssetOnly(assets []githubAsset, platform Platform, tool string) *githubAsset {
	commonAssets := g.toCommonAssets(assets)
	bestAsset, _ := FindBestAsset(commonAssets, platform, tool)
	if bestAsset == nil {
		return nil
	}

	// Map back to githubAsset
	for i := range assets {
		if assets[i].BrowserDownloadURL == bestAsset.URL {
			return &assets[i]
		}
	}
	return nil
}

// SupportsChecksum indicates whether this backend provides checksums.
func (g *GitHubBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (g *GitHubBackend) SupportsGPG() bool {
	return true
}

// AttestationType returns the type of attestation verification supported.
func (g *GitHubBackend) AttestationType() string {
	return "GitHub Attestation"
}

// fetchReleases fetches all releases from GitHub API.
func (g *GitHubBackend) fetchReleases(ctx context.Context, tool string) ([]githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", tool)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, NewBackendError("github", tool, "failed to create request", err)
	}

	// Add GitHub API headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Inject token to avoid rate limiting (403) in CI and for private repos.
	if token := resolveGitHubToken("github.com"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, NewBackendError("github", tool, "failed to fetch releases", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, NewBackendError("github", tool, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)), nil)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, NewBackendError("github", tool, "failed to decode releases", err)
	}

	// Sort releases by tag name (descending)
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].TagName > releases[j].TagName
	})

	return releases, nil
}
