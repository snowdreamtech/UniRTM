// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"runtime"
	"strings"
	"time"
)

// FlutterHandler handles Flutter SDK versions via its official storage API.
type FlutterHandler struct{}

func (h *FlutterHandler) Name() string {
	return "flutter"
}

type flutterReleases struct {
	BaseURL  string `json:"base_url"`
	Releases []struct {
		Hash           string `json:"hash"`
		Channel        string `json:"channel"`
		Version        string `json:"version"`
		Archive        string `json:"archive"`
		DartSdkVersion string `json:"dart_sdk_version,omitempty"`
	} `json:"releases"`
}

func (h *FlutterHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Flutter requires platform-specific JSON URLs
	platform := runtime.GOOS
	if platform == "darwin" {
		platform = "macos"
	}
	
	apiURL := fmt.Sprintf("https://storage.googleapis.com/flutter_infra_release/releases/releases_%s.json", platform)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flutter api: returned status %d", resp.StatusCode)
	}

	var data flutterReleases
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	// Map to deduplicate versions (Flutter has multiple channels for the same version)
	seen := make(map[string]bool)

	for _, rel := range data.Releases {
		if rel.Channel != "stable" {
			continue // Prioritize stable releases
		}
		
		if seen[rel.Version] {
			continue
		}
		seen[rel.Version] = true

		url := fmt.Sprintf("%s/%s", data.BaseURL, rel.Archive)
		
		// Determine arch from archive name (heuristics)
		arch := "amd64"
		if strings.Contains(rel.Archive, "arm64") {
			arch = "arm64"
		}

		versions = append(versions, VersionInfo{
			Version: rel.Version,
			Assets: []Asset{
				{
					Filename: filepathBase(url),
					URL:      url,
					OS:       runtime.GOOS, // This JSON is already platform-specific
					Arch:     arch,
				},
			},
		})
	}

	return versions, nil
}
