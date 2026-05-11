// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"fmt"
)

// RustHandler handles Rust distributions from static.rust-lang.org.
type RustHandler struct{}

func (h *RustHandler) Name() string {
	return "rust"
}

func (h *RustHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// For Rust, instead of a heavy TOML parser, we'll start with a curated list 
	// or use a simpler version discovery.
	// For now, let's provide the recent stable versions.
	// In a real implementation, we would fetch and parse the manifest.
	
	stableVersions := []string{"1.76.0", "1.75.0", "1.74.1", "1.74.0"}
	
	var versions []VersionInfo
	for _, v := range stableVersions {
		versions = append(versions, VersionInfo{
			Version: v,
			Assets:  h.generateAssets(v),
		})
	}

	return versions, nil
}

func (h *RustHandler) generateAssets(version string) []Asset {
	var assets []Asset
	
	// Common Rust targets
	targets := map[string]struct{ os, arch string }{
		"x86_64-unknown-linux-gnu":  {"linux", "amd64"},
		"aarch64-unknown-linux-gnu": {"linux", "arm64"},
		"x86_64-apple-darwin":       {"darwin", "amd64"},
		"aarch64-apple-darwin":      {"darwin", "arm64"},
		"x86_64-pc-windows-msvc":    {"windows", "amd64"},
	}

	for target, platform := range targets {
		url := fmt.Sprintf("https://static.rust-lang.org/dist/rust-%s-%s.tar.gz", version, target)
		if platform.os == "windows" {
			url = fmt.Sprintf("https://static.rust-lang.org/dist/rust-%s-%s.zip", version, target)
		}

		assets = append(assets, Asset{
			Filename: fmt.Sprintf("rust-%s-%s.tar.gz", version, target),
			URL:      url,
			OS:       platform.os,
			Arch:     platform.arch,
		})
	}

	return assets
}
