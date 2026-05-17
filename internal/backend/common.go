// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CommonAsset represents a generic asset from a hosting platform.
type CommonAsset struct {
	Name string
	URL  string
	Size int64
}

// CalculateAssetScore calculates a compatibility score for an asset name.
// Returns -1 if the asset is definitely incompatible.
func CalculateAssetScore(assetName string, platform Platform, toolName string) int {
	nameLower := strings.ToLower(assetName)

	// 1. Hard Exclusions (Negative Score)
	excludeSuffixes := []string{".sha256", ".sha256sum", ".md5", ".asc", ".sig", ".sha1", ".deb", ".rpm", ".msi", ".apk", ".pkg", ".txt", ".pdf", ".h", ".c", ".cpp", ".a", ".lib"}
	for _, suffix := range excludeSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return -1
		}
	}

	// Exclude non-runtime assets
	negatives := []string{"checksums", "sha256sums", "license", "source", "devel", "dev", "header", "static-lib", "manual", "doc", "man", "debug"}

	// Determine the short tool name to avoid false positives in negative check
	toolShortName := toolName
	if parts := strings.Split(toolName, "/"); len(parts) == 2 {
		toolShortName = parts[1]
	}
	toolShortName = strings.TrimPrefix(toolShortName, "github:")
	toolShortName = strings.ToLower(toolShortName)

	for _, neg := range negatives {
		if strings.Contains(nameLower, neg) {
			// If the negative keyword is part of the tool name itself, don't exclude.
			// Example: "addlicense" contains "license".
			if strings.Contains(toolShortName, neg) {
				continue
			}
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
	repoName := ""
	if parts := strings.Split(toolName, "/"); len(parts) == 2 {
		repoName = strings.ToLower(parts[1])
	}
	if repoName != "" && strings.Contains(nameLower, repoName) {
		score += 50
	}

	if strings.Contains(nameLower, "musl") {
		score -= 10
	}

	return score
}

// FindBestAsset finds the best matching asset for a platform from a list.
func FindBestAsset(assets []CommonAsset, platform Platform, toolName string) (*CommonAsset, int) {
	var bestAsset *CommonAsset
	bestScore := -1

	for i := range assets {
		asset := &assets[i]
		score := CalculateAssetScore(asset.Name, platform, toolName)
		if score > 0 && score > bestScore {
			bestScore = score
			bestAsset = asset
		}
	}
	return bestAsset, bestScore
}

// FetchAndParseChecksumFile downloads and parses a checksum file from a URL.
func FetchAndParseChecksumFile(ctx context.Context, client *http.Client, url string) map[string]string {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(req)
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
			// Handle format where filename is prefixed with * or space
			filename = strings.TrimPrefix(filename, "*")
			checksums[filename] = checksum
		}
	}

	return checksums
}

// FindChecksumForAsset attempts to find a matching checksum for an asset from a list of all assets.
func FindChecksumForAsset(ctx context.Context, client *http.Client, assets []CommonAsset, targetAsset *CommonAsset) string {
	if targetAsset == nil {
		return ""
	}

	// 1. Look for a checksum file
	var checksumAsset *CommonAsset
	for i := range assets {
		nameLower := strings.ToLower(assets[i].Name)
		if strings.HasSuffix(nameLower, ".sha256") ||
			strings.HasSuffix(nameLower, ".sha256sum") ||
			strings.Contains(nameLower, "checksums") ||
			strings.Contains(nameLower, "sha256sums") {
			checksumAsset = &assets[i]
			break
		}
	}

	if checksumAsset != nil {
		checksumMap := FetchAndParseChecksumFile(ctx, client, checksumAsset.URL)
		if checksumMap != nil {
			// Try exact match first
			if c, ok := checksumMap[targetAsset.Name]; ok {
				return c
			}
		}
	}

	return ""
}

// FindGPGSignatureForAsset attempts to find a matching GPG signature for an asset.
func FindGPGSignatureForAsset(assets []CommonAsset, targetAsset *CommonAsset) string {
	if targetAsset == nil {
		return ""
	}

	// Look for filename.asc or filename.sig
	ascName := targetAsset.Name + ".asc"
	sigName := targetAsset.Name + ".sig"

	for i := range assets {
		if assets[i].Name == ascName || assets[i].Name == sigName {
			return assets[i].URL
		}
	}

	return ""
}

// HostingProvider defines the interface for fetching data from a hosting platform (GitHub, GitLab, etc.).
type HostingProvider interface {
	Name() string
	FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error)
	FetchReleaseByTag(ctx context.Context, tool string, tag string) (*CommonRelease, error)
	GetAttestationType() string
	GetClient() *http.Client
}

// CommonRelease represents a generic release from any hosting platform.
type CommonRelease struct {
	Tag        string
	Prerelease bool
	Assets     []CommonAsset
}

// GenericResolveVersion implements the common logic for resolving a version request.
func GenericResolveVersion(ctx context.Context, p HostingProvider, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	releases, err := p.FetchReleases(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, release := range releases {
		// Find matching asset for platform
		bestAsset, _ := FindBestAsset(release.Assets, platform, tool)
		if bestAsset == nil {
			continue
		}

		v := strings.TrimPrefix(release.Tag, "v")
		versions = append(versions, VersionInfo{
			Version:     v,
			DownloadURL: bestAsset.URL,
			Platform:    platform,
			Metadata: map[string]string{
				"prerelease": fmt.Sprintf("%t", release.Prerelease),
			},
		})
	}

	if len(versions) == 0 {
		return nil, NewBackendError(p.Name(), tool, "no suitable releases found", nil)
	}

	// Resolution logic
	switch versionRequest {
	case "latest", "stable":
		for _, v := range versions {
			if v.Metadata["prerelease"] == "false" {
				return &v, nil
			}
		}
		return &versions[0], nil
	default:
		reqV := strings.TrimPrefix(versionRequest, "v")
		for _, v := range versions {
			if v.Version == reqV {
				return &v, nil
			}
		}
		return nil, NewBackendError(p.Name(), tool, "version not found", nil)
	}
}

// GenericGetDownloadInfo implements the common logic for retrieving download info.
func GenericGetDownloadInfo(ctx context.Context, p HostingProvider, tool string, version string, platform Platform) (*VersionInfo, error) {
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}

	release, err := p.FetchReleaseByTag(ctx, tool, tag)
	if err != nil {
		// try without 'v'
		tag = version
		release, err = p.FetchReleaseByTag(ctx, tool, tag)
		if err != nil {
			return nil, NewBackendError(p.Name(), tool, "release not found", err)
		}
	}

	bestAsset, _ := FindBestAsset(release.Assets, platform, tool)
	if bestAsset == nil {
		return nil, NewBackendError(p.Name(), tool, "no matching asset", nil)
	}

	checksum := FindChecksumForAsset(ctx, p.GetClient(), release.Assets, bestAsset)
	gpgSigURL := FindGPGSignatureForAsset(release.Assets, bestAsset)

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

// ProbeURL checks if a URL is accessible via HEAD request.
func ProbeURL(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
