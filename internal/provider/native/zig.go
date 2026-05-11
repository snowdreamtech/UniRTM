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

// ZigHandler handles Zig distribution via ziglang.org/download/index.json.
type ZigHandler struct{}

func (h *ZigHandler) Name() string {
	return "zig"
}

type zigIndex map[string]struct {
	Tarball string `json:"tarball"`
	Shasum  string `json:"shasum"`
	Size    string `json:"size"`
}

type zigResponse map[string]map[string]interface{}

func (h *ZigHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Zig index URL: https://ziglang.org/download/index.json
	url := "https://ziglang.org/download/index.json"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for version, platformsRaw := range data {
		if version == "master" {
			// We can handle master/nightly later
			continue
		}

		platforms, ok := platformsRaw.(map[string]interface{})
		if !ok {
			continue
		}

		var assets []Asset
		for platformKey, platformDataRaw := range platforms {
			pd, ok := platformDataRaw.(map[string]interface{})
			if !ok {
				continue
			}

			tarball, _ := pd["tarball"].(string)
			shasum, _ := pd["shasum"].(string)

			if tarball == "" {
				continue
			}

			os, arch := h.parsePlatform(platformKey)
			if os == "" || arch == "" {
				continue
			}

			assets = append(assets, Asset{
				Filename: filepathBase(tarball),
				URL:      tarball,
				Checksum: shasum,
				Algo:     "sha256",
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

func (h *ZigHandler) parsePlatform(key string) (string, string) {
	// Zig keys are like "x86_64-linux", "aarch64-macos", "x86-windows"
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return "", ""
	}

	archRaw := parts[0]
	osRaw := parts[1]

	var os, arch string

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

	switch archRaw {
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	case "x86":
		arch = "386"
	case "armv7l":
		arch = "arm"
	}

	return os, arch
}

func filepathBase(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
