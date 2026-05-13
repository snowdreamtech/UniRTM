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
					_, checksum := g.findMatchingAsset(r.Assets, platform, tool)
					v.Checksum = checksum
					break
				}
			}
		}
	}

	return v, nil
}

// findMatchingAssetOnly uses a scoring system to find the best asset for the platform.
func (g *GitHubBackend) findMatchingAssetOnly(assets []githubAsset, platform Platform, tool string) *githubAsset {
	var bestAsset *githubAsset
	bestScore := -1

	// Extract repository name from tool (e.g., "sharkdp/fd" -> "fd")
	// We'll use this to prioritize assets containing the tool name.
	repoName := ""
	if parts := strings.Split(tool, "/"); len(parts) == 2 {
		repoName = strings.ToLower(parts[1])
	}

	for i := range assets {
		asset := &assets[i]
		score := g.calculateAssetScore(asset.Name, platform, repoName)

		if score > 0 && score > bestScore {
			bestScore = score
			bestAsset = asset
		}
	}
	return bestAsset
}

// calculateAssetScore calculates a compatibility score for an asset.
// Returns -1 if the asset is definitely incompatible.
func (g *GitHubBackend) calculateAssetScore(assetName string, platform Platform, repoName string) int {
	nameLower := strings.ToLower(assetName)

	// 1. Hard Exclusions (Negative Score)
	excludeSuffixes := []string{".sha256", ".sha256sum", ".md5", ".asc", ".sig", ".sha1", ".deb", ".rpm", ".msi", ".apk", ".dmg", ".pkg", ".txt", ".pdf", ".h", ".c", ".cpp", ".a", ".lib"}
	for _, suffix := range excludeSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return -1
		}
	}
	
	// Exclude non-runtime assets
	negatives := []string{"checksums", "sha256sums", "license", "source", "devel", "dev", "header", "static-lib", "manual", "doc", "man", "debug"}
	for _, neg := range negatives {
		if strings.Contains(nameLower, neg) {
			return -1
		}
	}

	score := 0

	// 2. OS Match
	osMatch := false
	switch platform.OS {
	case "linux":
		if strings.Contains(nameLower, "linux") || strings.Contains(nameLower, "unknown-linux") {
			osMatch = true
			score += 100
		}
	case "darwin":
		if strings.Contains(nameLower, "darwin") || strings.Contains(nameLower, "macos") || strings.Contains(nameLower, "osx") || strings.Contains(nameLower, "apple") {
			osMatch = true
			score += 100
		}
	case "windows":
		if strings.Contains(nameLower, "windows") || strings.Contains(nameLower, "win") || strings.HasSuffix(nameLower, ".exe") {
			osMatch = true
			score += 100
		}
	}

	if !osMatch {
		return -1
	}

	// 3. Architecture Match
	archMatch := false
	switch platform.Arch {
	case "amd64":
		if strings.Contains(nameLower, "amd64") || strings.Contains(nameLower, "x86_64") || strings.Contains(nameLower, "x64") || strings.Contains(nameLower, "64bit") ||
			(platform.OS == "darwin" && strings.Contains(nameLower, "universal")) {
			archMatch = true
			score += 100
		}
	case "arm64":
		if strings.Contains(nameLower, "arm64") || strings.Contains(nameLower, "aarch64") || strings.Contains(nameLower, "armv8") ||
			(platform.OS == "darwin" && strings.Contains(nameLower, "universal")) {
			archMatch = true
			score += 100
		}
	case "386":
		if strings.Contains(nameLower, "386") || strings.Contains(nameLower, "i386") || strings.Contains(nameLower, "x86") || strings.Contains(nameLower, "32bit") {
			archMatch = true
			score += 100
		}
	}

	if !archMatch {
		return -1
	}

	// 4. Preferred Formats
	if strings.HasSuffix(nameLower, ".tar.gz") || strings.HasSuffix(nameLower, ".tgz") {
		score += 50
	} else if strings.HasSuffix(nameLower, ".zip") {
		score += 40
	} else if strings.HasSuffix(nameLower, ".tar.xz") || strings.HasSuffix(nameLower, ".txz") {
		score += 30
	} else if !strings.Contains(nameLower, ".") {
		score += 20 // Raw binary
	}

	// 5. Tool Name Bonus
	if repoName != "" && strings.Contains(nameLower, repoName) {
		score += 50
	}

	// 6. Avoid "musl" if on a glibc system (or vice versa - simple heuristic)
	// For now, let's just prioritize the one without "musl" unless we are specifically looking for it.
	if strings.Contains(nameLower, "musl") {
		score -= 10
	}

	return score
}

// SupportsChecksum indicates whether this backend provides checksums.
func (g *GitHubBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (g *GitHubBackend) SupportsGPG() bool {
	return false
}

// SupportsAttestation indicates whether this backend supports GitHub Attestation.
func (g *GitHubBackend) SupportsAttestation() bool {
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

	// Inject token to avoid rate limiting (403) in CI and for private repos.
	if token := resolveGitHubToken("github.com"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	fmt.Printf("ℹ github: fetching releases for %s...\n", tool)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, NewBackendError("github", tool, "failed to fetch releases", err)
	}
	fmt.Printf("✓ github: received API response with status %d\n", resp.StatusCode)
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
func (g *GitHubBackend) findMatchingAsset(assets []githubAsset, platform Platform, tool string) (*githubAsset, string) {
	bestAsset := g.findMatchingAssetOnly(assets, platform, tool)
	if bestAsset == nil {
		return nil, ""
	}

	// Now find the checksum for this specific asset
	var checksumAsset *githubAsset
	for i := range assets {
		asset := &assets[i]
		nameLower := strings.ToLower(asset.Name)
		if strings.HasSuffix(nameLower, ".sha256") ||
			strings.HasSuffix(nameLower, ".sha256sum") ||
			strings.Contains(nameLower, "checksums") ||
			strings.Contains(nameLower, "sha256sums") {
			checksumAsset = asset
			break
		}
	}

	var checksum string
	if checksumAsset != nil {
		checksumMap := g.parseChecksumFile(checksumAsset.BrowserDownloadURL)
		if checksumMap != nil {
			checksum = checksumMap[bestAsset.Name]
		}
	}

	return bestAsset, checksum
}

// matchesPlatform is kept for backward compatibility but calls calculateAssetScore internally.
func (g *GitHubBackend) matchesPlatform(assetName string, platform Platform) bool {
	return g.calculateAssetScore(assetName, platform, "") > 0
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
