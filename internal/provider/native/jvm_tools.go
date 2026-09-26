// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	gradleCache  sync.Map
	gradleFlight singleflight.Group
)

// ClearGradleCache clears in-memory cache for Gradle releases.
func ClearGradleCache() {
	gradleCache.Range(func(key, value any) bool {
		gradleCache.Delete(key)
		return true
	})
}

type GradleHandler struct{}

func (h *GradleHandler) Name() string {
	return "gradle"
}

type gradleVersion struct {
	Version        string `json:"version"`
	DownloadURL    string `json:"downloadUrl"`
	Snapshot       bool   `json:"snapshot"`
	Nightly        bool   `json:"nightly"`
	ReleaseNightly bool   `json:"releaseNightly"`
	Broken         bool   `json:"broken"`
}

func (h *GradleHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	cacheKey := "gradle_versions_all"
	if val, ok := gradleCache.Load(cacheKey); ok {
		if cached, ok := val.([]VersionInfo); ok {
			cp := make([]VersionInfo, len(cached))
			copy(cp, cached)
			return cp, nil
		}
	}

	res, err, _ := gradleFlight.Do(cacheKey, func() (interface{}, error) {
		// Gradle provides a nice JSON API
		req, err := http.NewRequestWithContext(ctx, "GET", "https://services.gradle.org/versions/all", nil)
		if err != nil {
			return nil, err
		}
		resp, err := pkgHttp.NewClient().Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var gv []gradleVersion
		if err := json.NewDecoder(resp.Body).Decode(&gv); err != nil {
			return nil, err
		}

		var res []VersionInfo
		for _, v := range gv {
			// Skip non-release versions and non-stable versions
			lowVersion := strings.ToLower(v.Version)
			if v.Snapshot || v.Nightly || v.ReleaseNightly || v.Broken ||
				strings.Contains(lowVersion, "milestone") ||
				strings.Contains(lowVersion, "rc") {
				continue
			}

			// Create assets for common architectures since Gradle is universal
			commonArches := []string{"x86_64", "amd64", "arm64", "aarch64"}
			var assets []Asset
			for _, osName := range []string{"linux", "darwin", "windows"} {
				for _, archName := range commonArches {
					assets = append(assets, Asset{
						OS:   osName,
						Arch: archName,
						URL:  v.DownloadURL,
					})
				}
			}

			res = append(res, VersionInfo{
				Version: v.Version,
				Assets:  assets,
			})
		}

		if len(res) > 0 {
			gradleCache.Store(cacheKey, res)
		}
		return res, nil
	})

	if err != nil {
		return nil, err
	}
	cached := res.([]VersionInfo)
	cp := make([]VersionInfo, len(cached))
	copy(cp, cached)
	return cp, nil
}


func (h *GradleHandler) IsMatch(filename, os, arch string) bool {
	// For Gradle, we use the same URL for all platforms as it is platform-independent
	return true
}

type MavenHandler struct{}

func (h *MavenHandler) Name() string {
	return "maven"
}

func (h *MavenHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Maven doesn't have a simple JSON API for all versions,
	// but we can use the archive page or a well-known pattern.
	// For simplicity, we'll implement a basic version resolver or hardcode recent ones for now,
	// or scrape the apache archive.

	// Better approach: use the Maven Central metadata for the distribution
	commonArches := []string{"x86_64", "amd64", "arm64", "aarch64"}
	var assets396 []Asset
	for _, osName := range []string{"linux", "darwin", "windows"} {
		for _, archName := range commonArches {
			assets396 = append(assets396, Asset{
				OS:   osName,
				Arch: archName,
				URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.tar.gz",
			})
		}
	}

	return []VersionInfo{
		{
			Version: "3.9.6",
			Assets:  assets396,
		},
	}, nil
}

func (h *MavenHandler) IsMatch(filename, os, arch string) bool {
	return strings.Contains(filename, "-bin.tar.gz") || strings.Contains(filename, "-bin.zip")
}
