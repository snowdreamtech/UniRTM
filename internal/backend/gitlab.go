// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

var (
	gitlabFlight            singleflight.Group
	gitlabReleasesCache     sync.Map // tool -> []CommonRelease
	gitlabReleaseByTagCache sync.Map // tool@tag -> *CommonRelease
)

// ClearGitlabCache clears the in-memory GitLab caches. Mainly used for testing.
func ClearGitlabCache() {
	gitlabReleasesCache.Range(func(key, _ interface{}) bool {
		gitlabReleasesCache.Delete(key)
		return true
	})
	gitlabReleaseByTagCache.Range(func(key, _ interface{}) bool {
		gitlabReleaseByTagCache.Delete(key)
		return true
	})
}

// GitlabBackend implements the Backend interface using GenericReleaseManager.
type GitlabBackend struct {
	client            *http.Client
	baseURL           string
	flight            *singleflight.Group
	releasesCache     *sync.Map
	releaseByTagCache *sync.Map
}

// NewGitlabBackend creates a new GitLab backend.
func NewGitlabBackend() *GitlabBackend {
	baseURL := env.Get("GITLAB_API_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	return &GitlabBackend{
		client:            pkgHttp.NewClientWithTimeout(15 * time.Second),
		baseURL:           baseURL,
		flight:            &gitlabFlight,
		releasesCache:     &gitlabReleasesCache,
		releaseByTagCache: &gitlabReleaseByTagCache,
	}
}

func (b *GitlabBackend) Name() string { return "gitlab" }

func (b *GitlabBackend) Dependencies() []string {
	return nil
}

func (b *GitlabBackend) GetClient() *http.Client { return b.client }
func (b *GitlabBackend) GetAttestationType() string {
	return "SLSA"
}
func (b *GitlabBackend) SupportsChecksum() bool { return true }
func (b *GitlabBackend) SupportsGPG() bool      { return true }
func (b *GitlabBackend) AttestationType() string {
	return b.GetAttestationType()
}

func (b *GitlabBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	releases, err := b.FetchReleases(ctx, tool)
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

func (b *GitlabBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	return GenericResolveVersion(ctx, b, tool, versionRequest, platform)
}

func (b *GitlabBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfo(ctx, b, tool, version, platform)
}

// FetchReleases implements HostingProvider.
func (b *GitlabBackend) FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "gitlab:")

	if val, ok := gitlabReleasesCache.Load(tool); ok {
		if rels, ok := val.([]CommonRelease); ok {
			return rels, nil
		}
	}

	val, err, _ := gitlabFlight.Do("releases:"+tool, func() (interface{}, error) {
		encodedRepo := url.PathEscape(tool)
		apiURL := fmt.Sprintf("%s/projects/%s/releases", b.baseURL, encodedRepo)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		if token := env.Get("GITLAB_TOKEN"); token != "" {
			req.Header.Set("PRIVATE-TOKEN", token)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitLab API status %d", resp.StatusCode)
		}

		var releases []gitlabRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

		res := make([]CommonRelease, len(releases))
		for i, r := range releases {
			var publishedAt time.Time
			if r.ReleasedAt != "" {
				if t, err := time.Parse(time.RFC3339, r.ReleasedAt); err == nil {
					publishedAt = t
				}
			}
			res[i] = CommonRelease{
				Tag:         r.TagName,
				Assets:      b.toCommonAssets(r.Assets.Links),
				PublishedAt: publishedAt,
			}
			gitlabReleaseByTagCache.Store(tool+"@"+r.TagName, &res[i])
			cleanTag := strings.TrimPrefix(r.TagName, "v")
			gitlabReleaseByTagCache.Store(tool+"@"+cleanTag, &res[i])
			gitlabReleaseByTagCache.Store(tool+"@v"+cleanTag, &res[i])
		}

		gitlabReleasesCache.Store(tool, res)
		return res, nil
	})

	if err != nil {
		return nil, err
	}
	return val.([]CommonRelease), nil
}

// FetchReleaseByTag implements HostingProvider.
func (b *GitlabBackend) FetchReleaseByTag(ctx context.Context, tool, tag string) (*CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "gitlab:")

	cacheKey := tool + "@" + tag
	if val, ok := gitlabReleaseByTagCache.Load(cacheKey); ok {
		if cr, ok := val.(*CommonRelease); ok {
			return cr, nil
		}
	}

	val, err, _ := gitlabFlight.Do("tag:"+cacheKey, func() (interface{}, error) {
		encodedRepo := url.PathEscape(tool)
		apiURL := fmt.Sprintf("%s/projects/%s/releases/%s", b.baseURL, encodedRepo, url.PathEscape(tag))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		if token := env.Get("GITLAB_TOKEN"); token != "" {
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

		var r gitlabRelease
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, err
		}

		var publishedAt time.Time
		if r.ReleasedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.ReleasedAt); err == nil {
				publishedAt = t
			}
		}
		cr := &CommonRelease{
			Tag:         r.TagName,
			Assets:      b.toCommonAssets(r.Assets.Links),
			PublishedAt: publishedAt,
		}
		gitlabReleaseByTagCache.Store(cacheKey, cr)
		return cr, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*CommonRelease), nil
}

func (b *GitlabBackend) toCommonAssets(assets []gitlabAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.URL}
	}
	return res
}

type gitlabRelease struct {
	TagName    string `json:"tag_name"`
	ReleasedAt string `json:"released_at"`
	Assets     struct {
		Links []gitlabAsset `json:"links"`
	} `json:"assets"`
}

type gitlabAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (b *GitlabBackend) IsRecommended() bool {
	return true
}

func (b *GitlabBackend) IsScriptless() bool {
	return true
}

func (b *GitlabBackend) GetReach() string {
	return "Large"
}

func (b *GitlabBackend) IsStable() bool {
	return true
}

func (b *GitlabBackend) SupportsOffline() bool {
	return false
}
