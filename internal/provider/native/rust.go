// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/sysinfo"
	"golang.org/x/sync/singleflight"
)

// RustHandler handles Rust distributions from static.rust-lang.org.
type RustHandler struct {
	flight singleflight.Group
	cache  sync.Map
}

func (h *RustHandler) Name() string {
	return "rust"
}

type ghRustTag struct {
	Name string `json:"name"`
}

func (h *RustHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}

	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	cacheKey := fmt.Sprintf("rust-lang/rust|%s", apiBase)
	if val, ok := h.cache.Load(cacheKey); ok {
		return val.([]VersionInfo), nil
	}

	var cachedVersions []VersionInfo
	hit, etag, cachedVersions := readNativeDiskCacheMeta("rust", cacheKey, &cachedVersions, 12*time.Hour)
	if hit && len(cachedVersions) > 0 {
		h.cache.Store(cacheKey, cachedVersions)
		return cachedVersions, nil
	}

	res, err, _ := h.flight.Do(cacheKey, func() (interface{}, error) {
		if val, ok := h.cache.Load(cacheKey); ok {
			return val, nil
		}
		if hit, _, cached := readNativeDiskCacheMeta("rust", cacheKey, &cachedVersions, 12*time.Hour); hit && len(cached) > 0 {
			h.cache.Store(cacheKey, cached)
			return cached, nil
		}

		apiURL := fmt.Sprintf("%s/repos/rust-lang/rust/tags?per_page=100", apiBase)
		timeout := 5 * time.Second
		if pkgHttp.MockTransport != nil {
			timeout = 60 * time.Second
		}

		client := pkgHttp.NewClientWithTimeout(timeout)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return h.fallbackVersions(), nil
		}

		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		req.Header.Set("Accept", "application/vnd.github+json")
		token := getGitHubToken()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}

		resp, err := client.Do(req)
		if err != nil {
			var stale []VersionInfo
			if readNativeDiskCacheStale("rust", cacheKey, &stale) && len(stale) > 0 {
				h.cache.Store(cacheKey, stale)
				return stale, nil
			}
			return h.fallbackVersions(), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotModified {
			touchNativeDiskCache("rust", cacheKey)
			if len(cachedVersions) > 0 {
				h.cache.Store(cacheKey, cachedVersions)
				return cachedVersions, nil
			}
		}

		if resp.StatusCode != http.StatusOK {
			var stale []VersionInfo
			if readNativeDiskCacheStale("rust", cacheKey, &stale) && len(stale) > 0 {
				h.cache.Store(cacheKey, stale)
				return stale, nil
			}
			return h.fallbackVersions(), nil
		}

		newETag := resp.Header.Get("ETag")
		var tags []ghRustTag
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			return h.fallbackVersions(), nil
		}

		semverRegex := regexp.MustCompile(`^1\.\d+(\.\d+)?$`)
		var versions []VersionInfo
		seen := make(map[string]bool)

		for _, tag := range tags {
			ver := strings.TrimPrefix(tag.Name, "v")
			if !semverRegex.MatchString(ver) {
				continue
			}
			if seen[ver] {
				continue
			}
			seen[ver] = true

			versions = append(versions, VersionInfo{
				Version: ver,
				Assets:  h.generateAssets(ver),
			})
		}

		if len(versions) == 0 {
			return h.fallbackVersions(), nil
		}

		// Sort semver descending
		sort.Slice(versions, func(i, j int) bool {
			return compareSemver(versions[i].Version, versions[j].Version) > 0
		})

		writeNativeDiskCacheWithETag("rust", cacheKey, versions, newETag, "")
		h.cache.Store(cacheKey, versions)
		return versions, nil
	})

	if err != nil {
		return h.fallbackVersions(), nil
	}

	return res.([]VersionInfo), nil
}

func (h *RustHandler) fallbackVersions() []VersionInfo {
	fallback := []string{
		"1.85.0", "1.84.1", "1.84.0", "1.83.0", "1.82.0", "1.81.0", "1.80.1", "1.80.0",
		"1.79.0", "1.78.0", "1.77.2", "1.77.1", "1.77.0",
		"1.76.0", "1.75.0", "1.74.1", "1.74.0", "1.73.0", "1.72.1", "1.71.1", "1.70.0",
	}
	var res []VersionInfo
	for _, v := range fallback {
		res = append(res, VersionInfo{
			Version: v,
			Assets:  h.generateAssets(v),
		})
	}
	return res
}

func (h *RustHandler) generateAssets(version string) []Asset {
	var assets []Asset

	// Always use the canonical dist server for asset URLs so the lockfile is
	// never polluted by RUSTUP_DIST_SERVER. The mirror is applied at
	// download time via RustDistMirror in DownloadOptions (see http.go).
	const distServer = "https://static.rust-lang.org"

	// Common Rust targets
	targets := map[string]struct{ os, arch string }{
		"x86_64-apple-darwin":    {"darwin", "amd64"},
		"aarch64-apple-darwin":   {"darwin", "arm64"},
		"x86_64-pc-windows-msvc": {"windows", "amd64"},
	}

	if env.RuntimeGOOS == "linux" && sysinfo.IsMusl() {
		targets["x86_64-unknown-linux-musl"] = struct{ os, arch string }{"linux", "amd64"}
		targets["aarch64-unknown-linux-musl"] = struct{ os, arch string }{"linux", "arm64"}
	} else {
		targets["x86_64-unknown-linux-gnu"] = struct{ os, arch string }{"linux", "amd64"}
		targets["aarch64-unknown-linux-gnu"] = struct{ os, arch string }{"linux", "arm64"}
	}

	for target, platform := range targets {
		url := fmt.Sprintf("%s/dist/rust-%s-%s.tar.gz", distServer, version, target)
		if platform.os == "windows" {
			url = fmt.Sprintf("%s/dist/rust-%s-%s.zip", distServer, version, target)
		}

		assets = append(assets, Asset{
			Filename: fmt.Sprintf("rust-%s-%s.tar.gz", version, target),
			URL:      url,
			OS:       platform.os,
			Arch:     platform.arch,
			Metadata: make(map[string]string),
		})
	}

	return assets
}
