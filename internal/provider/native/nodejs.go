// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/sysinfo"
	"golang.org/x/sync/singleflight"
)

// NodeJSHandler handles the official Node.js download metadata from nodejs.org/dist/index.json.
type NodeJSHandler struct {
	flight singleflight.Group
	cache  sync.Map // cacheKey -> []VersionInfo
}

type nodeVersion struct {
	Version string      `json:"version"`
	Date    string      `json:"date"`
	Files   []string    `json:"files"`
	Lts     interface{} `json:"lts"` // can be false or a string (the LTS name)
}

func (h *NodeJSHandler) Name() string {
	return "nodejs"
}

func (h *NodeJSHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, fmt.Errorf("net/http: nil Context")
	}

	// Canonical base URL is always nodejs.org/dist; mirror is only used to
	// fetch the metadata index so it never leaks into lockfile asset URLs.
	if baseURL == "" {
		baseURL = "https://nodejs.org/dist"
	}
	canonicalBase := baseURL

	// Determine metadata-fetch URL (may be a mirror, must not appear in lockfile)
	metaURL := canonicalBase
	if mirrorURL := env.Get("MISE_NODE_MIRROR_URL"); mirrorURL != "" {
		metaURL = mirrorURL
	} else if mirrorURL := env.Get("NODEJS_ORG_MIRROR"); mirrorURL != "" {
		metaURL = mirrorURL
	}

	flavor := env.Get("MISE_NODE_FLAVOR")
	if flavor == "" && env.RuntimeGOOS == "linux" && sysinfo.IsMusl() {
		flavor = "musl"
	}

	// Cache key is always based on the canonical URL so mirror changes don't
	// produce duplicate cache entries with mirror-specific download URLs.
	cacheKey := fmt.Sprintf("%s|%s", canonicalBase, flavor)
	if val, ok := h.cache.Load(cacheKey); ok {
		return val.([]VersionInfo), nil
	}

	var diskVersions []VersionInfo
	if readNativeDiskCache("node", cacheKey, &diskVersions, 10*time.Minute) {
		h.cache.Store(cacheKey, diskVersions)
		return diskVersions, nil
	}

	res, err, _ := h.flight.Do(cacheKey, func() (interface{}, error) {
		if val, ok := h.cache.Load(cacheKey); ok {
			return val, nil
		}

		// Fetch the index from the mirror (if configured) but build asset URLs
		// from the canonical base so the lockfile is never polluted.
		indexURL := fmt.Sprintf("%s/index.json", strings.TrimSuffix(metaURL, "/"))
		client := pkgHttp.NewClientWithTimeout(30 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", indexURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("nodejs: fetch metadata: %w", err)
		}
		defer resp.Body.Close()

		var nv []nodeVersion
		if err := json.NewDecoder(resp.Body).Decode(&nv); err != nil {
			return nil, fmt.Errorf("nodejs: decode metadata: %w", err)
		}

		var versions []VersionInfo
		for _, v := range nv {
			vi := VersionInfo{
				Version: strings.TrimPrefix(v.Version, "v"),
			}

			if s, ok := v.Lts.(string); ok {
				vi.IsLTS = true
				vi.LTSName = s
			}

			for _, f := range v.Files {
				osName, archName, rawArch, ext, isSupported := parseNodeFile(f)
				if !isSupported {
					continue
				}

				// Always use the canonical base for asset URLs so the lockfile
				// is never polluted with mirror-specific hostnames.
				downloadURL := fmt.Sprintf("%s/%s/node-%s-%s-%s%s", strings.TrimSuffix(canonicalBase, "/"), v.Version, v.Version, osName, rawArch, ext)
				if flavor == "musl" {
					// unofficial-builds naming convention: node-vX.Y.Z-linux-ARCH-musl.tar.gz
					downloadURL = fmt.Sprintf("%s/%s/node-%s-%s-%s-musl%s", strings.TrimSuffix(canonicalBase, "/"), v.Version, v.Version, osName, rawArch, ext)
				}

				vi.Assets = append(vi.Assets, Asset{
					URL:          downloadURL,
					Filename:     filepath.Base(downloadURL),
					OS:           osName,
					Arch:         archName,
					Algo:         "sha256",
					SignatureURL: fmt.Sprintf("%s/%s/SHASUMS256.txt.asc", strings.TrimSuffix(canonicalBase, "/"), v.Version),
					Metadata: map[string]string{
						"flavor": flavor,
					},
				})
			}

			if len(vi.Assets) > 0 {
				versions = append(versions, vi)
			}
		}

		h.cache.Store(cacheKey, versions)
		writeNativeDiskCache("node", cacheKey, versions)
		return versions, nil
	})

	if err != nil {
		return nil, err
	}

	return res.([]VersionInfo), nil
}

func parseNodeFile(f string) (string, string, string, string, bool) {
	// Node files format: os-arch (e.g., linux-x64, osx-arm64, win-x64-zip)
	parts := strings.Split(f, "-")
	if len(parts) < 2 {
		return "", "", "", "", false
	}

	osName := parts[0]
	rawArch := parts[1]
	ext := ".tar.gz"

	if len(parts) >= 3 {
		format := parts[2]
		switch format {
		case "zip":
			ext = ".zip"
		case "7z":
			ext = ".7z"
		case "msi":
			ext = ".msi"
		case "pkg":
			ext = ".pkg"
		case "exe":
			ext = ".exe"
		case "tar":
			ext = ".tar.gz"
		}
	}

	// Map osx to darwin
	if osName == "osx" {
		osName = "darwin"
	}

	// Map architecture names to UniRTM standards
	archName := rawArch
	switch rawArch {
	case "x64":
		archName = "amd64"
	case "x86":
		archName = "386"
	}

	// Skip non-tar.gz and non-zip formats for now
	if ext == ".7z" || ext == ".msi" || ext == ".pkg" || ext == ".exe" {
		return "", "", "", "", false
	}

	return osName, archName, rawArch, ext, true
}
