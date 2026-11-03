// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

var (
	forgejoFlight            singleflight.Group
	forgejoReleasesCache     sync.Map // tool -> []CommonRelease
	forgejoReleaseByTagCache sync.Map // tool@tag -> *CommonRelease
)

// ClearForgejoCache clears the in-memory Forgejo caches. Mainly used for testing.
func ClearForgejoCache() {
	forgejoReleasesCache.Range(func(key, _ interface{}) bool {
		forgejoReleasesCache.Delete(key)
		return true
	})
	forgejoReleaseByTagCache.Range(func(key, _ interface{}) bool {
		forgejoReleaseByTagCache.Delete(key)
		return true
	})
}

// ForgejoBackend implements the Backend interface using GenericReleaseManager.
type ForgejoBackend struct {
	client            *http.Client
	baseURL           string
	flight            *singleflight.Group
	releasesCache     *sync.Map
	releaseByTagCache *sync.Map
}

// NewForgejoBackend creates a new Forgejo backend.
func NewForgejoBackend() *ForgejoBackend {
	baseURL := env.Get("FORGEJO_API_URL")
	if baseURL == "" {
		baseURL = "https://codeberg.org/api/v1"
	}
	return &ForgejoBackend{
		client:            pkgHttp.NewClientWithTimeout(15 * time.Second),
		baseURL:           baseURL,
		flight:            &forgejoFlight,
		releasesCache:     &forgejoReleasesCache,
		releaseByTagCache: &forgejoReleaseByTagCache,
	}
}

func (b *ForgejoBackend) Name() string { return "forgejo" }

func (b *ForgejoBackend) Dependencies() []string {
	return nil
}

func (b *ForgejoBackend) GetClient() *http.Client { return b.client }
func (b *ForgejoBackend) GetAttestationType() string {
	return "SLSA"
}
func (b *ForgejoBackend) SupportsChecksum() bool { return true }
func (b *ForgejoBackend) SupportsGPG() bool      { return true }
func (b *ForgejoBackend) AttestationType() string {
	return b.GetAttestationType()
}

func (b *ForgejoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
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

func (b *ForgejoBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	return GenericResolveVersion(ctx, b, tool, versionRequest, platform)
}

func (b *ForgejoBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfo(ctx, b, tool, version, platform)
}

// FetchReleases implements HostingProvider.
func (b *ForgejoBackend) FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")

	if val, ok := forgejoReleasesCache.Load(tool); ok {
		if rels, ok := val.([]CommonRelease); ok {
			return rels, nil
		}
	}

	val, err, _ := forgejoFlight.Do("releases:"+tool, func() (interface{}, error) {
		apiURL := fmt.Sprintf("%s/repos/%s/releases", b.baseURL, tool)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		if token := env.Get("FORGEJO_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Forgejo API status %d", resp.StatusCode)
		}

		var releases []forgejoRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

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
				Assets:      b.toCommonAssets(r.Assets),
				PublishedAt: publishedAt,
			}
			forgejoReleaseByTagCache.Store(tool+"@"+r.TagName, &res[i])
			cleanTag := strings.TrimPrefix(r.TagName, "v")
			forgejoReleaseByTagCache.Store(tool+"@"+cleanTag, &res[i])
			forgejoReleaseByTagCache.Store(tool+"@v"+cleanTag, &res[i])
			if b.client.Transport == nil {
				writeHostingReleaseDiskCache("forgejo", tool, r.TagName, &res[i])
			}
		}

		forgejoReleasesCache.Store(tool, res)
		return res, nil
	})

	if err != nil {
		return nil, err
	}
	return val.([]CommonRelease), nil
}

// FetchReleaseByTag implements HostingProvider.
func (b *ForgejoBackend) FetchReleaseByTag(ctx context.Context, tool, tag string) (*CommonRelease, error) {
	tool = strings.TrimPrefix(tool, "forgejo:")

	cacheKey := tool + "@" + tag
	if val, ok := forgejoReleaseByTagCache.Load(cacheKey); ok {
		if cr, ok := val.(*CommonRelease); ok {
			return cr, nil
		}
	}

	var cachedRelease *CommonRelease
	if b.client.Transport == nil {
		if diskRel := readHostingReleaseDiskCache("forgejo", tool, tag); diskRel != nil {
			cachedRelease = diskRel
			if diskRel.ETag == "" {
				forgejoReleaseByTagCache.Store(cacheKey, diskRel)
				return diskRel, nil
			}
		}
	}

	val, err, _ := forgejoFlight.Do("tag:"+cacheKey, func() (interface{}, error) {
		if val, ok := forgejoReleaseByTagCache.Load(cacheKey); ok {
			if cr, ok := val.(*CommonRelease); ok {
				return cr, nil
			}
		}

		apiURL := fmt.Sprintf("%s/repos/%s/releases/tags/%s", b.baseURL, tool, tag)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		if cachedRelease != nil && cachedRelease.ETag != "" {
			req.Header.Set("If-None-Match", cachedRelease.ETag)
		}
		if token := env.Get("FORGEJO_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotModified && cachedRelease != nil {
			forgejoReleaseByTagCache.Store(cacheKey, cachedRelease)
			return cachedRelease, nil
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}

		var r forgejoRelease
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, err
		}

		var publishedAt time.Time
		if r.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
				publishedAt = t
			}
		}
		cr := &CommonRelease{
			Tag:         r.TagName,
			Assets:      b.toCommonAssets(r.Assets),
			PublishedAt: publishedAt,
			ETag:        resp.Header.Get("ETag"),
		}
		forgejoReleaseByTagCache.Store(cacheKey, cr)
		if b.client.Transport == nil {
			writeHostingReleaseDiskCache("forgejo", tool, tag, cr)
		}
		return cr, nil
	})

	if err != nil {
		return nil, err
	}
	return val.(*CommonRelease), nil
}

func (b *ForgejoBackend) toCommonAssets(assets []forgejoAsset) []CommonAsset {
	res := make([]CommonAsset, len(assets))
	for i, a := range assets {
		res[i] = CommonAsset{Name: a.Name, URL: a.URL, Size: a.Size}
	}
	return res
}

type forgejoRelease struct {
	TagName   string         `json:"tag_name"`
	CreatedAt string         `json:"created_at"`
	Assets    []forgejoAsset `json:"assets"`
}

type forgejoAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

func (b *ForgejoBackend) IsRecommended() bool {
	return true
}

func (b *ForgejoBackend) IsScriptless() bool {
	return true
}

func (b *ForgejoBackend) GetReach() string {
	return "Large"
}

func (b *ForgejoBackend) IsStable() bool {
	return true
}

func (b *ForgejoBackend) SupportsOffline() bool {
	return false
}
