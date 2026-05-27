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

// NinjaHandler handles Ninja build tool versions via GitHub releases.
type NinjaHandler struct {
	GithubHandler
}

func (h *NinjaHandler) Name() string {
	return "ninja"
}

func (h *NinjaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "ninja-build"
	h.Repo = "ninja"

	// We fetch directly because GithubHandler filters out assets without clear os/arch
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
				assets = append(assets, Asset{
					Filename: a.Name,
					URL:      a.BrowserDownloadURL,
					OS:       osName,
					Arch:     "amd64", // Ninja binaries are usually x86_64
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
