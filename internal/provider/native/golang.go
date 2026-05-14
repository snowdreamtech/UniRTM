// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// GolangHandler handles the official Go download metadata from go.dev/dl/?mode=json.
type GolangHandler struct{}

type goFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Sha256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

type goVersion struct {
	Version string   `json:"version"`
	Stable  bool     `json:"stable"`
	Files   []goFile `json:"files"`
}

func (h *GolangHandler) Name() string {
	return "golang"
}

func (h *GolangHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Support mirror overrides (compatible with mise)
	if mirror := env.Get("GO_DOWNLOAD_MIRROR"); mirror != "" {
		baseURL = mirror
	}

	url := fmt.Sprintf("%s/?mode=json&include=all", strings.TrimSuffix(baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("golang: fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("golang: unexpected status code: %d", resp.StatusCode)
	}

	var goVersions []goVersion
	if err := json.NewDecoder(resp.Body).Decode(&goVersions); err != nil {
		return nil, fmt.Errorf("golang: decode metadata: %w", err)
	}

	var versions []VersionInfo
	for _, gv := range goVersions {
		vi := VersionInfo{
			Version: strings.TrimPrefix(gv.Version, "go"),
			IsLTS:   gv.Stable, // For Go, we treat stable as a primary indicator
		}

		for _, gf := range gv.Files {
			// We only care about archives (tar.gz/zip) for portability
			if gf.Kind != "archive" {
				continue
			}

			asset := Asset{
				URL:      fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), gf.Filename),
				Filename: gf.Filename,
				OS:       gf.OS,
				Arch:     gf.Arch,
				Checksum: gf.Sha256,
				Algo:     "sha256",
			}
			vi.Assets = append(vi.Assets, asset)
		}

		if len(vi.Assets) > 0 {
			versions = append(versions, vi)
		}
	}

	return versions, nil
}
