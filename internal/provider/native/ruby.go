// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	rubyHandlerCache  sync.Map
	rubyHandlerFlight singleflight.Group
)

// ClearRubyHandlerCache clears the in-memory cache for Ruby releases.
func ClearRubyHandlerCache() {
	rubyHandlerCache.Range(func(key, value any) bool {
		rubyHandlerCache.Delete(key)
		return true
	})
}

// RubyHandler handles Ruby distributions via ruby/ruby-builder.
// It targets precompiled binaries used in GitHub Actions.
type RubyHandler struct {
	GithubHandler
}

func (h *RubyHandler) Name() string {
	return "ruby"
}

func (h *RubyHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Re-use GithubHandler's logic but with custom platform detection
	h.GithubHandler.Owner = "ruby"
	h.GithubHandler.Repo = "ruby-builder"

	releases, err := h.GithubHandler.ResolveVersions(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	var filtered []VersionInfo
	for _, rel := range releases {
		// Only consider MRI Ruby releases (skip jruby, truffleruby, etc.)
		if !strings.HasPrefix(rel.Version, "ruby-") && !strings.HasPrefix(rel.Version, "v") && !strings.Contains(rel.Version, ".") {
			continue
		}
		cleanVer := strings.TrimPrefix(rel.Version, "ruby-")
		cleanVer = strings.TrimPrefix(cleanVer, "v")

		var assets []Asset
		for _, a := range rel.Assets {
			filename := strings.ToLower(a.Filename)
			if strings.HasPrefix(filename, "ruby-") {
				assets = append(assets, a)
			}
		}
		if len(assets) > 0 {
			filtered = append(filtered, VersionInfo{
				Version: cleanVer,
				Assets:  assets,
			})
		}
	}

	return filtered, nil
}

// ResolveTag fetches a specific Ruby release by version tag (e.g. "ruby-3.4.1").
func (h *RubyHandler) ResolveTag(ctx context.Context, version string) (*VersionInfo, error) {
	cleanVer := strings.TrimPrefix(version, "ruby-")
	cleanVer = strings.TrimPrefix(cleanVer, "v")

	cacheKey := "tag:" + cleanVer
	if val, ok := rubyHandlerCache.Load(cacheKey); ok {
		return val.(*VersionInfo), nil
	}

	res, err, _ := rubyHandlerFlight.Do(cacheKey, func() (interface{}, error) {
		h.GithubHandler.Owner = "ruby"
		h.GithubHandler.Repo = "ruby-builder"

		// 1. Try GithubHandler.ResolveTag first
		info, err := h.GithubHandler.ResolveTag(ctx, "ruby-"+cleanVer)
		if err == nil && info != nil && len(info.Assets) > 0 {
			info.Version = cleanVer
			rubyHandlerCache.Store(cacheKey, info)
			return info, nil
		}

		// 2. Fallback to scraping expanded_assets if GitHub API failed (e.g. rate limit 403)
		expandedURL := fmt.Sprintf("https://github.com/ruby/ruby-builder/releases/expanded_assets/ruby-%s", cleanVer)
		client := pkgHttp.NewClientWithTimeout(30 * time.Second)
		req, reqErr := http.NewRequestWithContext(ctx, "GET", expandedURL, nil)
		if reqErr == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0")
			resp, doErr := client.Do(req)
			if doErr == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				html := string(bodyBytes)

				re := regexp.MustCompile(`/ruby/ruby-builder/releases/download/[^"'\s<>]+`)
				matches := re.FindAllString(html, -1)
				if len(matches) > 0 {
					var assets []Asset
					seen := make(map[string]bool)
					for _, m := range matches {
						parts := strings.Split(m, "/")
						filename := parts[len(parts)-1]
						if seen[filename] {
							continue
						}
						seen[filename] = true

						os, arch := h.detectPlatform(filename)
						if os == "" || arch == "" {
							continue
						}
						assets = append(assets, Asset{
							Filename: filename,
							URL:      "https://github.com" + m,
							OS:       os,
							Arch:     arch,
							Metadata: make(map[string]string),
						})
					}
					if len(assets) > 0 {
						fallbackInfo := &VersionInfo{
							Version: cleanVer,
							Assets:  assets,
						}
						rubyHandlerCache.Store(cacheKey, fallbackInfo)
						return fallbackInfo, nil
					}
				}
			} else if resp != nil {
				_ = resp.Body.Close()
			}
		}

		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("ruby: version %s not found", version)
	})

	if err != nil {
		return nil, err
	}
	return res.(*VersionInfo), nil
}

func (h *RubyHandler) isMatch(filename string) bool {
	filename = strings.ToLower(filename)
	if !strings.HasPrefix(filename, "ruby-") {
		return false
	}

	os := env.RuntimeGOOS
	arch := env.RuntimeGOARCH

	// Arch check: ruby-builder binaries are mostly x64 (amd64)
	if arch != "amd64" && !strings.Contains(filename, arch) {
		// If it's arm64, look for arm64 in filename
		if arch == "arm64" && !strings.Contains(filename, "arm64") {
			return false
		}
	} else if arch == "amd64" && !strings.Contains(filename, "x86_64") && !strings.Contains(filename, "amd64") {
		return false
	}

	if os == "darwin" {
		return strings.Contains(filename, "macos")
	}

	if os == "linux" {
		// ruby-builder is very Ubuntu-centric.
		// We'll try to match ubuntu, but on other distros this might be risky.
		return strings.Contains(filename, "ubuntu")
	}

	if os == "windows" {
		return strings.Contains(filename, "windows")
	}

	return false
}
