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
)

// PythonHandler specifically handles python-build-standalone.
type PythonHandler struct {
	Owner string
	Repo  string
}

func (h *PythonHandler) Name() string {
	return "python_standalone"
}

func (h *PythonHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Support GitHub Proxy Acceleration
	githubProxy := ""
	if env.Get("ENABLE_GITHUB_PROXY") == "1" {
		githubProxy = env.Get("GITHUB_PROXY")
		if githubProxy == "" {
			githubProxy = "https://gh-proxy.com/"
		}
	}

	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")
	// Use smaller per_page to avoid 504
	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=10", apiBase, h.Owner, h.Repo)

	var resp *http.Response
	var lastErr error
	var bodyBytes []byte

	for i := 0; i < 3; i++ {
		client := pkgHttp.NewClientWithTimeout(60 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		req.Header.Set("Accept", "application/vnd.github+json")

		token := env.Get("GITHUB_TOKEN")
		if token == "" {
			token = env.Get("GH_TOKEN")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: request failed: %w", i+1, err)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("attempt %d: github api returned status %d", i+1, resp.StatusCode)
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}

		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: reading body failed: %w", i+1, err)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
			continue
		}

		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("github api call failed after 3 attempts (base: %s): %w", apiBase, lastErr)
	}

	var releases []ghRelease
	if err := json.Unmarshal(bodyBytes, &releases); err != nil {
		return nil, err
	}

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
			if existing, exists := vi.bestAssets[platKey]; !exists || score > existing.score {
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
// Debug builds are strictly excluded (-1). Standard standalone install_only distributions
// receive highest priority, followed by pgo and shared variants.
func pythonAssetPreference(name string) int {
	nameLower := strings.ToLower(name)

	// Explicitly exclude debug packages
	if strings.Contains(nameLower, "debug") {
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
