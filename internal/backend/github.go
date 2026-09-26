// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

var (
	githubSFG          singleflight.Group
	githubReleaseCache sync.Map // key: tool+"@"+tag -> *CommonRelease
	githubListCache    sync.Map // key: tool -> cachedReleaseList
)

// ClearGitHubCache clears GitHub in-memory caches. Mainly used for testing.
func ClearGitHubCache() {
	githubReleaseCache.Range(func(key, _ interface{}) bool {
		githubReleaseCache.Delete(key)
		return true
	})
	githubListCache.Range(func(key, _ interface{}) bool {
		githubListCache.Delete(key)
		return true
	})
}

// GitHubBackend implements the Backend interface using GenericReleaseManager.
type GitHubBackend struct {
	client       *http.Client
	sfg          *singleflight.Group
	releaseCache *sync.Map
	listCache    *sync.Map
}

type cachedReleaseList struct {
	releases  []CommonRelease
	etag      string
	fetchedAt time.Time
}

// NewGitHubBackend creates a new GitHub backend.
func NewGitHubBackend() *GitHubBackend {
	return &GitHubBackend{
		client:       pkgHttp.NewClientWithTimeout(30 * time.Second),
		sfg:          &githubSFG,
		releaseCache: &githubReleaseCache,
		listCache:    &githubListCache,
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

func (g *GitHubBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	return GenericResolveVersion(ctx, g, tool, versionRequest, platform)
}

func (g *GitHubBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfo(ctx, g, tool, version, platform)
}

// GetDownloadInfoWithPatterns is like GetDownloadInfo but accepts assetPatterns
// and platformKey to bypass heuristic scoring for the named platform.
func (g *GitHubBackend) GetDownloadInfoWithPatterns(ctx context.Context, tool, version, platformKey string, platform Platform, assetPatterns map[string]string) (*VersionInfo, error) {
	return GenericGetDownloadInfoWithPatterns(ctx, g, tool, version, platformKey, platform, assetPatterns)
}

// FetchReleases implements HostingProvider.
func (g *GitHubBackend) FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "github:")
	cacheKey := "list:" + tool

	var lastEtag string
	var lastReleases []CommonRelease
	if val, ok := githubListCache.Load(cacheKey); ok {
		if cached, ok := val.(cachedReleaseList); ok {
			if time.Since(cached.fetchedAt) < 15*time.Minute {
				return cached.releases, nil
			}
			lastEtag = cached.etag
			lastReleases = cached.releases
		}
	}

	v, err, _ := githubSFG.Do(cacheKey, func() (interface{}, error) {
		if val, ok := githubListCache.Load(cacheKey); ok {
			if cached, ok := val.(cachedReleaseList); ok {
				if time.Since(cached.fetchedAt) < 15*time.Minute {
					return cached.releases, nil
				}
				lastEtag = cached.etag
				lastReleases = cached.releases
			}
		}

		url := fmt.Sprintf("https://api.github.com/repos/%s/releases", tool)
		releases, newEtag, is304, err := g.fetchReleasesListHTTP(ctx, url, lastEtag)
		if err != nil {
			if len(lastReleases) > 0 {
				logger.Warn("FetchReleases: request failed, using stale cache", map[string]interface{}{"error": err.Error()})
				return lastReleases, nil
			}
			return nil, err
		}

		if is304 && len(lastReleases) > 0 {
			logger.Debug("FetchReleases: 304 Not Modified, reusing cached releases without consuming rate limit", map[string]interface{}{"tool": tool})
			githubListCache.Store(cacheKey, cachedReleaseList{
				releases:  lastReleases,
				etag:      lastEtag,
				fetchedAt: time.Now(),
			})
			return lastReleases, nil
		}

		githubListCache.Store(cacheKey, cachedReleaseList{
			releases:  releases,
			etag:      newEtag,
			fetchedAt: time.Now(),
		})
		return releases, nil
	})

	if err != nil {
		return nil, err
	}
	return v.([]CommonRelease), nil
}

func (g *GitHubBackend) fetchReleasesListHTTP(ctx context.Context, url, etag string) ([]CommonRelease, string, bool, error) {
	var resp *http.Response
	var err error
	var bodyBytes []byte

	for i := 0; i < 3; i++ {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if reqErr != nil {
			return nil, "", false, reqErr
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		if token := resolveGitHubToken("github.com"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = g.client.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil, "", false, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}

		if resp.StatusCode == http.StatusNotModified {
			resp.Body.Close()
			return nil, etag, true, nil
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			break
		}

		// Don't retry on 404
		if resp.StatusCode == http.StatusNotFound {
			return nil, "", false, fmt.Errorf("GitHub API status %d", resp.StatusCode)
		}

		// Check rate limits
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if sec, parseErr := strconv.Atoi(retryAfter); parseErr == nil && sec > 0 && sec <= 5 {
					select {
					case <-ctx.Done():
						return nil, "", false, ctx.Err()
					case <-time.After(time.Duration(sec) * time.Second):
						continue
					}
				}
			}
			return nil, "", false, parseGitHubRateLimitError(resp, bodyBytes)
		}

		select {
		case <-ctx.Done():
			return nil, "", false, ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	if len(bodyBytes) == 0 || (resp != nil && resp.StatusCode != http.StatusOK) {
		if resp != nil {
			return nil, "", false, parseGitHubRateLimitError(resp, bodyBytes)
		}
		return nil, "", false, fmt.Errorf("GitHub API request failed: %w", err)
	}

	newEtag := resp.Header.Get("ETag")

	var releases []githubRelease
	if err := json.Unmarshal(bodyBytes, &releases); err != nil {
		return nil, "", false, err
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
	return res, newEtag, false, nil
}

// FetchReleaseByTag implements HostingProvider.
func (g *GitHubBackend) FetchReleaseByTag(ctx context.Context, tool, tag string) (*CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "github:")
	cacheKey := tool + "@" + tag

	// 1. In-memory cache hit
	if val, ok := githubReleaseCache.Load(cacheKey); ok {
		if rel, ok := val.(*CommonRelease); ok {
			return rel, nil
		}
	}

	// 2. Persistent disk cache hit (immutable release tag, only for real network clients)
	if g.client.Transport == nil {
		if rel := readReleaseDiskCache(tool, tag); rel != nil {
			githubReleaseCache.Store(cacheKey, rel)
			return rel, nil
		}
	}

	// 3. Singleflight deduplication across concurrent callers
	v, err, _ := githubSFG.Do(cacheKey, func() (interface{}, error) {
		if val, ok := githubReleaseCache.Load(cacheKey); ok {
			return val.(*CommonRelease), nil
		}

		url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", tool, tag)
		rel, err := g.fetchReleaseByTagHTTP(ctx, url)
		if err != nil {
			return nil, err
		}

		githubReleaseCache.Store(cacheKey, rel)
		if g.client.Transport == nil {
			writeReleaseDiskCache(tool, tag, rel)
		}
		return rel, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*CommonRelease), nil
}

func (g *GitHubBackend) fetchReleaseByTagHTTP(ctx context.Context, url string) (*CommonRelease, error) {
	var resp *http.Response
	var err error
	var bodyBytes []byte

	for i := 0; i < 3; i++ {
		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		if token := resolveGitHubToken("github.com"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = g.client.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			break
		}

		// Don't retry on 404 — the requested tag doesn't exist
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}

		// Check rate limits
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if sec, parseErr := strconv.Atoi(retryAfter); parseErr == nil && sec > 0 && sec <= 5 {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(time.Duration(sec) * time.Second):
						continue
					}
				}
			}
			return nil, parseGitHubRateLimitError(resp, bodyBytes)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	if len(bodyBytes) == 0 || (resp != nil && resp.StatusCode != http.StatusOK) {
		if resp != nil {
			return nil, parseGitHubRateLimitError(resp, bodyBytes)
		}
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
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

func parseGitHubRateLimitError(resp *http.Response, body []byte) error {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	var ghErr struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
	}
	_ = json.Unmarshal(body, &ghErr)

	msg := ghErr.Message
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		if remaining == "0" || strings.Contains(strings.ToLower(msg), "rate limit") {
			resetTimeStr := ""
			if reset != "" {
				if sec, err := strconv.ParseInt(reset, 10, 64); err == nil {
					resetTime := time.Unix(sec, 0)
					resetTimeStr = fmt.Sprintf(" (resets at %s)", resetTime.Format("15:04:05 MST"))
				}
			}
			return fmt.Errorf("GitHub API rate limit exceeded%s: %s. Set GITHUB_TOKEN or GH_TOKEN to increase your limit from 60 to 5,000 requests/hour", resetTimeStr, msg)
		}
	}
	return fmt.Errorf("GitHub API status %d: %s", resp.StatusCode, msg)
}

func getReleaseDiskCachePath(tool, tag string) string {
	safeTool := strings.NewReplacer("/", "_", ":", "_").Replace(tool)
	safeTag := strings.NewReplacer("/", "_", ":", "_").Replace(tag)
	return filepath.Join(env.GetCacheDir(), "github_releases", safeTool, safeTag+".json")
}

func readReleaseDiskCache(tool, tag string) *CommonRelease {
	p := getReleaseDiskCachePath(tool, tag)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var rel CommonRelease
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil
	}
	return &rel
}

func writeReleaseDiskCache(tool, tag string, rel *CommonRelease) {
	if rel == nil {
		return
	}
	p := getReleaseDiskCachePath(tool, tag)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}
	data, err := json.Marshal(rel)
	if err != nil {
		return
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err == nil {
		_ = os.Rename(tmp, p)
	}
}

func (g *GitHubBackend) toCommonAssets(assets []githubAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.BrowserDownloadURL, Size: a.Size}
	}
	return res
}

// githubRelease and githubAsset kept as internal helpers.
type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	CreatedAt  string        `json:"created_at"`
	Assets     []githubAsset `json:"assets"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
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
