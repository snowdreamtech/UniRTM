// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
)

func cleanVersionReq(val, tool string) string {
	val = strings.TrimSpace(val)
	if (strings.HasPrefix(val, "v") || strings.HasPrefix(val, "V")) && len(val) > 1 && unicode.IsDigit(rune(val[1])) {
		val = val[1:]
	}
	if tool == "go" || tool == "golang" || strings.Contains(tool, "go") {
		if strings.HasPrefix(val, "go") && len(val) > 2 && unicode.IsDigit(rune(val[2])) {
			val = val[2:]
		}
	}
	return val
}

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
				score := 1000
				if a.Filename != "" {
					fileScore := CalculateAssetScore(a.Filename, platform, tool)
					if fileScore < 0 {
						continue
					}
					score += fileScore
				}
				if score > bestScore {
					bestScore = score
					bestAsset = &a
				}
				continue
			}
			if a.Filename != "" {
				score := CalculateAssetScore(a.Filename, platform, tool)
				if score > bestScore && score >= 0 {
					bestScore = score
					bestAsset = &a
				}
			}
		}

		if bestAsset != nil && platform.Musl {
			nameLower := strings.ToLower(bestAsset.Filename)
			if !strings.Contains(nameLower, "musl") && !strings.Contains(nameLower, "alpine") {
				if bestAsset.Metadata == nil {
					bestAsset.Metadata = make(map[string]string)
				}
				bestAsset.Metadata["IsGlibcFallback"] = "true"
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

func (b *NativeBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	recipe, ok := b.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	req := cleanVersionReq(versionRequest, tool)

	// 1. Check for static aliases
	if alias, ok := recipe.Aliases[req]; ok {
		req = alias
	}

	// Fetch versions to resolve dynamic aliases / prefixes
	versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("native: no versions found for %s", tool)
	}

	// 2. Handle "latest"
	if strings.EqualFold(req, "latest") {
		req = versions[0].Version
	} else if strings.EqualFold(req, "lts") || req == "lts/*" {
		// 3. Handle dynamic LTS
		foundLTS := false
		for _, v := range versions {
			if v.IsLTS {
				req = v.Version
				foundLTS = true
				break
			}
		}
		if !foundLTS {
			if alias, ok := recipe.Aliases["lts"]; ok {
				req = alias
			} else {
				req = versions[0].Version
			}
		}
	} else if strings.HasPrefix(strings.ToLower(req), "lts/") {
		// 4. Handle LTS by codename (e.g. lts/iron, lts/hydrogen)
		codename := strings.ToLower(strings.TrimPrefix(req, "lts/"))
		foundNamedLTS := false
		for _, v := range versions {
			if v.IsLTS && strings.EqualFold(v.LTSName, codename) {
				req = v.Version
				foundNamedLTS = true
				break
			}
		}
		if !foundNamedLTS {
			return nil, fmt.Errorf("native: no LTS release found with codename %q for %s", codename, tool)
		}
	} else {
		// 5. Check if exact match exists
		exactMatch := false
		for _, v := range versions {
			if v.Version == req {
				exactMatch = true
				break
			}
		}

		// 6. Fuzzy prefix / semver matching (e.g. "20", "20.x", "1.23", "3.12")
		if !exactMatch {
			cleanPrefix := strings.TrimSuffix(strings.TrimSuffix(req, ".*"), ".x")
			for _, v := range versions {
				if strings.HasPrefix(v.Version, cleanPrefix+".") || v.Version == cleanPrefix {
					req = v.Version
					exactMatch = true
					break
				}
			}
		}
	}

	return b.GetDownloadInfo(ctx, tool, req, platform)
}

func (b *NativeBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	recipe, ok := b.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	version = cleanVersionReq(version, tool)

	var targetVersion *native.VersionInfo
	versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
	if err == nil {
		for _, v := range versions {
			if v.Version == version {
				targetVersion = &v
				break
			}
		}
	}

	if targetVersion == nil {
		// Fallback: If handler supports direct tag resolution (e.g. GithubHandler.ResolveTag)
		if tagResolver, ok := recipe.Handler.(interface {
			ResolveTag(ctx context.Context, version string) (*native.VersionInfo, error)
		}); ok {
			if resolved, tagErr := tagResolver.ResolveTag(ctx, version); tagErr == nil && resolved != nil {
				targetVersion = resolved
			}
		}
	}

	if targetVersion == nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("native: version %s not found for %s", version, tool)
	}

	var bestAsset *native.Asset
	bestScore := -1

	for _, a := range targetVersion.Assets {
		// 1. Strict Match
		if a.OS == platform.OS && a.Arch == platform.Arch {
			score := 1000
			if a.Filename != "" {
				fileScore := CalculateAssetScore(a.Filename, platform, tool)
				if fileScore < 0 {
					// Exclude negative assets (e.g. debug, checksums, source, etc.)
					continue
				}
				score += fileScore
			}
			if score > bestScore {
				bestScore = score
				bestAsset = &a
			}
			continue
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

	if bestAsset != nil && platform.Musl {
		nameLower := strings.ToLower(bestAsset.Filename)
		if !strings.Contains(nameLower, "musl") && !strings.Contains(nameLower, "alpine") {
			if bestAsset.Metadata == nil {
				bestAsset.Metadata = make(map[string]string)
			}
			bestAsset.Metadata["IsGlibcFallback"] = "true"
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
