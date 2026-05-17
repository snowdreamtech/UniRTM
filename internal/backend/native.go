// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
)

// NativeBackend implements the Backend interface using built-in native recipes.
type NativeBackend struct {
	recipes map[string]native.Recipe
}

func NewNativeBackend() *NativeBackend {
	return &NativeBackend{
		recipes: native.GetBuiltinRecipes(),
	}
}

func (b *NativeBackend) Name() string {
	return "native"
}

func (b *NativeBackend) Dependencies() []string {
	return nil
}
func (b *NativeBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	recipe, ok := b.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
	if err != nil {
		return nil, err
	}

	var res []VersionInfo
	for _, v := range versions {
		vi := VersionInfo{
			Version: v.Version,
		}

		// Find matching asset to fill in details if possible
		var bestAsset *native.Asset
		bestScore := -1
		for _, a := range v.Assets {
			if a.OS == platform.OS && a.Arch == platform.Arch {
				bestAsset = &a
				bestScore = 999
				break
			}
			if a.Filename != "" {
				score := CalculateAssetScore(a.Filename, platform, tool)
				if score > bestScore && score >= 0 {
					bestScore = score
					bestAsset = &a
				}
			}
		}

		if bestAsset != nil {
			vi.DownloadURL = bestAsset.URL
			vi.Checksum = bestAsset.Checksum
			vi.SignatureURL = bestAsset.SignatureURL
			vi.GPGSignature = bestAsset.Signature
			vi.GPGKeys = recipe.GPGKeys
			vi.Platform = platform
			vi.Metadata = bestAsset.Metadata
		}

		res = append(res, vi)
	}
	return res, nil
}

func (b *NativeBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	recipe, ok := b.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	// 1. Check for aliases
	if alias, ok := recipe.Aliases[versionRequest]; ok {
		versionRequest = alias
	}

	// 2. Handle "latest" if not already resolved by alias
	if versionRequest == "latest" {
		versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, fmt.Errorf("native: no versions found for %s", tool)
		}
		// Use the first version as latest (assuming handlers return newest first)
		versionRequest = versions[0].Version
	}

	return b.GetDownloadInfo(ctx, tool, versionRequest, platform)
}

func (b *NativeBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	recipe, ok := b.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
	if err != nil {
		return nil, err
	}

	var targetVersion *native.VersionInfo
	for _, v := range versions {
		if v.Version == version {
			targetVersion = &v
			break
		}
	}

	if targetVersion == nil {
		return nil, fmt.Errorf("native: version %s not found for %s", version, tool)
	}

	var bestAsset *native.Asset
	bestScore := -1

	for _, a := range targetVersion.Assets {
		// 1. Strict Match (Priority: 999)
		if a.OS == platform.OS && a.Arch == platform.Arch {
			bestAsset = &a
			bestScore = 999
			break
		}

		// 2. Guessing Logic (Fallback)
		if a.Filename != "" {
			score := CalculateAssetScore(a.Filename, platform, tool)
			if score > bestScore && score >= 0 {
				bestScore = score
				bestAsset = &a
			}
		}
	}

	if bestAsset == nil {
		return nil, fmt.Errorf("native: no compatible asset found for %s %s on %s/%s", tool, version, platform.OS, platform.Arch)
	}

	return &VersionInfo{
		Version:      version,
		DownloadURL:  bestAsset.URL,
		Checksum:     bestAsset.Checksum,
		SignatureURL: bestAsset.SignatureURL,
		GPGSignature: bestAsset.Signature,
		GPGKeys:      recipe.GPGKeys,
		Platform:     platform,
		Metadata:     bestAsset.Metadata,
	}, nil
}

func (b *NativeBackend) SupportsChecksum() bool {
	return true
}

func (b *NativeBackend) SupportsGPG() bool {
	return true
}

func (b *NativeBackend) AttestationType() string {
	return "Native"
}

func (b *NativeBackend) IsRecommended() bool {
	return true
}

func (b *NativeBackend) IsScriptless() bool {
	return true
}

func (b *NativeBackend) GetReach() string {
	return "Small"
}

func (b *NativeBackend) IsStable() bool {
	return true
}

func (b *NativeBackend) SupportsOffline() bool {
	return true
}
