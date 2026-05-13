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

// AquaBackend implements the Backend interface for Aqua registry.
// Aqua is a declarative CLI version manager that provides a curated registry of tools.
type AquaBackend struct {
	client      *http.Client
	registryURL string
}

// NewAquaBackend creates a new Aqua backend.
func NewAquaBackend() *AquaBackend {
	return &AquaBackend{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		registryURL: "https://raw.githubusercontent.com/aquaproj/aqua-registry/main/pkgs",
	}
}

// Name returns the backend identifier.
func (a *AquaBackend) Name() string {
	return "aqua"
}

// aquaPackage represents an Aqua package definition.
type aquaPackage struct {
	Type          string            `json:"type"`
	RepoOwner     string            `json:"repo_owner"`
	RepoName      string            `json:"repo_name"`
	Asset         string            `json:"asset,omitempty"`
	Files         []aquaFile        `json:"files,omitempty"`
	Replacements  map[string]string `json:"replacements,omitempty"`
	Format        string            `json:"format,omitempty"`
	Overrides     []aquaOverride    `json:"overrides,omitempty"`
	VersionFilter string            `json:"version_filter,omitempty"`
	VersionPrefix string            `json:"version_prefix,omitempty"`
}

// aquaFile represents a file entry in an Aqua package.
type aquaFile struct {
	Name string `json:"name"`
	Src  string `json:"src,omitempty"`
}

// aquaOverride represents platform-specific overrides.
type aquaOverride struct {
	GOOS   string     `json:"goos,omitempty"`
	GOARCH string     `json:"goarch,omitempty"`
	Asset  string     `json:"asset,omitempty"`
	Files  []aquaFile `json:"files,omitempty"`
	Format string     `json:"format,omitempty"`
}

// ListVersions returns all available versions from Aqua registry.
func (a *AquaBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	pkg, err := a.fetchPackageMetadata(ctx, tool)
	if err != nil {
		return nil, err
	}

	// For GitHub-based packages, fetch releases
	if pkg.Type == "github_release" || pkg.Type == "github_archive" {
		return a.listGitHubVersions(ctx, tool, pkg, platform)
	}

	return nil, NewBackendError("aqua", tool, fmt.Sprintf("unsupported package type: %s", pkg.Type), nil)
}

// ResolveVersion resolves a version request to a concrete version.
func (a *AquaBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	versions, err := a.ListVersions(ctx, tool, platform)
	if err != nil {
		return nil, err
	}

	// Handle special version requests
	switch versionRequest {
	case "latest", "stable":
		if len(versions) > 0 {
			return &versions[0], nil
		}
		return nil, NewBackendError("aqua", tool, "no versions available", nil)

	default:
		// Exact version match
		versionRequest = strings.TrimPrefix(versionRequest, "v")
		for _, v := range versions {
			if v.Version == versionRequest {
				return &v, nil
			}
		}
		return nil, NewBackendError("aqua", tool, fmt.Sprintf("version %s not found", versionRequest), nil)
	}
}

// GetDownloadInfo retrieves download information for a specific version.
func (a *AquaBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return a.ResolveVersion(ctx, tool, version, platform)
}

// SupportsChecksum indicates whether this backend provides checksums.
func (a *AquaBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (a *AquaBackend) SupportsGPG() bool {
	return false
}

// AttestationType returns the type of attestation verification supported.
func (a *AquaBackend) AttestationType() string {
	return "SLSA"
}

// fetchPackageMetadata fetches package metadata from Aqua registry.
func (a *AquaBackend) fetchPackageMetadata(ctx context.Context, tool string) (*aquaPackage, error) {
	// Tool format: "owner/repo" or "package-name"
	// Convert to registry path
	var registryPath string
	if strings.Contains(tool, "/") {
		parts := strings.Split(tool, "/")
		registryPath = fmt.Sprintf("%s/%s/%s/pkg.yaml", a.registryURL, parts[0], parts[1])
	} else {
		// Try to find in standard registry
		registryPath = fmt.Sprintf("%s/%s/pkg.yaml", a.registryURL, tool)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", registryPath, nil)
	if err != nil {
		return nil, NewBackendError("aqua", tool, "failed to create request", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, NewBackendError("aqua", tool, "failed to fetch package metadata", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError("aqua", tool, fmt.Sprintf("package not found in registry (status %d)", resp.StatusCode), nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewBackendError("aqua", tool, "failed to read package metadata", err)
	}

	var pkg aquaPackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, NewBackendError("aqua", tool, "failed to parse package metadata", err)
	}

	return &pkg, nil
}

// listGitHubVersions lists versions from GitHub releases for an Aqua package.
func (a *AquaBackend) listGitHubVersions(ctx context.Context, tool string, pkg *aquaPackage, platform Platform) ([]VersionInfo, error) {
	repoPath := fmt.Sprintf("%s/%s", pkg.RepoOwner, pkg.RepoName)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repoPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, NewBackendError("aqua", tool, "failed to create request", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, NewBackendError("aqua", tool, "failed to fetch releases", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, NewBackendError("aqua", tool, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)), nil)
	}

	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, NewBackendError("aqua", tool, "failed to decode releases", err)
	}

	var versions []VersionInfo
	for _, release := range releases {
		if release.Draft {
			continue
		}

		// Apply version filter if specified
		if pkg.VersionFilter != "" && !strings.Contains(release.TagName, pkg.VersionFilter) {
			continue
		}

		// Build download URL using Aqua's asset template
		downloadURL := a.buildDownloadURL(pkg, release.TagName, platform)
		if downloadURL == "" {
			continue
		}

		version := strings.TrimPrefix(release.TagName, "v")
		if pkg.VersionPrefix != "" {
			version = strings.TrimPrefix(version, pkg.VersionPrefix)
		}

		versions = append(versions, VersionInfo{
			Version:     version,
			DownloadURL: downloadURL,
			Checksum:    "", // Aqua doesn't always provide checksums in metadata
			Platform:    platform,
			Metadata: map[string]string{
				"tag_name":   release.TagName,
				"name":       release.Name,
				"repo":       repoPath,
				"created_at": release.CreatedAt,
			},
		})
	}

	// Sort versions (descending)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	if len(versions) == 0 {
		return nil, NewBackendError("aqua", tool, fmt.Sprintf("no releases found for platform %s", platform.String()), nil)
	}

	return versions, nil
}

// buildDownloadURL constructs the download URL using Aqua's asset template.
func (a *AquaBackend) buildDownloadURL(pkg *aquaPackage, version string, platform Platform) string {
	asset := pkg.Asset

	// Check for platform-specific overrides
	for _, override := range pkg.Overrides {
		if (override.GOOS == "" || override.GOOS == platform.OS) &&
			(override.GOARCH == "" || override.GOARCH == platform.Arch) {
			if override.Asset != "" {
				asset = override.Asset
			}
			break
		}
	}

	if asset == "" {
		return ""
	}

	// Apply replacements
	replacements := map[string]string{
		"version": strings.TrimPrefix(version, "v"),
		"Version": strings.TrimPrefix(version, "v"),
		"os":      platform.OS,
		"arch":    platform.Arch,
		"OS":      platform.OS,
		"ARCH":    platform.Arch,
	}

	// Add custom replacements from package
	for k, v := range pkg.Replacements {
		replacements[k] = v
	}

	// Replace placeholders in asset template
	url := asset
	for k, v := range replacements {
		url = strings.ReplaceAll(url, "{{."+k+"}}", v)
		url = strings.ReplaceAll(url, "${"+k+"}", v)
	}

	// If URL is relative, make it absolute
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			pkg.RepoOwner, pkg.RepoName, version, url)
	}

	return url
}

func (a *AquaBackend) IsRecommended() bool {
	return true
}

func (a *AquaBackend) IsScriptless() bool {
	return true
}

func (a *AquaBackend) GetReach() string {
	return "Large"
}

func (a *AquaBackend) IsStable() bool {
	return true
}

func (a *AquaBackend) SupportsOffline() bool {
	return false
}
