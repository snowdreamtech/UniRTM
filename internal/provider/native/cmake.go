// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"strings"
)

// CMakeHandler handles CMake build tool versions via GitHub releases.
type CMakeHandler struct {
	GithubHandler
}

func (h *CMakeHandler) Name() string {
	return "cmake"
}

func (h *CMakeHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "Kitware"
	h.Repo = "CMake"

	versions, err := h.GithubHandler.ResolveVersions(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	// CMake specific filtering:
	// - Exclude source packages (.tar.gz without platform, .zip without platform)
	// - Handle 'macos-universal' as both amd64 and arm64
	var filteredVersions []VersionInfo
	for _, v := range versions {
		var assets []Asset
		for _, a := range v.Assets {
			// Skip source or documentation assets
			if strings.Contains(a.Filename, "-SHA-256") || strings.HasSuffix(a.Filename, ".asc") {
				continue
			}

			if a.OS == "darwin" && (strings.Contains(a.Filename, "macos-universal") || a.Arch == "universal") {
				// MacOS universal package supports both
				amd64Asset := a
				amd64Asset.Arch = "amd64"
				assets = append(assets, amd64Asset)

				// Also add as arm64
				arm64Asset := a
				arm64Asset.Arch = "arm64"
				assets = append(assets, arm64Asset)
			} else if a.OS != "" && a.Arch != "" {
				assets = append(assets, a)
			}
		}

		if len(assets) > 0 {
			filteredVersions = append(filteredVersions, VersionInfo{
				Version: v.Version,
				Assets:  assets,
			})
		}
	}

	return filteredVersions, nil
}
