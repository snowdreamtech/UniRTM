// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/sysinfo"
	"golang.org/x/sync/singleflight"
)

var (
	javaHandlerCache  sync.Map
	javaHandlerFlight singleflight.Group
)

// ClearJavaHandlerCache clears the in-memory cache for Adoptium Java releases.
func ClearJavaHandlerCache() {
	javaHandlerCache.Range(func(key, value any) bool {
		javaHandlerCache.Delete(key)
		return true
	})
}

// JavaHandler handles Java distributions via Adoptium (Temurin) API.
type JavaHandler struct {
	ImageType string // "jdk" or "jre"
}

func (h *JavaHandler) Name() string {
	return "java"
}

type adoptiumRelease struct {
	Binaries []struct {
		Package struct {
			Name string `json:"name"`
			Link string `json:"link"`
		} `json:"package"`
		SignatureLink string `json:"signature_link"`
	} `json:"binaries"`
	ReleaseName string `json:"release_name"`
	VersionData struct {
		OpenjdkVersion string `json:"openjdk_version"`
	} `json:"version_data"`
}

func (h *JavaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Adoptium versions: 23 (GA), 21 (LTS), 17 (LTS), 11 (LTS), 8 (LTS)
	majorVersions := []string{"23", "21", "17", "11", "8"}

	// Map OS/Arch to Adoptium values
	os := env.RuntimeGOOS
	if os == "darwin" {
		os = "mac"
	} else if os == "linux" && sysinfo.IsMusl() {
		os = "alpine-linux"
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "arm64" {
		arch = "aarch64"
	}

	imageType := h.ImageType
	if imageType == "" {
		imageType = "jdk"
	}

	apiBase := env.Get("ADOPTIUM_API_BASEURL")
	if apiBase == "" {
		apiBase = env.Get("JAVA_MIRROR_URL")
	}
	if apiBase == "" {
		apiBase = "https://api.adoptium.net"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	cacheKey := fmt.Sprintf("%s|%s|%s|%s", apiBase, imageType, os, arch)
	if val, ok := javaHandlerCache.Load(cacheKey); ok {
		if cached, ok := val.([]VersionInfo); ok {
			cp := make([]VersionInfo, len(cached))
			copy(cp, cached)
			return cp, nil
		}
	}

	res, err, _ := javaHandlerFlight.Do(cacheKey, func() (interface{}, error) {
		var allVersions []VersionInfo
		client := pkgHttp.NewClientWithTimeout(30 * time.Second)

		results := make([][]VersionInfo, len(majorVersions))
		var wg sync.WaitGroup

		for idx, v := range majorVersions {
			wg.Add(1)
			go func(i int, major string) {
				defer wg.Done()
				url := fmt.Sprintf("%s/v3/assets/feature_releases/%s/ga?architecture=%s&heap_size=normal&image_type=%s&jvm_impl=hotspot&os=%s&project=jdk&vendor=eclipse", apiBase, major, arch, imageType, os)

				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

				resp, err := client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return
				}

				var releases []adoptiumRelease
				if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
					return
				}

				var subVersions []VersionInfo
				for _, rel := range releases {
					version := rel.VersionData.OpenjdkVersion
					// Clean version string (e.g. 21.0.2+13-LTS -> 21.0.2)
					// Remove -LTS if present
					version = strings.ReplaceAll(version, "-LTS", "")
					if idx := strings.Index(version, "+"); idx != -1 {
						version = version[:idx]
					}

					for _, bin := range rel.Binaries {
						assets := []Asset{
							{
								Filename:     bin.Package.Name,
								URL:          bin.Package.Link,
								SignatureURL: bin.SignatureLink,
								OS:           env.RuntimeGOOS,
								Arch:         runtime.GOARCH,
								Metadata:     make(map[string]string),
							},
						}

						subVersions = append(subVersions, VersionInfo{
							Version: version,
							Assets:  assets,
						})
						// Just take the first binary that matches our query filters
						break
					}
				}
				results[i] = subVersions
			}(idx, v)
		}
		wg.Wait()

		for _, sub := range results {
			allVersions = append(allVersions, sub...)
		}

		if len(allVersions) > 0 {
			javaHandlerCache.Store(cacheKey, allVersions)
		}
		return allVersions, nil
	})

	if err != nil {
		return nil, err
	}
	cached := res.([]VersionInfo)
	cp := make([]VersionInfo, len(cached))
	copy(cp, cached)
	return cp, nil
}

