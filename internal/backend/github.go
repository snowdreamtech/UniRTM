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

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// GitHubBackend implements the Backend interface using GenericReleaseManager.
type GitHubBackend struct {
	client *http.Client
}

// NewGitHubBackend creates a new GitHub backend.
func NewGitHubBackend() *GitHubBackend {
	return &GitHubBackend{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}
}

func (g *GitHubBackend) Name() string { return "github" }

func (b *GitHubBackend) Dependencies() []string {
	return nil
}

func (g *GitHubBackend) GetClient() *http.Client { return g.client }
func (g *GitHubBackend) GetAttestationType() string {
	return "GitHub Attestation"
}
func (g *GitHubBackend) SupportsChecksum() bool { return true }
func (g *GitHubBackend) SupportsGPG() bool      { return true }
func (g *GitHubBackend) AttestationType() string {
	return g.GetAttestationType()
}

// ListVersions returns all available versions from GitHub Releases.
func (g *GitHubBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// For GitHub, we need to filter assets that match the platform
	releases, err := g.FetchReleases(ctx, tool)
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

func (g *GitHubBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return GenericResolveVersion(ctx, g, tool, versionRequest, platform)
}

func (g *GitHubBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfo(ctx, g, tool, version, platform)
}

// FetchReleases implements HostingProvider.
func (g *GitHubBackend) FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "github:")
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", tool)

	var resp *http.Response
	var err error
	var bodyBytes []byte

	for i := 0; i < 3; i++ {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if token := resolveGitHubToken("github.com"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = g.client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		break
	}

	if len(bodyBytes) == 0 {
		if resp != nil && resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("GitHub API request failed: %v", err)
	}

	var releases []githubRelease
	if err := json.Unmarshal(bodyBytes, &releases); err != nil {
		return nil, err
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].TagName > releases[j].TagName
	})

	res := make([]CommonRelease, len(releases))
	for i, r := range releases {
		var publishedAt time.Time
		if r.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
				publishedAt = t
			}
		}
		res[i] = CommonRelease{
			Tag:         r.TagName,
			Prerelease:  r.Prerelease,
			Assets:      g.toCommonAssets(r.Assets),
			PublishedAt: publishedAt,
		}
	}
	return res, nil
}

// FetchReleaseByTag implements HostingProvider.
func (g *GitHubBackend) FetchReleaseByTag(ctx context.Context, tool string, tag string) (*CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "github:")
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", tool, tag)

	var resp *http.Response
	var err error
	var bodyBytes []byte

	for i := 0; i < 3; i++ {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if token := resolveGitHubToken("github.com"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = g.client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		break
	}

	if len(bodyBytes) == 0 {
		if resp != nil && resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("GitHub API request failed: %v", err)
	}

	var r githubRelease
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		return nil, err
	}

	var publishedAt time.Time
	if r.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
			publishedAt = t
		}
	}

	return &CommonRelease{
		Tag:         r.TagName,
		Prerelease:  r.Prerelease,
		Assets:      g.toCommonAssets(r.Assets),
		PublishedAt: publishedAt,
	}, nil
}

func (g *GitHubBackend) toCommonAssets(assets []githubAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.BrowserDownloadURL, Size: a.Size}
	}
	return res
}

// githubRelease and githubAsset kept as internal helpers
type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
	CreatedAt  string        `json:"created_at"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func (g *GitHubBackend) IsRecommended() bool {
	return true
}

func (g *GitHubBackend) IsScriptless() bool {
	return true
}

func (g *GitHubBackend) GetReach() string {
	return "Large"
}

func (g *GitHubBackend) IsStable() bool {
	return true
}

func (g *GitHubBackend) SupportsOffline() bool {
	return false
}
