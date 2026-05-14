// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

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
	// Support GitHub Proxy Acceleration
	githubProxy := ""
	if env.Get("ENABLE_GITHUB_PROXY") == "1" {
		githubProxy = env.Get("GITHUB_PROXY")
		if githubProxy == "" {
			githubProxy = "https://gh-proxy.com/" // Default fallback
		}
	}

	// BaseURL for github is used to construct the API URL if needed, 
	// but we primarily use Owner/Repo.
	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")
	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", apiBase, h.Owner, h.Repo)

	var resp *http.Response
	var lastErr error

	for i := 0; i < 3; i++ {
		client := &http.Client{Timeout: 60 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		// GitHub API requires a User-Agent header
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
		req.Header.Set("Accept", "application/vnd.github+json")
		
		// Add GitHub token if available to increase rate limits
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
			resp.Body.Close()
		}

		// Backoff before retry
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

	var versions []VersionInfo
	for _, rel := range releases {
		version := strings.TrimPrefix(rel.TagName, "v")
		
		// Map to store signatures for later matching
		sigs := make(map[string]string)
		for _, a := range rel.Assets {
			downloadURL := a.BrowserDownloadURL
			if githubProxy != "" {
				downloadURL = strings.TrimSuffix(githubProxy, "/") + "/" + downloadURL
			}

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
			if githubProxy != "" {
				downloadURL = strings.TrimSuffix(githubProxy, "/") + "/" + downloadURL
			}

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

	return versions, nil
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
	if strings.Contains(filename, "linux") {
		os = "linux"
	} else if strings.Contains(filename, "darwin") || strings.Contains(filename, "macos") || strings.Contains(filename, "apple") || strings.Contains(filename, "mac") {
		os = "darwin"
	} else if strings.Contains(filename, "windows") || strings.Contains(filename, "win") {
		os = "windows"
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
