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
	// BaseURL for github is used to construct the API URL if needed, 
	// but we primarily use Owner/Repo.
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", h.Owner, h.Repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: returned status %d", resp.StatusCode)
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, rel := range releases {
		// Filter out pre-releases if needed (tag naming varies)
		version := strings.TrimPrefix(rel.TagName, "v")
		
		var assets []Asset
		for _, a := range rel.Assets {
			// Basic heuristic for python-build-standalone assets
			// Format example: cpython-3.11.5+20230826-x86_64-unknown-linux-gnu-install_ready.tar.gz
			os, arch := h.detectPlatform(a.Name)
			if os == "" || arch == "" {
				continue
			}

			assets = append(assets, Asset{
				Filename: a.Name,
				URL:      a.BrowserDownloadURL,
				OS:       os,
				Arch:     arch,
			})
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
	
	// Skip debug or specific builds not intended for general use
	if strings.Contains(filename, "debug") || strings.Contains(filename, "pgo") {
		// keep it simple for now
	}

	var os, arch string

	// OS Detection
	if strings.Contains(filename, "linux") {
		os = "linux"
	} else if strings.Contains(filename, "darwin") || strings.Contains(filename, "macos") {
		os = "darwin"
	} else if strings.Contains(filename, "windows") {
		os = "windows"
	}

	// Arch Detection
	if strings.Contains(filename, "x86_64") || strings.Contains(filename, "amd64") {
		arch = "amd64"
	} else if strings.Contains(filename, "aarch64") || strings.Contains(filename, "arm64") {
		arch = "arm64"
	} else if strings.Contains(filename, "i686") || strings.Contains(filename, "386") {
		arch = "386"
	}

	// Special case for python-build-standalone which uses "apple-darwin"
	if strings.Contains(filename, "apple-darwin") {
		os = "darwin"
	}

	return os, arch
}
