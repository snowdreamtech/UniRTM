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
	elixirFlight singleflight.Group
	elixirCache  sync.Map // key: "elixir_versions" -> []VersionInfo
)

// ClearElixirCache clears the in-memory cache for elixir versions.
func ClearElixirCache() {
	elixirCache.Range(func(key, _ interface{}) bool {
		elixirCache.Delete(key)
		return true
	})
}

// ElixirHandler handles Elixir versions via GitHub releases.
type ElixirHandler struct {
	GithubHandler
}

func (h *ElixirHandler) Name() string {
	return "elixir"
}

func (h *ElixirHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if val, ok := elixirCache.Load("elixir_versions"); ok {
		cached := val.([]VersionInfo)
		cp := make([]VersionInfo, len(cached))
		copy(cp, cached)
		return cp, nil
	}

	res, err, _ := elixirFlight.Do("elixir_versions", func() (interface{}, error) {
		if val, ok := elixirCache.Load("elixir_versions"); ok {
			return val.([]VersionInfo), nil
		}

		h.Owner = "elixir-lang"
		h.Repo = "elixir"

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

		var releases []ghRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

		var versions []VersionInfo
		for _, rel := range releases {
			version := strings.TrimPrefix(rel.TagName, "v")

			var assets []Asset
			for _, a := range rel.Assets {
				// Elixir precompiled assets are platform-independent
				if a.Name == "Precompiled.zip" || strings.HasPrefix(a.Name, "elixir-otp-") {
					// Mark as universal for all platforms
					assets = append(assets, Asset{
						Filename: a.Name,
						URL:      a.BrowserDownloadURL,
						OS:       "linux", // Map to all major ones for resolution
						Arch:     "amd64",
					})
					assets = append(assets, Asset{
						Filename: a.Name,
						URL:      a.BrowserDownloadURL,
						OS:       "darwin",
						Arch:     "amd64",
					})
					assets = append(assets, Asset{
						Filename: a.Name,
						URL:      a.BrowserDownloadURL,
						OS:       "darwin",
						Arch:     "arm64",
					})
					assets = append(assets, Asset{
						Filename: a.Name,
						URL:      a.BrowserDownloadURL,
						OS:       "windows",
						Arch:     "amd64",
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

		elixirCache.Store("elixir_versions", versions)
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
