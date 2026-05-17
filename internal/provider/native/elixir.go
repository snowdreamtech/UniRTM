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

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// ElixirHandler handles Elixir versions via GitHub releases.
type ElixirHandler struct {
	GithubHandler
}

func (h *ElixirHandler) Name() string {
	return "elixir"
}

func (h *ElixirHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "elixir-lang"
	h.Repo = "elixir"

	// We call a modified logic that doesn't strictly filter by OS/Arch initially
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", h.Owner, h.Repo)

	client := pkgHttp.NewClientWithTimeout(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

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

	return versions, nil
}
