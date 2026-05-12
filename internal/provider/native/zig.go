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

// ZigHandler handles Zig tool versions via its official JSON API.
type ZigHandler struct{}

func (h *ZigHandler) Name() string {
	return "zig"
}

type zigVersionMap map[string]zigVersion

type zigVersion struct {
	Version string                 `json:"version"` // Only present in some contexts
	Tarball string                 `json:"tarball"` // Only present in platform-specific map
	Size    string                 `json:"size"`
	Hash    string                 `json:"hash"`
	Master  map[string]interface{} `json:"master"` // We might handle master/nightly later
}

func (h *ZigHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if baseURL == "" {
		baseURL = "https://ziglang.org/download/index.json"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zig api: returned status %d", resp.StatusCode)
	}

	var data map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for verStr, platformMap := range data {
		if verStr == "master" {
			// Skip master/nightly for now to keep it stable
			continue
		}

		var assets []Asset
		for platKey, platData := range platformMap {
			// platKey format: "x86_64-linux", "aarch64-macos", etc.
			os, arch := h.parsePlatform(platKey)
			if os == "" || arch == "" {
				continue
			}

			// platData is a map containing "tarball", "shasum", etc.
			m, ok := platData.(map[string]interface{})
			if !ok {
				continue
			}

			url, _ := m["tarball"].(string)
			hash, _ := m["shasum"].(string)

			if url == "" {
				continue
			}

			assets = append(assets, Asset{
				Filename: filepathBase(url),
				URL:      url,
				OS:       os,
				Arch:     arch,
				Checksum: hash,
			})
		}

		if len(assets) > 0 {
			versions = append(versions, VersionInfo{
				Version: verStr,
				Assets:  assets,
			})
		}
	}

	return versions, nil
}

func (h *ZigHandler) parsePlatform(platKey string) (string, string) {
	parts := strings.Split(platKey, "-")
	if len(parts) != 2 {
		return "", ""
	}

	archRaw := parts[0]
	osRaw := parts[1]

	var os, arch string

	// OS Mapping
	switch osRaw {
	case "linux":
		os = "linux"
	case "macos":
		os = "darwin"
	case "windows":
		os = "windows"
	case "freebsd":
		os = "freebsd"
	}

	// Arch Mapping
	switch archRaw {
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	case "x86":
		arch = "386"
	case "riscv64":
		arch = "riscv64"
	}

	return os, arch
}

func filepathBase(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
