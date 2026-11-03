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

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	gradleCache  sync.Map
	gradleFlight singleflight.Group
	mavenCache   sync.Map
	mavenFlight  singleflight.Group
)

// ClearGradleCache clears in-memory cache for Gradle releases.
func ClearGradleCache() {
	gradleCache.Range(func(key, value any) bool {
		gradleCache.Delete(key)
		return true
	})
}

// ClearMavenCache clears in-memory cache for Maven releases.
func ClearMavenCache() {
	mavenCache.Range(func(key, value any) bool {
		mavenCache.Delete(key)
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

type mavenMetadata struct {
	Versioning struct {
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
		Latest  string `xml:"latest"`
		Release string `xml:"release"`
	} `xml:"versioning"`
}

func (h *MavenHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	cacheKey := "maven_versions_all"
	if val, ok := mavenCache.Load(cacheKey); ok {
		if cached, ok := val.([]VersionInfo); ok {
			cp := make([]VersionInfo, len(cached))
			copy(cp, cached)
			return cp, nil
		}
	}

	res, err, _ := mavenFlight.Do(cacheKey, func() (interface{}, error) {
		metadataURL := "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/maven-metadata.xml"
		if baseURL != "" && !strings.Contains(baseURL, "archive.apache.org") {
			if strings.HasSuffix(baseURL, ".xml") {
				metadataURL = baseURL
			} else {
				metadataURL = strings.TrimSuffix(baseURL, "/") + "/maven-metadata.xml"
			}
		}

		client := pkgHttp.NewClientWithTimeout(15 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("maven metadata: returned status %d", resp.StatusCode)
		}

		var meta mavenMetadata
		if err := xml.NewDecoder(resp.Body).Decode(&meta); err != nil {
			return nil, err
		}

		commonArches := []string{"x86_64", "amd64", "arm64", "aarch64"}
		var versions []VersionInfo
		rawVersions := meta.Versioning.Versions.Version

		// Reverse iteration to show newer versions first
		for i := len(rawVersions) - 1; i >= 0; i-- {
			v := strings.TrimSpace(rawVersions[i])
			if v == "" {
				continue
			}

			var assets []Asset
			for _, osName := range []string{"linux", "darwin", "windows"} {
				for _, archName := range commonArches {
					assets = append(assets, Asset{
						OS:   osName,
						Arch: archName,
						URL:  fmt.Sprintf("https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/%s/apache-maven-%s-bin.tar.gz", v, v),
					})
				}
			}

			versions = append(versions, VersionInfo{
				Version: v,
				Assets:  assets,
			})
		}

		if len(versions) > 0 {
			mavenCache.Store(cacheKey, versions)
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

func (h *MavenHandler) IsMatch(filename, os, arch string) bool {
	return strings.Contains(filename, "-bin.tar.gz") || strings.Contains(filename, "-bin.zip")
}
