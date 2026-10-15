// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
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
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
		Package      struct {
			Name     string `json:"name"`
			Link     string `json:"link"`
			Checksum string `json:"checksum"`
		} `json:"package"`
		SignatureLink string `json:"signature_link"`
	} `json:"binaries"`
	ReleaseName string `json:"release_name"`
	VersionData struct {
		OpenjdkVersion string `json:"openjdk_version"`
	} `json:"version_data"`
}

func mapAdoptiumOS(adoptOS string) string {
	switch adoptOS {
	case "mac":
		return "darwin"
	case "windows":
		return "windows"
	case "linux", "alpine-linux":
		return "linux"
	case "aix":
		return "aix"
	case "solaris":
		return "solaris"
	default:
		return adoptOS
	}
}

func mapAdoptiumArch(adoptArch string) string {
	switch adoptArch {
	case "x64":
		return "amd64"
	case "aarch64":
		return "arm64"
	case "x32", "x86":
		return "386"
	case "arm":
		return "arm"
	case "ppc64le":
		return "ppc64le"
	case "ppc64":
		return "ppc64"
	case "s390x":
		return "s390x"
	case "riscv64":
		return "riscv64"
	default:
		return adoptArch
	}
}

type adoptiumAvailableReleasesResponse struct {
	AvailableReleases []int `json:"available_releases"`
}

func fetchAdoptiumMajorVersions(ctx context.Context, client *http.Client, apiBase string) []string {
	req, err := http.NewRequestWithContext(ctx, "GET", apiBase+"/v3/info/available_releases", nil)
	if err != nil {
		return []string{"25", "23", "21", "17", "11", "8"}
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := client.Do(req)
	if err != nil {
		return []string{"25", "23", "21", "17", "11", "8"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{"25", "23", "21", "17", "11", "8"}
	}

	var data adoptiumAvailableReleasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data.AvailableReleases) == 0 {
		return []string{"25", "23", "21", "17", "11", "8"}
	}

	// Sort descending so newer releases come first
	sort.Slice(data.AvailableReleases, func(i, j int) bool {
		return data.AvailableReleases[i] > data.AvailableReleases[j]
	})

	res := make([]string, len(data.AvailableReleases))
	for i, v := range data.AvailableReleases {
		res[i] = strconv.Itoa(v)
	}
	return res
}

func (h *JavaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
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

	cacheKey := fmt.Sprintf("%s|%s", apiBase, imageType)
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

		majorVersions := fetchAdoptiumMajorVersions(ctx, client, apiBase)

		results := make([][]VersionInfo, len(majorVersions))
		var wg sync.WaitGroup

		for idx, v := range majorVersions {
			wg.Add(1)
			go func(i int, major string) {
				defer wg.Done()
				url := fmt.Sprintf("%s/v3/assets/feature_releases/%s/ga?heap_size=normal&image_type=%s&jvm_impl=hotspot&project=jdk&vendor=eclipse&page_size=50", apiBase, major, imageType)

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

					var assets []Asset
					for _, bin := range rel.Binaries {
						if bin.Package.Link == "" {
							continue
						}
						targetOS := mapAdoptiumOS(bin.OS)
						if targetOS == "" {
							targetOS = env.RuntimeGOOS
						}
						targetArch := mapAdoptiumArch(bin.Architecture)
						if targetArch == "" {
							targetArch = runtime.GOARCH
						}

						asset := Asset{
							Filename:     bin.Package.Name,
							URL:          bin.Package.Link,
							SignatureURL: bin.SignatureLink,
							OS:           targetOS,
							Arch:         targetArch,
							Metadata:     make(map[string]string),
						}
						if bin.Package.Checksum != "" {
							asset.Checksum = bin.Package.Checksum
							asset.Algo = "sha256"
						}
						assets = append(assets, asset)
					}

					if len(assets) > 0 {
						subVersions = append(subVersions, VersionInfo{
							Version: version,
							Assets:  assets,
						})
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
