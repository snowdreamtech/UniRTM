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
	ninjaFlight singleflight.Group
	ninjaCache  sync.Map // key: "ninja_versions" -> []VersionInfo
)

// ClearNinjaCache clears the in-memory cache for ninja versions.
func ClearNinjaCache() {
	ninjaCache.Range(func(key, _ interface{}) bool {
		ninjaCache.Delete(key)
		return true
	})
}

// NinjaHandler handles Ninja build tool versions via GitHub releases.
type NinjaHandler struct {
	GithubHandler
}

func (h *NinjaHandler) Name() string {
	return "ninja"
}

func (h *NinjaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if val, ok := ninjaCache.Load("ninja_versions"); ok {
		cached := val.([]VersionInfo)
		cp := make([]VersionInfo, len(cached))
		copy(cp, cached)
		return cp, nil
	}

	res, err, _ := ninjaFlight.Do("ninja_versions", func() (interface{}, error) {
		if val, ok := ninjaCache.Load("ninja_versions"); ok {
			return val.([]VersionInfo), nil
		}

		h.Owner = "ninja-build"
		h.Repo = "ninja"

		apiBase := env.Get("GITHUB_API_BASEURL")
		if apiBase == "" {
			apiBase = "https://api.github.com"
		}
		apiBase = strings.TrimSuffix(apiBase, "/")
		apiURL := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=100", apiBase, h.Owner, h.Repo)

		client := pkgHttp.NewClientWithTimeout(15 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		if token := getGitHubToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var releases []struct {
			TagName string `json:"tag_name"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

		var versions []VersionInfo
		for _, rel := range releases {
			version := strings.TrimPrefix(rel.TagName, "v")
			var assets []Asset

			for _, a := range rel.Assets {
				osName := ""
				lowerName := strings.ToLower(a.Name)
				if strings.Contains(lowerName, "linux") {
					osName = "linux"
				} else if strings.Contains(lowerName, "mac") {
					osName = "darwin"
				} else if strings.Contains(lowerName, "win") {
					osName = "windows"
				}

				if osName != "" {
					archName := "amd64"
					if strings.Contains(lowerName, "arm64") || strings.Contains(lowerName, "aarch64") {
						archName = "arm64"
					}
					assets = append(assets, Asset{
						Filename: a.Name,
						URL:      a.BrowserDownloadURL,
						OS:       osName,
						Arch:     archName,
					})
				}
			}

			if len(assets) > 0 {
				versions = append(versions, VersionInfo{
					Version: version,
					Assets:  assets,
				})
			}
		}

		ninjaCache.Store("ninja_versions", versions)
		return versions, nil
	})

	if err != nil {
		return nil, err
	}

	cached := res.([]VersionInfo)
	cp := make([]VersionInfo, len(cached))
	copy(cp, cached)
	return cp, nil
}
