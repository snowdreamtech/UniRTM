// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	kubectlCache  sync.Map
	kubectlFlight singleflight.Group
)

// ClearKubectlCache clears in-memory cache for kubectl versions.
func ClearKubectlCache() {
	kubectlCache.Range(func(key, value any) bool {
		kubectlCache.Delete(key)
		return true
	})
}

// KubectlHandler handles kubectl distribution via dl.k8s.io.
type KubectlHandler struct{}

func (h *KubectlHandler) Name() string {
	return "kubectl"
}

type k8sRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

func (h *KubectlHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	cacheKey := "kubectl_all_versions"
	if val, ok := kubectlCache.Load(cacheKey); ok {
		if cached, ok := val.([]VersionInfo); ok {
			cp := make([]VersionInfo, len(cached))
			copy(cp, cached)
			return cp, nil
		}
	}

	res, err, _ := kubectlFlight.Do(cacheKey, func() (interface{}, error) {
		client := pkgHttp.NewClientWithTimeout(15 * time.Second)
		seen := make(map[string]bool)
		var versions []VersionInfo

		// 1. Try fetching full release history from GitHub Releases
		apiBase := env.Get("GITHUB_API_BASEURL")
		if apiBase == "" {
			apiBase = "https://api.github.com"
		}
		apiBase = strings.TrimSuffix(apiBase, "/")
		ghURL := fmt.Sprintf("%s/repos/kubernetes/kubernetes/releases?per_page=100", apiBase)

		req, err := http.NewRequestWithContext(ctx, "GET", ghURL, nil)
		if err == nil {
			if token := getGitHubToken(); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var releases []k8sRelease
				if err := json.NewDecoder(resp.Body).Decode(&releases); err == nil {
					for _, rel := range releases {
						if rel.Draft {
							continue
						}
						tag := strings.TrimSpace(rel.TagName)
						if tag == "" || seen[tag] {
							continue
						}
						// Focus on v1.x.y releases (and standard prereleases if any)
						if strings.HasPrefix(tag, "v") && strings.Contains(tag, ".") {
							seen[tag] = true
							versions = append(versions, VersionInfo{
								Version: tag,
								Assets:  h.generateAssets(tag),
							})
						}
					}
				}
			} else if resp != nil {
				resp.Body.Close()
			}
		}

		// 2. Also fetch latest stable version from dl.k8s.io as guaranteed fallback/addition
		stableURL := "https://dl.k8s.io/release/stable.txt"
		stableReq, err := http.NewRequestWithContext(ctx, "GET", stableURL, nil)
		if err == nil {
			stableResp, err := client.Do(stableReq)
			if err == nil && stableResp.StatusCode == http.StatusOK {
				defer stableResp.Body.Close()
				body, err := io.ReadAll(stableResp.Body)
				if err == nil {
					latestVersion := strings.TrimSpace(string(body))
					if latestVersion != "" && !seen[latestVersion] {
						seen[latestVersion] = true
						// Prepend latest stable version if not already present
						versions = append([]VersionInfo{
							{
								Version: latestVersion,
								Assets:  h.generateAssets(latestVersion),
							},
						}, versions...)
					}
				}
			} else if stableResp != nil {
				stableResp.Body.Close()
			}
		}

		if len(versions) == 0 {
			return nil, fmt.Errorf("kubectl api: failed to resolve versions from both GitHub releases and stable.txt")
		}

		kubectlCache.Store(cacheKey, versions)
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

func (h *KubectlHandler) generateAssets(version string) []Asset {
	var assets []Asset

	// Supported OS/Arch combinations
	platforms := []struct{ os, arch string }{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	for _, p := range platforms {
		ext := ""
		if p.os == "windows" {
			ext = ".exe"
		}

		url := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl%s", version, p.os, p.arch, ext)

		assets = append(assets, Asset{
			Filename: "kubectl" + ext,
			URL:      url,
			OS:       p.os,
			Arch:     p.arch,
		})
	}

	return assets
}
