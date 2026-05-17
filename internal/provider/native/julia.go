// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// JuliaHandler handles Julia language versions via official versions.json.
type JuliaHandler struct{}

type juliaVersion struct {
	Version string `json:"version"`
	Files   []struct {
		OS     string `json:"os"`
		Arch   string `json:"arch"`
		Kind   string `json:"kind"`
		URL    string `json:"url"`
		SHA256 string `json:"sha256"`
		ASC    string `json:"asc"`
	} `json:"files"`
}

func (h *JuliaHandler) Name() string {
	return "julia"
}

func (h *JuliaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	url := fmt.Sprintf("%s/versions.json", strings.TrimSuffix(baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("julia: fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	var jv map[string]juliaVersion
	if err := json.NewDecoder(resp.Body).Decode(&jv); err != nil {
		return nil, fmt.Errorf("julia: decode metadata: %w", err)
	}

	var versions []VersionInfo
	for vStr, v := range jv {
		vi := VersionInfo{
			Version: vStr,
		}

		for _, f := range v.Files {
			// Skip source and other kinds for now
			if f.Kind != "archive" && f.Kind != "installer" {
				continue
			}

			os, arch := mapPlatform(f.OS, f.Arch)
			if os == "" || arch == "" {
				continue
			}

			asset := Asset{
				URL:      f.URL,
				Filename: vStr + "-" + f.OS + "-" + f.Arch,
				OS:       os,
				Arch:     arch,
				Checksum: f.SHA256,
				Algo:     "sha256",
			}

			// If it has an embedded signature, we'll need to handle it.
			// For now, we assume the installation manager can handle SignatureURL.
			// We can provide a special URL or handle it in the backend.
			if f.ASC != "" {
				asset.Signature = f.ASC
			}

			vi.Assets = append(vi.Assets, asset)
		}

		if len(vi.Assets) > 0 {
			versions = append(versions, vi)
		}
	}

	return versions, nil
}

func mapPlatform(os, arch string) (string, string) {
	// Map Julia OS/Arch names to standard ones
	// Julia OS: mac, linux, win, freebsd
	// Julia Arch: x86_64, aarch64, i686, armv7l

	var resOS, resArch string
	switch os {
	case "mac":
		resOS = "darwin"
	case "linux":
		resOS = "linux"
	case "win":
		resOS = "windows"
	case "freebsd":
		resOS = "freebsd"
	}

	switch arch {
	case "x86_64", "x64":
		resArch = "amd64"
	case "aarch64", "arm64":
		resArch = "arm64"
	case "i686", "x86":
		resArch = "386"
	}

	return resOS, resArch
}
