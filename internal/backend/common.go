// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/version"
)

// CommonAsset represents a generic asset from a hosting platform.
type CommonAsset struct {
	Name            string
	URL             string
	Size            int64
	IsGlibcFallback bool
}

// containsWord checks whether word appears as a complete token in s,
// where token boundaries are defined as the start/end of string or any common
// filename separator character (-, _, ., space).
// This prevents false positives such as "win" matching "darwin".
func containsWord(s, word string) bool {
	isSep := func(b byte) bool {
		return b == '-' || b == '_' || b == '.' || b == ' '
	}
	for i := 0; i <= len(s)-len(word); {
		idx := strings.Index(s[i:], word)
		if idx < 0 {
			return false
		}
		abs := i + idx
		startOK := abs == 0 || isSep(s[abs-1])
		endAbs := abs + len(word)
		endOK := endAbs >= len(s) || isSep(s[endAbs])
		if startOK && endOK {
			return true
		}
		i = abs + 1
	}
	return false
}

// CalculateAssetScore calculates a compatibility score for an asset name.
// Returns -1 if the asset is definitely incompatible.
func CalculateAssetScore(assetName string, platform Platform, toolName string) int {
	nameLower := strings.ToLower(assetName)

	// 1. Hard Exclusions (Negative Score)
	excludeSuffixes := []string{
		// Checksums and signatures — never a binary
		".sha256", ".sha256sum", ".md5", ".asc", ".sig", ".sha1",
		// OS package formats — require OS-specific installation tooling
		".deb", ".rpm", ".msi", ".apk", ".pkg", ".snap", ".flatpak",
		// Source / document files
		".txt", ".pdf",
		// C/C++ header and library files
		".h", ".c", ".cpp", ".a", ".lib",
		// bzip2 archives — obsolete for binary distributions
		".tar.bz2", ".tbz2",
		// Security / attestation files
		".pem", ".crt", ".pub",
		// JSON files (SBOM, SLSA provenance, in-toto attestation)
		".json", ".jsonl",
		// macOS disk images — require manual mount+drag, not CLI-installable
		".dmg",
		// Markdown and YAML docs/configs — never runnable binaries
		".md", ".yaml", ".yml",
	}
	for _, suffix := range excludeSuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return -1
		}
	}

	// Determine the short tool name to avoid false positives in negative keyword checks.
	toolShortName := toolName
	if parts := strings.Split(toolName, "/"); len(parts) == 2 {
		toolShortName = parts[1]
	}
	toolShortName = strings.TrimPrefix(toolShortName, "github:")
	toolShortName = strings.ToLower(toolShortName)

	// Exclude non-runtime assets.
	// NOTE: short keywords ("dev", "doc", "man", "sbom") use word-boundary
	// containsWord() instead of plain Contains to avoid false-positive matches:
	//   "doc"  would match "docker", "document"
	//   "man"  would match "manifest", "manager", "command"
	//   "dev"  would match "devops", "devenv"
	//   "sbom" would match tool names with "sbom" as a substring
	negatives := []string{"checksums", "sha256sums", "license", "source", "devel", "header", "static-lib", "manual", "debug", "provenance", "attestation"}

	for _, neg := range negatives {
		if strings.Contains(nameLower, neg) {
			// If the negative keyword is part of the tool name itself, don't exclude.
			// Example: "addlicense" contains "license".
			if strings.Contains(toolShortName, neg) {
				continue
			}
			return -1
		}
	}
	// Word-boundary checks for short keywords to avoid false-positive substring matches.
	if containsWord(nameLower, "dev") && !strings.Contains(toolShortName, "dev") {
		return -1
	}
	if containsWord(nameLower, "doc") && !strings.Contains(toolShortName, "doc") {
		return -1
	}
	if containsWord(nameLower, "man") && !strings.Contains(toolShortName, "man") {
		return -1
	}
	if containsWord(nameLower, "sbom") && !strings.Contains(toolShortName, "sbom") {
		return -1
	}
	// 'symbols' packages contain debug information, not runnable binaries.
	// Use word-boundary to avoid matching tool names that contain 'symbols'.
	if containsWord(nameLower, "symbols") && !strings.Contains(toolShortName, "symbols") {
		return -1
	}
	// Exclude WebAssembly targets — wasm/wasi binaries cannot run natively.
	if strings.Contains(nameLower, "wasm") || strings.Contains(nameLower, "wasi") {
		return -1
	}

	score := 0

	// 2. OS Match
	osMatch := false
	switch platform.OS {
	case "linux":
		// Direct Linux keyword matches.
		if strings.Contains(nameLower, "linux") || strings.Contains(nameLower, "unknown-linux") {
			// Exclude android targets — android uses the Linux kernel but has a
			// different ABI (Bionic libc) and cannot run standard Linux binaries.
			if strings.Contains(nameLower, "android") {
				return -1
			}
			osMatch = true
			score += 100
		}
		// musl implies Linux — musl libc only runs on Linux, so an asset named
		// e.g. "tool-musl-amd64" implicitly targets Linux even without the word.
		if !osMatch && strings.Contains(nameLower, "musl") {
			osMatch = true
			score += 80 // slightly lower than explicit linux
		}
		// alpine implies Linux+musl — Alpine Linux assets often omit "linux".
		if !osMatch && strings.Contains(nameLower, "alpine") {
			osMatch = true
			score += 80
		}
	case "darwin":
		if strings.Contains(nameLower, "darwin") || strings.Contains(nameLower, "macos") || strings.Contains(nameLower, "osx") || strings.Contains(nameLower, "apple") {
			osMatch = true
			score += 100
		}
	case "windows":
		// NOTE: do NOT use plain strings.Contains(nameLower, "win") here — it would
		// match "dar*win*" inside darwin asset names. Use containsWord instead.
		if strings.Contains(nameLower, "windows") ||
			strings.Contains(nameLower, "win64") ||
			strings.Contains(nameLower, "win32") ||
			containsWord(nameLower, "win") ||
			strings.HasSuffix(nameLower, ".exe") {
			osMatch = true
			score += 100
		} else if strings.HasSuffix(nameLower, ".zip") && !strings.Contains(nameLower, "linux") && !strings.Contains(nameLower, "darwin") && !strings.Contains(nameLower, "macos") && !strings.Contains(nameLower, "apple") {
			osMatch = true
			score += 50
		}
	}

	if !osMatch {
		return -1
	}

	// 3. Architecture Match
	archMatch := false
	// isDarwinUniversal detects macOS universal/fat binaries that work on any architecture.
	// Some projects publish a single "darwin-all" or "darwin-universal" binary instead of
	// separate amd64/arm64 builds (e.g., editorconfig-checker v4.0.0).
	isDarwinUniversal := platform.OS == "darwin" &&
		(strings.Contains(nameLower, "universal") || containsWord(nameLower, "all"))

	switch platform.Arch {
	case "amd64":
		// amd64, x86_64, x86-64: standard 64-bit x86 naming (underscore and hyphen variants).
		// x64: shorthand used by some tools.
		// 64bit: suffix used by a few older projects.
		if strings.Contains(nameLower, "amd64") || strings.Contains(nameLower, "x86_64") ||
			strings.Contains(nameLower, "x86-64") || strings.Contains(nameLower, "x64") ||
			strings.Contains(nameLower, "64bit") {
			archMatch = true
			score += 100
		} else if isDarwinUniversal {
			archMatch = true
			score += 80 // Lower than exact match so darwin-amd64 is preferred over darwin-all
		} else if platform.OS == "windows" && strings.HasSuffix(nameLower, ".zip") && !strings.Contains(nameLower, "386") && !strings.Contains(nameLower, "arm64") && !strings.Contains(nameLower, "aarch64") && !strings.Contains(nameLower, "armv8") {
			archMatch = true
			score += 50
		}
	case "arm64":
		// arm64, aarch64: standard 64-bit ARM names.
		// armv8, armv8l: explicit ARMv8 naming (little-endian variant included).
		// arm64e: Apple Silicon variant used in some Homebrew/macOS tools.
		if strings.Contains(nameLower, "arm64") || strings.Contains(nameLower, "aarch64") ||
			strings.Contains(nameLower, "armv8") || strings.Contains(nameLower, "arm64e") {
			archMatch = true
			score += 100
		} else if isDarwinUniversal {
			archMatch = true
			score += 80 // Lower than exact match so darwin-arm64 is preferred over darwin-all
		}
	case "386":
		// i386, i686: common 32-bit x86 naming conventions.
		// x86: generic 32-bit x86 identifier.
		// 32bit: some tools use "32bit" as a suffix.
		if strings.Contains(nameLower, "386") || strings.Contains(nameLower, "i386") ||
			strings.Contains(nameLower, "i686") || strings.Contains(nameLower, "x86") ||
			strings.Contains(nameLower, "32bit") {
			archMatch = true
			score += 100
		}
	case "arm":
		// armhf: ARM Hard Float — used in Raspberry Pi OS 32-bit and Debian armhf.
		// armv7l, armv7: ARMv7 (Cortex-A) — the most common 32-bit ARM today.
		// armv6l, armv6: ARMv6 — older Raspberry Pi (1st gen, Zero).
		// arm (word boundary): generic 32-bit ARM when no version suffix is given.
		// NOTE: containsWord is essential here — bare Contains("arm") would
		//       also match "arm64", "aarch64", and "armv8" (64-bit targets).
		if strings.Contains(nameLower, "armhf") ||
			strings.Contains(nameLower, "armv7") ||
			strings.Contains(nameLower, "armv6") ||
			containsWord(nameLower, "arm") {
			archMatch = true
			score += 100
		}
	}

	if !archMatch {
		return -1
	}

	// 4. Preferred Formats (scores reflect unpack ergonomics and cross-platform prevalence).
	switch {
	case strings.HasSuffix(nameLower, ".tar.gz") || strings.HasSuffix(nameLower, ".tgz"):
		score += 50
	case strings.HasSuffix(nameLower, ".zip"):
		score += 40
	case strings.HasSuffix(nameLower, ".tar.xz") || strings.HasSuffix(nameLower, ".txz"):
		score += 30
	case strings.HasSuffix(nameLower, ".tar.zst") || strings.HasSuffix(nameLower, ".tzst"):
		score += 28 // zstd: fast modern format, slightly below xz
	case !strings.Contains(nameLower, ".") || strings.HasSuffix(nameLower, ".exe"):
		score += 20 // Raw binary or Windows exe
	}

	// 5. Tool Name Bonus
	repoName := ""
	if parts := strings.Split(toolName, "/"); len(parts) == 2 {
		repoName = strings.ToLower(parts[1])
	}
	if repoName != "" && strings.Contains(nameLower, repoName) {
		score += 50
	}

	// 6. GNU libc soft preference (non-musl Linux only).
	// When the target is a standard glibc Linux, give a small bonus to assets
	// that explicitly declare the gnu ABI in their target triple
	// (e.g. x86_64-unknown-linux-gnu). This makes the scorer prefer the
	// canonical glibc build when both gnu and musl variants are published.
	if platform.OS == "linux" && !platform.Musl && strings.Contains(nameLower, "-gnu") {
		score += 10
	}

	hasMusl := strings.Contains(nameLower, "musl")
	if platform.Musl {
		if hasMusl {
			score += 50
		} else if platform.OS == "linux" {
			score -= 10
		}
	} else if hasMusl {
		score -= 50
	}

	return score
}

// FindBestAsset finds the best matching asset for a platform from a list.
func FindBestAsset(assets []CommonAsset, platform Platform, toolName string) (*CommonAsset, int) {
	var bestAsset *CommonAsset
	bestScore := -1

	for i := range assets {
		asset := &assets[i]
		score := CalculateAssetScore(asset.Name, platform, toolName)
		if score > 0 && score > bestScore {
			bestScore = score
			bestAsset = asset
		}
	}

	if bestAsset != nil && platform.Musl {
		nameLower := strings.ToLower(bestAsset.Name)
		if !strings.Contains(nameLower, "musl") && !strings.Contains(nameLower, "alpine") {
			bestAsset.IsGlibcFallback = true
		}
	}

	return bestAsset, bestScore
}

// FindBestAssetWithOverride is like FindBestAsset but first checks assetPatterns
// for an exact asset filename override keyed by platformKey (e.g. "macos-amd64").
// When an override is found the named asset is located in the release and returned
// directly, bypassing the heuristic scoring entirely.
//
// assetPatterns may be nil (or empty), in which case normal scoring is used.
func FindBestAssetWithOverride(assets []CommonAsset, platform Platform, toolName, platformKey string, assetPatterns map[string]string) (*CommonAsset, int) {
	if len(assetPatterns) > 0 {
		if wantName, ok := assetPatterns[platformKey]; ok && wantName != "" {
			for i := range assets {
				if assets[i].Name == wantName {
					return &assets[i], 1000 // override score — always wins
				}
			}
			// Override specified but asset not found in release — fall through to
			// normal matching so we still get something rather than nothing.
		}
	}
	return FindBestAsset(assets, platform, toolName)
}

// FetchAndParseChecksumFile downloads and parses a checksum file from a URL.
func FetchAndParseChecksumFile(ctx context.Context, client *http.Client, url string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch checksum file: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	checksums := make(map[string]string)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "checksum  filename" or "checksum filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[1]
			// Handle format where filename is prefixed with * or space
			filename = strings.TrimPrefix(filename, "*")
			checksums[filename] = checksum
		}
	}

	return checksums, nil
}

// FindChecksumForAsset attempts to find a matching checksum for an asset from a list of all assets.
func FindChecksumForAsset(ctx context.Context, client *http.Client, assets []CommonAsset, targetAsset *CommonAsset) (string, error) {
	if targetAsset == nil {
		return "", nil
	}

	// 1. Look for a checksum file
	var checksumAsset *CommonAsset
	for i := range assets {
		nameLower := strings.ToLower(assets[i].Name)
		if strings.HasSuffix(nameLower, ".sha256") ||
			strings.HasSuffix(nameLower, ".sha256sum") ||
			strings.Contains(nameLower, "checksums") ||
			strings.Contains(nameLower, "sha256sums") {
			checksumAsset = &assets[i]
			break
		}
	}

	if checksumAsset != nil {
		checksumMap, err := FetchAndParseChecksumFile(ctx, client, checksumAsset.URL)
		if err != nil {
			return "", err
		}
		if checksumMap != nil {
			// Try exact match first
			if c, ok := checksumMap[targetAsset.Name]; ok {
				return c, nil
			}
		}
	}

	return "", nil
}

// FindGPGSignatureForAsset attempts to find a matching GPG signature for an asset.
func FindGPGSignatureForAsset(assets []CommonAsset, targetAsset *CommonAsset) string {
	if targetAsset == nil {
		return ""
	}

	// Look for filename.asc or filename.sig
	ascName := targetAsset.Name + ".asc"
	sigName := targetAsset.Name + ".sig"

	for i := range assets {
		if assets[i].Name == ascName || assets[i].Name == sigName {
			return assets[i].URL
		}
	}

	return ""
}

// HostingProvider defines the interface for fetching data from a hosting platform (GitHub, GitLab, etc.).
type HostingProvider interface {
	Name() string
	FetchReleases(ctx context.Context, tool string) ([]CommonRelease, error)
	FetchReleaseByTag(ctx context.Context, tool, tag string) (*CommonRelease, error)
	GetAttestationType() string
	GetClient() *http.Client
}

// CommonRelease represents a generic release from any hosting platform.
type CommonRelease struct {
	PublishedAt time.Time
	Tag         string
	Assets      []CommonAsset
	Prerelease  bool
}

// GenericResolveVersion implements the common logic for resolving a version request.
func GenericResolveVersion(ctx context.Context, p HostingProvider, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	releases, err := p.FetchReleases(ctx, tool)
	if err != nil {
		return nil, err
	}

	var versions []VersionInfo
	for _, release := range releases {
		// Find matching asset for platform
		bestAsset, _ := FindBestAsset(release.Assets, platform, tool)
		if bestAsset == nil {
			continue
		}

		v := strings.TrimPrefix(release.Tag, "v")
		versions = append(versions, VersionInfo{
			Version:     v,
			DownloadURL: bestAsset.URL,
			Platform:    platform,
			PublishedAt: release.PublishedAt,
			Metadata: map[string]string{
				"prerelease":      fmt.Sprintf("%t", release.Prerelease),
				"IsGlibcFallback": fmt.Sprintf("%t", bestAsset.IsGlibcFallback),
			},
		})
	}

	if len(versions) == 0 {
		return nil, NewBackendError(p.Name(), tool, "no suitable releases found", nil)
	}

	// Resolution logic
	switch versionRequest {
	case "latest", "stable":
		// Filter out floating tags to prevent them from breaking SemVer latest detection
		var validVersions []VersionInfo
		for _, v := range versions {
			lower := strings.ToLower(v.Version)
			if lower == "stable" || lower == "latest" || lower == "nightly" || lower == "master" || lower == "main" {
				continue
			}
			validVersions = append(validVersions, v)
		}
		if len(validVersions) > 0 {
			versions = validVersions
		}

		// Sort versions in descending order using SemVer logic
		sort.Slice(versions, func(i, j int) bool {
			return version.CompareVersions(versions[i].Version, versions[j].Version) > 0
		})
		for _, v := range versions {
			if v.Metadata["prerelease"] == "false" {
				return &v, nil
			}
		}
		return &versions[0], nil
	default:
		reqV := strings.TrimPrefix(versionRequest, "v")
		for _, v := range versions {
			if v.Version == reqV {
				return &v, nil
			}
		}
		return nil, NewBackendError(p.Name(), tool, "version not found", nil)
	}
}

// GenericGetDownloadInfo implements the common logic for retrieving download info.
func GenericGetDownloadInfo(ctx context.Context, p HostingProvider, tool, version string, platform Platform) (*VersionInfo, error) {
	return GenericGetDownloadInfoWithPatterns(ctx, p, tool, version, "", platform, nil)
}

// GenericGetDownloadInfoWithPatterns is like GenericGetDownloadInfo but accepts
// assetPatterns (a map from platform key -> exact asset filename) to bypass
// heuristic scoring for the given platform.
//
// platformKey is the canonical lock key for the target platform (e.g. "macos-amd64").
// When assetPatterns[platformKey] is set the asset is selected by exact name.
//
// Fallback order:
//  1. User-configured asset_patterns (assetPatterns arg) — highest priority
//  2. Heuristic asset scoring (CalculateAssetScore)
//  3. Built-in community AssetRegistry (LookupRegistryPatterns) — automatic fallback
func GenericGetDownloadInfoWithPatterns(ctx context.Context, p HostingProvider, tool, version, platformKey string, platform Platform, assetPatterns map[string]string) (*VersionInfo, error) {
	tag := version
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}

	release, err := p.FetchReleaseByTag(ctx, tool, tag)
	if err != nil {
		// try without 'v'
		tag = version
		release, err = p.FetchReleaseByTag(ctx, tool, tag)
		if err != nil {
			return nil, NewBackendError(p.Name(), tool, "release not found", err)
		}
	}

	// Priority 1 & 2: user patterns + heuristic scoring.
	bestAsset, _ := FindBestAssetWithOverride(release.Assets, platform, tool, platformKey, assetPatterns)

	// Priority 3: built-in community registry fallback.
	// Only consulted when there are no user-level patterns configured AND the
	// heuristic found nothing.
	if bestAsset == nil && len(assetPatterns) == 0 {
		if registryEntry, ok := LookupRegistryPatterns(tool); ok && len(registryEntry.Patterns) > 0 {
			bestAsset, _ = FindBestAssetWithOverride(release.Assets, platform, tool, platformKey, registryEntry.Patterns)
		}
	}

	if bestAsset == nil {
		return nil, NewBackendError(p.Name(), tool, "no matching asset", nil)
	}

	checksum, err := FindChecksumForAsset(ctx, p.GetClient(), release.Assets, bestAsset)
	if err != nil {
		return nil, NewBackendError(p.Name(), tool, "failed to fetch checksum file", err)
	}
	gpgSigURL := FindGPGSignatureForAsset(release.Assets, bestAsset)

	return &VersionInfo{
		Version:     version,
		DownloadURL: bestAsset.URL,
		Checksum:    checksum,
		Platform:    platform,
		PublishedAt: release.PublishedAt,
		Metadata: map[string]string{
			"gpg_signature_url": gpgSigURL,
			"IsGlibcFallback":   fmt.Sprintf("%t", bestAsset.IsGlibcFallback),
		},
	}, nil
}

// ProbeURL checks if a URL is accessible via HEAD request.
func ProbeURL(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// NormalizeVersionPrefix intelligently ensures the version string has the correct 'v' prefix behavior.
// If requireV is true, it prepends 'v' if the version starts with a digit.
// If requireV is false, it strips 'v' or 'V' if it's followed by a digit.
func NormalizeVersionPrefix(versionRequest string, requireV bool) string {
	if versionRequest == "latest" || versionRequest == "stable" {
		return versionRequest
	}

	if !requireV {
		if (strings.HasPrefix(versionRequest, "v") || strings.HasPrefix(versionRequest, "V")) && len(versionRequest) > 1 {
			if versionRequest[1] >= '0' && versionRequest[1] <= '9' {
				return versionRequest[1:]
			}
		}
		return versionRequest
	} else {
		if !strings.HasPrefix(versionRequest, "v") && len(versionRequest) > 0 {
			if versionRequest[0] >= '0' && versionRequest[0] <= '9' {
				return "v" + versionRequest
			}
		}
		return versionRequest
	}
}
