// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	phpHandlerCache  sync.Map
	phpHandlerFlight singleflight.Group
)

// ClearPhpHandlerCache clears the in-memory cache of PHP releases.
func ClearPhpHandlerCache() {
	phpHandlerCache.Range(func(key, value any) bool {
		phpHandlerCache.Delete(key)
		return true
	})
}

// PhpHandler handles PHP distributions using StaticPHP (standalone static binaries
// for Linux and macOS) and official Windows binaries from windows.php.net.
type PhpHandler struct {
}

func (h *PhpHandler) Name() string {
	return "php"
}

type staticPHPFileItem struct {
	IsDir    bool   `json:"is_dir"`
	FullPath string `json:"full_path"`
	Name     string `json:"name"`
}

var staticPHPRegex = regexp.MustCompile(`^php-([0-9.]+)-cli-([a-z0-9_]+)-([a-z0-9_]+)\.tar\.gz$`)

// phpWinCompiler returns the MSVC compiler tag used by windows.php.net for a given PHP version.
func phpWinCompiler(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		major := parts[0]
		minor := parts[1]
		if major == "8" {
			minNum, err := strconv.Atoi(minor)
			if err == nil && minNum >= 3 {
				return "vs17"
			}
			return "vs16"
		}
		if major == "7" {
			return "vc15"
		}
	}
	return "vs17"
}

// ResolveVersions fetches and parses available PHP versions.
func (h *PhpHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	if baseURL == "" {
		baseURL = "https://dl.static-php.dev/static-php-cli/bulk/?format=json"
	}

	cacheKey := "php:versions:" + baseURL
	if cached, ok := phpHandlerCache.Load(cacheKey); ok {
		return cached.([]VersionInfo), nil
	}

	res, err, _ := phpHandlerFlight.Do(cacheKey, func() (interface{}, error) {
		client := pkgHttp.NewClientWithTimeout(30 * time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch php versions: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch php versions returned status %d", resp.StatusCode)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read php versions body: %w", err)
		}

		var items []staticPHPFileItem
		if err := json.Unmarshal(bodyBytes, &items); err != nil {
			return nil, fmt.Errorf("unmarshal php releases json: %w", err)
		}

		// Group static-php assets by version
		versionAssets := make(map[string][]Asset)
		for _, item := range items {
			if item.IsDir {
				continue
			}
			matches := staticPHPRegex.FindStringSubmatch(item.Name)
			if len(matches) != 4 {
				continue
			}

			ver := matches[1]
			osRaw := matches[2]
			archRaw := matches[3]

			var osName, archName string
			switch osRaw {
			case "linux":
				osName = "linux"
			case "macos":
				osName = "darwin"
			default:
				continue
			}

			switch archRaw {
			case "x86_64":
				archName = "amd64"
			case "aarch64":
				archName = "arm64"
			default:
				continue
			}

			url := "https://dl.static-php.dev" + item.FullPath
			if strings.HasPrefix(item.FullPath, "http://") || strings.HasPrefix(item.FullPath, "https://") {
				url = item.FullPath
			}

			versionAssets[ver] = append(versionAssets[ver], Asset{
				URL:      url,
				Filename: item.Name,
				OS:       osName,
				Arch:     archName,
				Metadata: map[string]string{
					"static": "true",
				},
			})
		}

		// Also attach Windows assets for each version
		for ver, assets := range versionAssets {
			compiler := phpWinCompiler(ver)
			winFilenameX64 := fmt.Sprintf("php-%s-Win32-%s-x64.zip", ver, compiler)
			winURLX64 := fmt.Sprintf("https://windows.php.net/downloads/releases/archives/%s", winFilenameX64)
			assets = append(assets, Asset{
				URL:      winURLX64,
				Filename: winFilenameX64,
				OS:       "windows",
				Arch:     "amd64",
				Metadata: map[string]string{},
			})

			winFilenameX86 := fmt.Sprintf("php-%s-Win32-%s-x86.zip", ver, compiler)
			winURLX86 := fmt.Sprintf("https://windows.php.net/downloads/releases/archives/%s", winFilenameX86)
			assets = append(assets, Asset{
				URL:      winURLX86,
				Filename: winFilenameX86,
				OS:       "windows",
				Arch:     "386",
				Metadata: map[string]string{},
			})

			versionAssets[ver] = assets
		}

		var result []VersionInfo
		for ver, assets := range versionAssets {
			result = append(result, VersionInfo{
				Version: ver,
				Assets:  assets,
			})
		}

		// Sort semver descending
		sort.Slice(result, func(i, j int) bool {
			return compareSemver(result[i].Version, result[j].Version) > 0
		})

		phpHandlerCache.Store(cacheKey, result)
		return result, nil
	})

	if err != nil {
		return nil, err
	}
	return res.([]VersionInfo), nil
}

// ResolveTag resolves a specific version of PHP.
func (h *PhpHandler) ResolveTag(ctx context.Context, version string) (*VersionInfo, error) {
	cleanVer := strings.TrimPrefix(version, "php-")
	cleanVer = strings.TrimPrefix(cleanVer, "v")

	// Try resolving all versions first
	versions, err := h.ResolveVersions(ctx, "")
	if err == nil {
		for _, v := range versions {
			if v.Version == cleanVer {
				return &v, nil
			}
		}
	}

	// If exact version wasn't in static-php-cli list, construct fallback assets:
	// 1. Windows official archive zip
	compiler := phpWinCompiler(cleanVer)
	winFilenameX64 := fmt.Sprintf("php-%s-Win32-%s-x64.zip", cleanVer, compiler)
	winURLX64 := fmt.Sprintf("https://windows.php.net/downloads/releases/archives/%s", winFilenameX64)
	winFilenameX86 := fmt.Sprintf("php-%s-Win32-%s-x86.zip", cleanVer, compiler)
	winURLX86 := fmt.Sprintf("https://windows.php.net/downloads/releases/archives/%s", winFilenameX86)

	assets := []Asset{
		{
			URL:      winURLX64,
			Filename: winFilenameX64,
			OS:       "windows",
			Arch:     "amd64",
			Metadata: map[string]string{},
		},
		{
			URL:      winURLX86,
			Filename: winFilenameX86,
			OS:       "windows",
			Arch:     "386",
			Metadata: map[string]string{},
		},
	}

	// 2. Check if a nearby patch in the same minor version exists for static Linux/macOS
	if len(versions) > 0 {
		cleanParts := strings.Split(cleanVer, ".")
		if len(cleanParts) >= 2 {
			majorMinor := cleanParts[0] + "." + cleanParts[1]
			for _, v := range versions {
				if strings.HasPrefix(v.Version, majorMinor+".") {
					for _, a := range v.Assets {
						if a.OS == "linux" || a.OS == "darwin" {
							assets = append(assets, a)
						}
					}
					break
				}
			}
		}
	}

	return &VersionInfo{
		Version: cleanVer,
		Assets:  assets,
	}, nil
}

func compareSemver(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		num1 := 0
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		num2 := 0
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}
		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}
	return 0
}
