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
	// Tool format: "owner/repo"
	if !strings.Contains(tool, "/") {
		return nil, NewBackendError("github", tool, "invalid tool format, expected 'owner/repo'", nil)
	}

	releases, err := g.fetchReleases(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, release := range releases {
		// Skip drafts
		if release.Draft {
			continue
		}

		// Find matching asset for platform
		asset, checksum := g.findMatchingAsset(release.Assets, platform)
		if asset == nil {
			continue
		}

		version := strings.TrimPrefix(release.TagName, "v")
		versions = append(versions, VersionInfo{
			Version:     version,
			DownloadURL: asset.BrowserDownloadURL,
			Checksum:    checksum,
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
	return g.ResolveVersion(ctx, tool, version, platform)
}

// SupportsChecksum indicates whether this backend provides checksums.
func (g *GitHubBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (g *GitHubBackend) SupportsGPG() bool {
	return false
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

// findMatchingAsset finds the asset that matches the platform and its checksum.
func (g *GitHubBackend) findMatchingAsset(assets []githubAsset, platform Platform) (*githubAsset, string) {
	var checksumAsset *githubAsset
	var checksumMap map[string]string

	// First pass: find checksum file
	for i := range assets {
		asset := &assets[i]
		if strings.HasSuffix(asset.Name, ".sha256") ||
			strings.HasSuffix(asset.Name, ".sha256sum") ||
			strings.Contains(asset.Name, "checksums") ||
			strings.Contains(asset.Name, "SHA256SUMS") {
			checksumAsset = asset
			break
		}
	}

	// If checksum file found, parse it
	if checksumAsset != nil {
		checksumMap = g.parseChecksumFile(checksumAsset.BrowserDownloadURL)
	}

	// Second pass: find matching binary asset
	for i := range assets {
		asset := &assets[i]

		// Skip checksum files
		if strings.HasSuffix(asset.Name, ".sha256") ||
			strings.HasSuffix(asset.Name, ".sha256sum") ||
			strings.Contains(asset.Name, "checksums") ||
			strings.Contains(asset.Name, "SHA256SUMS") {
			continue
		}

		// Check if asset matches platform
		if g.matchesPlatform(asset.Name, platform) {
			checksum := ""
			if checksumMap != nil {
				checksum = checksumMap[asset.Name]
			}
			return asset, checksum
		}
	}

	return nil, ""
}

// matchesPlatform checks if an asset name matches the platform.
func (g *GitHubBackend) matchesPlatform(assetName string, platform Platform) bool {
	assetLower := strings.ToLower(assetName)

	// Check OS match
	osMatch := false
	switch platform.OS {
	case "linux":
		osMatch = strings.Contains(assetLower, "linux")
	case "darwin":
		osMatch = strings.Contains(assetLower, "darwin") || strings.Contains(assetLower, "macos") || strings.Contains(assetLower, "osx")
	case "windows":
		osMatch = strings.Contains(assetLower, "windows") || strings.Contains(assetLower, "win")
	}

	if !osMatch {
		return false
	}

	// Check architecture match
	archMatch := false
	switch platform.Arch {
	case "amd64":
		archMatch = strings.Contains(assetLower, "amd64") || strings.Contains(assetLower, "x86_64") || strings.Contains(assetLower, "x64")
	case "arm64":
		archMatch = strings.Contains(assetLower, "arm64") || strings.Contains(assetLower, "aarch64")
	case "386":
		archMatch = strings.Contains(assetLower, "386") || strings.Contains(assetLower, "i386") || strings.Contains(assetLower, "x86")
	}

	return archMatch
}

// parseChecksumFile downloads and parses a checksum file.
func (g *GitHubBackend) parseChecksumFile(url string) map[string]string {
	resp, err := g.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	checksums := make(map[string]string)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "checksum  filename" or "checksum filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[1]
			checksums[filename] = checksum
		}
	}

	return checksums
}
