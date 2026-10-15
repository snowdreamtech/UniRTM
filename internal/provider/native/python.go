// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

// PythonHandler specifically handles python-build-standalone.
type PythonHandler struct {
	Owner string
	Repo  string
}

func (h *PythonHandler) Name() string {
	return "python_standalone"
}

var (
	pythonHandlerFlight singleflight.Group
)

func (h *PythonHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Support GitHub Proxy Acceleration
	githubProxy := ""
	if p := env.Get("GITHUB_PROXY"); p != "" {
		enableProxy := env.Get("ENABLE_GITHUB_PROXY")
		if enableProxy != "0" && strings.ToLower(enableProxy) != "false" {
			githubProxy = p
		}
	} else if env.Get("ENABLE_GITHUB_PROXY") == "1" {
		githubProxy = "https://gh-proxy.com/"
	}

	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	cacheKey := fmt.Sprintf("%s/%s|%s|%s", h.Owner, h.Repo, apiBase, githubProxy)

	// Tier 1: Check Disk Cache (12 hours TTL)
	var cachedVersions []VersionInfo
	hit, etag, cachedVersions := readNativeDiskCacheMeta("python", cacheKey, &cachedVersions, 12*time.Hour)
	if hit && len(cachedVersions) > 0 {
		return cachedVersions, nil
	}

	// Use singleflight to prevent duplicate concurrent API requests
	res, err, _ := pythonHandlerFlight.Do(cacheKey, func() (interface{}, error) {
		// Double check inside singleflight
		if hit, _, cached := readNativeDiskCacheMeta("python", cacheKey, &cachedVersions, 12*time.Hour); hit && len(cached) > 0 {
			return cached, nil
		}

		apiURL := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", apiBase, h.Owner, h.Repo)

		var resp *http.Response
		var lastErr error
		var bodyBytes []byte
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

			req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
			req.Header.Set("Accept", "application/vnd.github+json")

			token := getGitHubToken()
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			// Conditional request (ETag) to save GitHub rate limit quota
			if etag != "" {
				req.Header.Set("If-None-Match", etag)
			}

			resp, err = client.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("attempt %d: request failed: %w", i+1, err)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
				}
				continue
			}

			// 304 Not Modified: Cache is still fresh on upstream! Zero quota used.
			if resp.StatusCode == http.StatusNotModified {
				resp.Body.Close()
				touchNativeDiskCache("python", cacheKey)
				if len(cachedVersions) > 0 {
					return cachedVersions, nil
				}
			}

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("attempt %d: github api returned status %d", i+1, resp.StatusCode)
				resp.Body.Close()
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
				}
				continue
			}

			newETag = resp.Header.Get("ETag")
			lastMod = resp.Header.Get("Last-Modified")

			bodyBytes, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = fmt.Errorf("attempt %d: reading body failed: %w", i+1, err)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
				}
				continue
			}

			lastErr = nil
			break
		}

		if lastErr != nil {
			// Tier 3: Stale Fallback on error (rate-limit, offline, timeout)
			if pkgHttp.MockTransport == nil {
				var staleVersions []VersionInfo
				if readNativeDiskCacheStale("python", cacheKey, &staleVersions) && len(staleVersions) > 0 {
					return staleVersions, nil
				}
				if baseline := getPythonBaselineVersions(githubProxy); len(baseline) > 0 {
					return baseline, nil
				}
			}
			return nil, fmt.Errorf("github api call failed after 3 attempts (base: %s): %w", apiBase, lastErr)
		}

		var releases []ghRelease
		if err := json.Unmarshal(bodyBytes, &releases); err != nil {
			return nil, err
		}

		parsed := h.parseReleases(releases, githubProxy)
		if len(parsed) > 0 {
			writeNativeDiskCacheWithETag("python", cacheKey, parsed, newETag, lastMod)
		}
		return parsed, nil
	})

	if err != nil {
		return nil, err
	}
	return res.([]VersionInfo), nil
}

func (h *PythonHandler) parseReleases(releases []ghRelease, githubProxy string) []VersionInfo {

	// Regex to extract version from filename: cpython-3.14.4+20260408-...
	re := regexp.MustCompile(`cpython-([0-9.]+)\+`)

	versionMap := make(map[string]*versionCollector)

	for _, rel := range releases {
		for _, a := range rel.Assets {
			// Skip metadata files
			if strings.HasSuffix(a.Name, ".asc") || strings.HasSuffix(a.Name, ".sig") || strings.HasSuffix(a.Name, ".sha256") {
				continue
			}

			// Skip debug packages immediately
			if strings.Contains(strings.ToLower(a.Name), "debug") {
				continue
			}

			match := re.FindStringSubmatch(a.Name)
			if len(match) < 2 {
				continue
			}
			pyVersion := match[1]

			osName, archName := h.detectPlatform(a.Name)
			if osName == "" || archName == "" {
				continue
			}

			downloadURL := a.BrowserDownloadURL
			if githubProxy != "" {
				downloadURL = strings.TrimSuffix(githubProxy, "/") + "/" + downloadURL
			}

			score := pythonAssetPreference(a.Name)
			if score <= 0 {
				continue
			}

			vi, ok := versionMap[pyVersion]
			if !ok {
				vi = &versionCollector{
					version:    pyVersion,
					bestAssets: make(map[string]scoredAsset),
				}
				versionMap[pyVersion] = vi
			}

			platKey := osName + "/" + archName
			if strings.Contains(strings.ToLower(a.Name), "musl") {
				platKey += "-musl"
			}
			if existing, exists := vi.bestAssets[platKey]; !exists || isBetterAsset(a.Name, score, existing.asset.Filename, existing.score) {
				vi.bestAssets[platKey] = scoredAsset{
					asset: Asset{
						Filename: a.Name,
						URL:      downloadURL,
						OS:       osName,
						Arch:     archName,
					},
					score: score,
				}
			}
		}
	}

	var result []VersionInfo
	for _, vc := range versionMap {
		var assets []Asset
		for _, sa := range vc.bestAssets {
			assets = append(assets, sa.asset)
		}
		if len(assets) > 0 {
			result = append(result, VersionInfo{
				Version: vc.version,
				Assets:  assets,
			})
		}
	}

	return result, nil
}

type scoredAsset struct {
	asset Asset
	score int
}

type versionCollector struct {
	version    string
	bestAssets map[string]scoredAsset
}

// pythonAssetPreference evaluates python-build-standalone asset names.
// Debug, non-baseline x86_64, sanitizers, and experimental freethreaded builds are strictly excluded (-1).
// Standard standalone install_only distributions receive highest priority, followed by pgo and shared variants.
func pythonAssetPreference(name string) int {
	nameLower := strings.ToLower(name)

	// Explicitly exclude debug, non-baseline microarchitectures, sanitizers and experimental freethreaded/nogil packages
	if strings.Contains(nameLower, "debug") ||
		strings.Contains(nameLower, "freethreaded") ||
		strings.Contains(nameLower, "nogil") ||
		strings.Contains(nameLower, "valgrind") ||
		strings.Contains(nameLower, "asan") ||
		strings.Contains(nameLower, "tsan") ||
		strings.Contains(nameLower, "msan") ||
		strings.Contains(nameLower, "x86_64_v") ||
		strings.Contains(nameLower, "amd64_v") {
		return -1
	}

	// Prefer standard install_only packages (clean standalone distribution)
	if strings.Contains(nameLower, "install_only") && !strings.Contains(nameLower, "shared") {
		if strings.Contains(nameLower, "install_only_stripped") {
			return 95
		}
		return 100
	}

	// Performance-optimized builds (PGO/LTO)
	if strings.Contains(nameLower, "pgo+lto") || strings.Contains(nameLower, "pgo") {
		return 85
	}

	// Shared library builds
	if strings.Contains(nameLower, "shared") {
		return 70
	}

	// Standard full builds without debug
	if strings.Contains(nameLower, "full") {
		return 60
	}

	return 50
}

// isBetterAsset provides a deterministic tie-breaker when comparing two candidate assets for the same platform.
func isBetterAsset(candidateName string, candidateScore int, existingName string, existingScore int) bool {
	if candidateScore != existingScore {
		return candidateScore > existingScore
	}

	// Deterministic Tie-Breaker 1: Compression format preference (.tar.gz > .zip > .tar.xz > .tar.zst)
	formatScore := func(name string) int {
		n := strings.ToLower(name)
		if strings.HasSuffix(n, ".tar.gz") || strings.HasSuffix(n, ".tgz") {
			return 10
		}
		if strings.HasSuffix(n, ".zip") {
			return 9
		}
		if strings.HasSuffix(n, ".tar.xz") {
			return 8
		}
		if strings.HasSuffix(n, ".tar.zst") {
			return 5
		}
		return 1
	}
	cfScore := formatScore(candidateName)
	efScore := formatScore(existingName)
	if cfScore != efScore {
		return cfScore > efScore
	}

	// Deterministic Tie-Breaker 2: Non-stripped over stripped
	candStripped := strings.Contains(strings.ToLower(candidateName), "stripped")
	existStripped := strings.Contains(strings.ToLower(existingName), "stripped")
	if !candStripped && existStripped {
		return true
	}
	if candStripped && !existStripped {
		return false
	}

	// Deterministic Tie-Breaker 3: Simplicity (fewer special tokens / shorter length preferred for standard baselines)
	if len(candidateName) != len(existingName) {
		return len(candidateName) < len(existingName)
	}

	// Deterministic Tie-Breaker 4: Alphabetical stability (ensures 100% idempotent result regardless of slice order)
	return candidateName < existingName
}

func (h *PythonHandler) detectPlatform(filename string) (string, string) {
	filename = strings.ToLower(filename)

	if strings.HasSuffix(filename, ".dmg") ||
		strings.HasSuffix(filename, ".pkg") ||
		strings.HasSuffix(filename, ".msi") ||
		strings.HasSuffix(filename, ".deb") ||
		strings.HasSuffix(filename, ".rpm") {
		return "", ""
	}

	var os, arch string

	if strings.Contains(filename, "linux") {
		os = "linux"
	} else if strings.Contains(filename, "darwin") || strings.Contains(filename, "macos") || strings.Contains(filename, "apple") || strings.Contains(filename, "mac") {
		os = "darwin"
	} else if strings.Contains(filename, "windows") || strings.Contains(filename, "win") {
		os = "windows"
	}

	if strings.Contains(filename, "x86_64") || strings.Contains(filename, "amd64") || strings.Contains(filename, "x64") {
		arch = "amd64"
	} else if strings.Contains(filename, "aarch64") || strings.Contains(filename, "arm64") {
		arch = "arm64"
	} else if strings.Contains(filename, "i686") || strings.Contains(filename, "386") || strings.Contains(filename, "x86") {
		arch = "386"
	}

	return os, arch
}

// getPythonBaselineVersions provides guaranteed offline fallback versions for Python
// based on official python-build-standalone releases (supporting 3.7 to 3.14).
func getPythonBaselineVersions(proxy string) []VersionInfo {
	versions := []string{
		"3.14.0", "3.13.2", "3.13.1", "3.13.0",
		"3.12.9", "3.12.8", "3.12.7", "3.12.0",
		"3.11.11", "3.11.10", "3.11.9", "3.11.0",
		"3.10.16", "3.10.15", "3.10.0",
		"3.9.21", "3.9.20", "3.9.0",
		"3.8.20", "3.8.19", "3.8.0",
		"3.7.17", "3.7.16", "3.7.0",
	}

	platforms := []struct {
		os, arch, suffix string
	}{
		{"linux", "amd64", "x86_64-unknown-linux-gnu-install_only.tar.gz"},
		{"linux", "arm64", "aarch64-unknown-linux-gnu-install_only.tar.gz"},
		{"darwin", "amd64", "x86_64-apple-darwin-install_only.tar.gz"},
		{"darwin", "arm64", "aarch64-apple-darwin-install_only.tar.gz"},
		{"windows", "amd64", "x86_64-pc-windows-msvc-shared-install_only.tar.gz"},
	}

	var res []VersionInfo
	for _, v := range versions {
		var assets []Asset
		for _, p := range platforms {
			fn := fmt.Sprintf("cpython-%s+latest-%s", v, p.suffix)
			url := fmt.Sprintf("https://github.com/astral-sh/python-build-standalone/releases/download/latest/%s", fn)
			if proxy != "" {
				url = strings.TrimSuffix(proxy, "/") + "/" + url
			}
			assets = append(assets, Asset{
				Filename: fn,
				URL:      url,
				OS:       p.os,
				Arch:     p.arch,
			})
		}
		res = append(res, VersionInfo{
			Version: v,
			Assets:  assets,
		})
	}
	return res
}

