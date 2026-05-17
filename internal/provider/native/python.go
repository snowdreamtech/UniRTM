// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
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

	for i := 0; i < 3; i++ {
		client := pkgHttp.NewClientWithTimeout(60 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		req.Header.Set("Accept", "application/vnd.github+json")

		if token := env.Get("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			lastErr = nil
			break
		}

		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", i+1, err)
		} else {
			lastErr = fmt.Errorf("attempt %d: github api returned status %d", i+1, resp.StatusCode)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("github api call failed after 3 attempts (base: %s): %w", apiBase, lastErr)
	}
	defer resp.Body.Close()

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// Regex to extract version from filename: cpython-3.14.4+20260408-...
	re := regexp.MustCompile(`cpython-([0-9.]+)\+`)

	versionMap := make(map[string]*VersionInfo)

	for _, rel := range releases {
		for _, a := range rel.Assets {
			// Skip metadata files
			if strings.HasSuffix(a.Name, ".asc") || strings.HasSuffix(a.Name, ".sig") || strings.HasSuffix(a.Name, ".sha256") {
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

			vi, ok := versionMap[pyVersion]
			if !ok {
				vi = &VersionInfo{
					Version: pyVersion,
					Assets:  []Asset{},
				}
				versionMap[pyVersion] = vi
			}

			// Avoid duplicate assets for the same platform within the same version
			// (python-build-standalone often has multiple release types like debug, pgo+lto, etc.)
			// We prefer "pgo+lto" or "install_only" over "debug".
			isBetter := true
			for _, existing := range vi.Assets {
				if existing.OS == osName && existing.Arch == archName {
					// If we already have a non-debug one, and current is debug, skip
					if !strings.Contains(existing.Filename, "debug") && strings.Contains(a.Name, "debug") {
						isBetter = false
						break
					}
					// If current is pgo+lto and existing is not, it's better
					if strings.Contains(a.Name, "pgo+lto") && !strings.Contains(existing.Filename, "pgo+lto") {
						isBetter = true
						// We'll replace it later or just append.
						// To keep it simple, we just don't add if not better.
					}
				}
			}

			if isBetter {
				vi.Assets = append(vi.Assets, Asset{
					Filename: a.Name,
					URL:      downloadURL,
					OS:       osName,
					Arch:     archName,
				})
			}
		}
	}

	var result []VersionInfo
	for _, vi := range versionMap {
		if len(vi.Assets) > 0 {
			result = append(result, *vi)
		}
	}

	return result, nil
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
