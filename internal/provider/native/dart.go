// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"encoding/xml"
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
	dartCache  sync.Map
	dartFlight singleflight.Group
)

// ClearDartCache clears in-memory cache for Dart SDK versions.
func ClearDartCache() {
	dartCache.Range(func(key, value any) bool {
		dartCache.Delete(key)
		return true
	})
}

// DartHandler handles Dart SDK downloads via Google Storage.
type DartHandler struct{}

func (h *DartHandler) Name() string {
	return "dart"
}

func (h *DartHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	platform := env.RuntimeGOOS
	if platform == "darwin" {
		platform = "macos"
	}

	arch := env.RuntimeGOARCH
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "386" {
		arch = "ia32"
	}

	cacheKey := fmt.Sprintf("%s|%s", platform, arch)

	if val, ok := dartCache.Load(cacheKey); ok {
		if cached, ok := val.([]VersionInfo); ok {
			cp := make([]VersionInfo, len(cached))
			copy(cp, cached)
			return cp, nil
		}
	}

	res, err, _ := dartFlight.Do(cacheKey, func() (interface{}, error) {
		var versions []VersionInfo

		// Fetch the full XML bucket list to get all versions
		listURL := "https://storage.googleapis.com/dart-archive/?prefix=channels/stable/release/&delimiter=/"
		client := pkgHttp.NewClientWithTimeout(10 * time.Second)

		req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch dart versions: HTTP %d", resp.StatusCode)
		}

		// Simple XML parsing to find CommonPrefixes -> Prefix
		type Bucket struct {
			Prefixes []struct {
				Prefix string `xml:"Prefix"`
			} `xml:"CommonPrefixes"`
		}

		var bucket Bucket
		if err := xml.NewDecoder(resp.Body).Decode(&bucket); err != nil {
			return nil, err
		}


	platform := env.RuntimeGOOS
	if platform == "darwin" {
		platform = "macos"
	}

	arch := env.RuntimeGOARCH
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "386" {
		arch = "ia32"
	}
	ext := "zip"

	for _, p := range bucket.Prefixes {
		// p.Prefix looks like "channels/stable/release/3.4.0/"
		parts := strings.Split(p.Prefix, "/")
		if len(parts) >= 4 {
			ver := parts[3]
			if ver == "latest" || len(ver) < 5 {
				// Ignore non-semver like 3.0.0 (length is 5) or 30036
				if !strings.Contains(ver, ".") {
					continue
				}
			}

			versions = append(versions, VersionInfo{
				Version: ver,
				Assets: []Asset{
					{
						Filename: fmt.Sprintf("dartsdk-%s-%s-release.%s", platform, arch, ext),
						URL:      fmt.Sprintf("https://storage.googleapis.com/dart-archive/channels/stable/release/%s/sdk/dartsdk-%s-%s-release.%s", ver, platform, arch, ext),
						OS:       env.RuntimeGOOS,
						Arch:     env.RuntimeGOARCH,
					},
				},
			})
		}
	}

	// Add an alias for latest using the separate JSON file
	reqLatest, _ := http.NewRequestWithContext(ctx, "GET", "https://storage.googleapis.com/dart-archive/channels/stable/release/latest/VERSION", nil)
	if respLatest, err := client.Do(reqLatest); err == nil {
		defer respLatest.Body.Close()
		var verData struct {
			Version string `json:"version"`
		}
		if err := json.NewDecoder(respLatest.Body).Decode(&verData); err == nil {
			versions = append(versions, VersionInfo{
				Version: "latest",
				Assets: []Asset{
					{
						Filename: fmt.Sprintf("dartsdk-%s-%s-release.%s", platform, arch, ext),
						URL:      fmt.Sprintf("https://storage.googleapis.com/dart-archive/channels/stable/release/%s/sdk/dartsdk-%s-%s-release.%s", verData.Version, platform, arch, ext),
						OS:       env.RuntimeGOOS,
						Arch:     env.RuntimeGOARCH,
					},
				},
			})
		}
	}

		if len(versions) > 0 {
			dartCache.Store(cacheKey, versions)
		}
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

