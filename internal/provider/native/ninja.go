// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"strings"
)

// NinjaHandler handles Ninja build tool versions via GitHub releases.
type NinjaHandler struct {
	GithubHandler
}

func (h *NinjaHandler) Name() string {
	return "ninja"
}

func (h *NinjaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "ninja-build"
	h.Repo = "ninja"
	
	versions, err := h.GithubHandler.ResolveVersions(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	// Post-process to fix Ninja's specific naming if GithubHandler missed it.
	// Ninja releases usually look like: ninja-linux.zip, ninja-mac.zip, ninja-win.zip
	for i := range versions {
		for j := range versions[i].Assets {
			asset := &versions[i].Assets[j]
			if asset.OS == "" {
				lowerName := strings.ToLower(asset.Filename)
				if strings.Contains(lowerName, "linux") {
					asset.OS = "linux"
				} else if strings.Contains(lowerName, "mac") {
					asset.OS = "darwin"
				} else if strings.Contains(lowerName, "win") {
					asset.OS = "windows"
				}
			}
			// Ninja binaries are usually x86_64 unless specified
			if asset.Arch == "" && asset.OS != "" {
				asset.Arch = "amd64"
			}
		}
	}

	return versions, nil
}
