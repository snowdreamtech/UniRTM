// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// HaskellHandler handles GHC downloads via downloads.haskell.org.
type HaskellHandler struct{}

func (h *HaskellHandler) Name() string {
	return "haskell"
}

func (h *HaskellHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Fetch the main page to get versions
	client := pkgHttp.NewClientWithTimeout(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://downloads.haskell.org/~ghc/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch haskell versions: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Regex to extract version numbers like href="9.8.4/"
	re := regexp.MustCompile(`href="([0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?)/"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	platform := env.RuntimeGOOS
	arch := env.RuntimeGOARCH

	// Map generic os/arch to GHC specific naming
	// e.g. x86_64-apple-darwin, aarch64-apple-darwin
	// x86_64-deb10-linux, aarch64-deb10-linux
	// x86_64-unknown-mingw32

	ghcArch := "x86_64"
	if arch == "arm64" {
		ghcArch = "aarch64"
	} else if arch == "386" {
		ghcArch = "i386"
	}

	ghcOS := "unknown-linux"
	ext := "tar.xz"
	if platform == "darwin" {
		ghcOS = "apple-darwin"
	} else if platform == "windows" {
		ghcOS = "unknown-mingw32"
	} else if platform == "linux" {
		// Use a generic widely compatible glibc release, deb10 or deb11 usually works well
		ghcOS = "deb10-linux"
	}

	seen := make(map[string]bool)
	var versions []VersionInfo

	for _, m := range matches {
		ver := m[1]
		if seen[ver] {
			continue
		}
		seen[ver] = true

		filename := fmt.Sprintf("ghc-%s-%s-%s.%s", ver, ghcArch, ghcOS, ext)
		url := fmt.Sprintf("https://downloads.haskell.org/~ghc/%s/%s", ver, filename)

		versions = append(versions, VersionInfo{
			Version: ver,
			Assets: []Asset{
				{
					Filename: filename,
					URL:      url,
					OS:       platform,
					Arch:     arch,
				},
			},
		})
	}

	return versions, nil
}
