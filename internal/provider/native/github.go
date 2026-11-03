// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	githubHandlerCache  sync.Map // key: owner/repo|apiBase|proxy -> []VersionInfo
	githubHandlerFlight singleflight.Group
)

// ClearGithubHandlerCache clears the in-memory cache of GitHubHandler releases (useful for testing).
func ClearGithubHandlerCache() {
	githubHandlerCache.Range(func(key, value interface{}) bool {
		githubHandlerCache.Delete(key)
		return true
	})
}

// GithubHandler handles tools distributed via GitHub releases.
// It specifically targets python-build-standalone style release naming.
type GithubHandler struct {
	Owner string
	Repo  string
}

func (h *GithubHandler) Name() string {
	return "github_release"
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (h *GithubHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}

	// Support GitHub Proxy Acceleration
	githubProxy := ""
	if p := env.Get("GITHUB_PROXY"); p != "" {
		enableProxy := env.Get("ENABLE_GITHUB_PROXY")
		if enableProxy != "0" && strings.ToLower(enableProxy) != "false" {
			githubProxy = p
		}
	} else if env.Get("ENABLE_GITHUB_PROXY") == "1" {
		githubProxy = "https://gh-proxy.com/" // Default fallback
	}

	// BaseURL for github is used to construct the API URL if needed,
	// but we primarily use Owner/Repo.
	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	cacheKey := fmt.Sprintf("%s/%s|%s|%s", h.Owner, h.Repo, apiBase, githubProxy)
	if val, ok := githubHandlerCache.Load(cacheKey); ok {
		return val.([]VersionInfo), nil
	}

	// Tier 1: Check persistent disk cache (12 hours TTL)
	var cachedVersions []VersionInfo
	hit, etag, cachedVersions := readNativeDiskCacheMeta(h.Name(), cacheKey, &cachedVersions, 12*time.Hour)
	if hit && len(cachedVersions) > 0 {
		githubHandlerCache.Store(cacheKey, cachedVersions)
		return cachedVersions, nil
	}

	res, err, _ := githubHandlerFlight.Do(cacheKey, func() (interface{}, error) {
		if val, ok := githubHandlerCache.Load(cacheKey); ok {
			return val, nil
		}
		if hit, _, cached := readNativeDiskCacheMeta(h.Name(), cacheKey, &cachedVersions, 12*time.Hour); hit && len(cached) > 0 {
			githubHandlerCache.Store(cacheKey, cached)
			return cached, nil
		}

		apiURL := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", apiBase, h.Owner, h.Repo)

		var resp *http.Response
		var lastErr error
		var newETag, lastMod string

		timeout := 5 * time.Second
		if pkgHttp.MockTransport != nil {
			timeout = 60 * time.Second
		}

		for i := 0; i < 3; i++ {
			client := pkgHttp.NewClientWithTimeout(timeout)
			req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
			if err != nil {
				return nil, err
			}

			// GitHub API requires a User-Agent header
			req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
			req.Header.Set("Accept", "application/vnd.github+json")

			// Add GitHub token if available to increase rate limits
			token := getGitHubToken()
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			// Conditional request (ETag) to save GitHub rate limit quota
			if etag != "" {
				req.Header.Set("If-None-Match", etag)
			}

			resp, err = client.Do(req)
			if err == nil {
				if resp.StatusCode == http.StatusNotModified {
					resp.Body.Close()
					touchNativeDiskCache(h.Name(), cacheKey)
					if len(cachedVersions) > 0 {
						githubHandlerCache.Store(cacheKey, cachedVersions)
						return cachedVersions, nil
					}
				} else if resp.StatusCode == http.StatusOK {
					newETag = resp.Header.Get("ETag")
					lastMod = resp.Header.Get("Last-Modified")
					lastErr = nil
					break
				} else {
					lastErr = fmt.Errorf("attempt %d: github api returned status %d", i+1, resp.StatusCode)
					resp.Body.Close()
				}
			} else {
				lastErr = fmt.Errorf("attempt %d: %w", i+1, err)
			}

			// Backoff before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
			}
		}

		if lastErr != nil {
			if pkgHttp.MockTransport == nil {
				var staleVersions []VersionInfo
				if readNativeDiskCacheStale(h.Name(), cacheKey, &staleVersions) && len(staleVersions) > 0 {
					githubHandlerCache.Store(cacheKey, staleVersions)
					return staleVersions, nil
				}
			}
			return nil, fmt.Errorf("github api call failed after 3 attempts (base: %s): %w", apiBase, lastErr)
		}
		defer resp.Body.Close()

		var releases []ghRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

		var versions []VersionInfo
		for _, rel := range releases {
			version := rel.TagName
			if h.Repo != "" {
				version = strings.TrimPrefix(version, h.Repo+"-v")
				version = strings.TrimPrefix(version, h.Repo+"-")
				version = strings.TrimPrefix(version, h.Repo+"_v")
				version = strings.TrimPrefix(version, h.Repo+"_")
			}
			version = strings.TrimPrefix(version, "v")
			version = strings.TrimPrefix(version, "V")

			// Map to store signatures for later matching
			sigs := make(map[string]string)
			for _, a := range rel.Assets {
				downloadURL := a.BrowserDownloadURL

				if strings.HasSuffix(a.Name, ".asc") || strings.HasSuffix(a.Name, ".sig") {
					sigs[a.Name] = downloadURL
				}
			}

			var assets []Asset
			for _, a := range rel.Assets {
				if strings.HasSuffix(a.Name, ".asc") || strings.HasSuffix(a.Name, ".sig") || strings.HasSuffix(a.Name, ".sha256") {
					continue
				}

				os, arch := h.detectPlatform(a.Name)
				if os == "" || arch == "" {
					continue
				}

				downloadURL := a.BrowserDownloadURL

				asset := Asset{
					Filename: a.Name,
					URL:      downloadURL,
					OS:       os,
					Arch:     arch,
					Metadata: make(map[string]string),
				}

				// Try to find matching signature
				if sigURL, ok := sigs[a.Name+".asc"]; ok {
					asset.SignatureURL = sigURL
				} else if sigURL, ok := sigs[a.Name+".sig"]; ok {
					asset.SignatureURL = sigURL
				}

				assets = append(assets, asset)
			}

			if len(assets) > 0 {
				versions = append(versions, VersionInfo{
					Version: version,
					Assets:  assets,
				})
			}
		}

		githubHandlerCache.Store(cacheKey, versions)
		if len(versions) > 0 {
			writeNativeDiskCacheWithETag(h.Name(), cacheKey, versions, newETag, lastMod)
		}
		return versions, nil
	})

	if err != nil {
		return nil, err
	}

	return res.([]VersionInfo), nil
}

// ResolveTag fetches a specific release by tag name directly when not found in recent releases.
func (h *GithubHandler) ResolveTag(ctx context.Context, version string) (*VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}

	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	tagCandidates := []string{
		h.Repo + "-v" + version,
		h.Repo + "-" + version,
		"v" + version,
		version,
	}

	client := pkgHttp.NewClientWithTimeout(30 * time.Second)
	token := getGitHubToken()

	for _, tag := range tagCandidates {
		apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", apiBase, h.Owner, h.Repo, tag)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		req.Header.Set("Accept", "application/vnd.github+json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				_ = resp.Body.Close()
			}
			continue
		}

		var rel ghRelease
		err = json.NewDecoder(resp.Body).Decode(&rel)
		_ = resp.Body.Close()
		if err != nil {
			continue
		}

		var assets []Asset
		for _, a := range rel.Assets {
			if strings.HasSuffix(a.Name, ".asc") || strings.HasSuffix(a.Name, ".sig") || strings.HasSuffix(a.Name, ".sha256") {
				continue
			}
			osName, arch := h.detectPlatform(a.Name)
			if osName == "" || arch == "" {
				continue
			}
			downloadURL := a.BrowserDownloadURL
			assets = append(assets, Asset{
				Filename: a.Name,
				URL:      downloadURL,
				OS:       osName,
				Arch:     arch,
				Metadata: make(map[string]string),
			})
		}

		if len(assets) > 0 {
			return &VersionInfo{
				Version: version,
				Assets:  assets,
			}, nil
		}
	}

	return nil, fmt.Errorf("github release tag not found for version %s", version)
}

func (h *GithubHandler) detectPlatform(filename string) (string, string) {
	filename = strings.ToLower(filename)

	// Skip installer formats and packages that UniRTM doesn't handle natively
	if strings.HasSuffix(filename, ".dmg") ||
		strings.HasSuffix(filename, ".pkg") ||
		strings.HasSuffix(filename, ".msi") ||
		strings.HasSuffix(filename, ".deb") ||
		strings.HasSuffix(filename, ".rpm") {
		return "", ""
	}

	var os, arch string

	// OS Detection
	if strings.Contains(filename, "darwin") || strings.Contains(filename, "macos") || strings.Contains(filename, "apple") || strings.Contains(filename, "mac") {
		os = "darwin"
	} else if strings.Contains(filename, "windows") || strings.Contains(filename, "win") {
		os = "windows"
	} else {
		// Fallback: Linux distributions and libc variants are so fragmented
		// (ubuntu, alpine, debian, centos, rhel, musl, glibc, el8, etc.)
		// that it's safest to assume any non-Mac/Win binary is for Linux.
		os = "linux"
	}

	// Arch Detection
	if strings.Contains(filename, "x86_64") || strings.Contains(filename, "amd64") || strings.Contains(filename, "x64") {
		arch = "amd64"
	} else if strings.Contains(filename, "aarch64") || strings.Contains(filename, "arm64") {
		arch = "arm64"
	} else if strings.Contains(filename, "i686") || strings.Contains(filename, "386") || strings.Contains(filename, "x86") {
		arch = "386"
	} else if strings.Contains(filename, "universal") || strings.Contains(filename, "all") {
		arch = "universal"
	}

	return os, arch
}
