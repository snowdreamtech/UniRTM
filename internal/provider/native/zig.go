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

// ZigHandler handles Zig tool versions via its official JSON API.
type ZigHandler struct {
	cache  sync.Map
	flight singleflight.Group
}

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
	// Canonical index URL: always ziglang.org — never a mirror.
	const canonicalIndexURL = "https://ziglang.org/download/index.json"
	const canonicalHost = "https://ziglang.org"

	if baseURL == "" {
		baseURL = canonicalIndexURL
	}

	// Fetch-side URL: may be a mirror, but MUST NOT appear in lockfile entries.
	metaURL := baseURL
	if mirrorURL := env.Get("ZIG_MIRROR_URL"); mirrorURL != "" {
		metaURL = mirrorURL
	}

	// Cache key uses the canonical URL so mirror changes don't produce entries
	// with mirror-specific download URLs.
	cacheKey := canonicalIndexURL
	if val, ok := h.cache.Load(cacheKey); ok {
		return val.([]VersionInfo), nil
	}

	var diskVersions []VersionInfo
	if readNativeDiskCache("zig", cacheKey, &diskVersions, 10*time.Minute) {
		h.cache.Store(cacheKey, diskVersions)
		return diskVersions, nil
	}

	res, err, _ := h.flight.Do(cacheKey, func() (interface{}, error) {
		if val, ok := h.cache.Load(cacheKey); ok {
			return val, nil
		}

		client := pkgHttp.NewClientWithTimeout(30 * time.Second)
		// Fetch index from mirror (if configured) to avoid rate-limiting;
		// asset URLs will be normalized to the canonical host before storage.
		req, err := http.NewRequestWithContext(ctx, "GET", metaURL, nil)
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

				// Normalize tarball URL: if the mirror's JSON encodes mirror-specific
				// hostnames, replace the host part with the canonical ziglang.org so
				// the lockfile is never polluted by transient mirror URLs.
				if metaURL != canonicalIndexURL {
					// Strip the mirror host prefix and replace with canonical host.
					// Mirrors typically serve the same path structure as ziglang.org.
					mirrorBase := strings.TrimSuffix(metaURL, "/")
					// Remove trailing "/download/index.json" or "/index.json" so we
					// get just the host+prefix portion that the mirror substituted.
					for _, suffix := range []string{"/download/index.json", "/index.json"} {
						if strings.HasSuffix(mirrorBase, suffix) {
							mirrorBase = mirrorBase[:len(mirrorBase)-len(suffix)]
							break
						}
					}
					if mirrorBase != "" && strings.HasPrefix(url, mirrorBase) {
						url = canonicalHost + url[len(mirrorBase):]
					}
				}

				assets = append(assets, Asset{
					Filename: filepathBase(url),
					URL:      url,
					OS:       os,
					Arch:     arch,
					Checksum: hash,
					Metadata: make(map[string]string),
				})
			}

			if len(assets) > 0 {
				versions = append(versions, VersionInfo{
					Version: verStr,
					Assets:  assets,
				})
			}
		}

		h.cache.Store(cacheKey, versions)
		writeNativeDiskCache("zig", cacheKey, versions)
		return versions, nil
	})

	if err != nil {
		return nil, err
	}

	return res.([]VersionInfo), nil
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
