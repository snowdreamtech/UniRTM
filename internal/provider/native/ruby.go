// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"runtime"
	"strings"
)

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

	// The default GithubHandler uses a generic detectPlatform.
	// ruby-builder has very specific naming: ruby-3.3.0-ubuntu-22.04.tar.gz
	// We need to filter and ensure the assets match the current host environment.
	var filtered []VersionInfo
	for _, rel := range releases {
		var assets []Asset
		for _, a := range rel.Assets {
			if h.isMatch(a.Filename) {
				assets = append(assets, a)
			}
		}
		if len(assets) > 0 {
			filtered = append(filtered, VersionInfo{
				Version: rel.Version,
				Assets:  assets,
			})
		}
	}

	return filtered, nil
}

func (h *RubyHandler) isMatch(filename string) bool {
	filename = strings.ToLower(filename)
	if !strings.HasPrefix(filename, "ruby-") {
		return false
	}

	os := runtime.GOOS
	arch := runtime.GOARCH

	// Arch check: ruby-builder binaries are mostly x64 (amd64)
	if arch != "amd64" && !strings.Contains(filename, arch) {
		// If it's arm64, look for arm64 in filename
		if arch == "arm64" && !strings.Contains(filename, "arm64") {
			return false
		}
		if arch == "amd64" && !strings.Contains(filename, "x86_64") && !strings.Contains(filename, "amd64") {
			return false
		}
	}

	if os == "darwin" {
		return strings.Contains(filename, "macos")
	}

	if os == "linux" {
		// ruby-builder is very Ubuntu-centric.
		// We'll try to match ubuntu, but on other distros this might be risky.
		return strings.Contains(filename, "ubuntu")
	}

	return false
}
