// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"fmt"
	"runtime"
	"strings"
)

// StandardPlatforms is the canonical list of platforms that `unirtm lock --all-platforms`
// will resolve. These mirror the platform keys used in mise.lock.
var StandardPlatforms = []string{
	"linux-amd64",
	"linux-amd64-musl",
	"linux-arm64",
	"linux-arm64-musl",
	"macos-amd64",
	"macos-arm64",
	"windows-amd64",
	"windows-arm64",
}

// osNames maps GOOS → lockfile OS segment.
var osNames = map[string]string{
	"linux":   "linux",
	"darwin":  "macos",
	"windows": "windows",
}

// archNames maps GOARCH → lockfile arch segment.
var archNames = map[string]string{
	"amd64": "amd64",
	"arm64": "arm64",
	"386":   "386",
	"arm":   "arm",
}

// CurrentPlatformKey returns the platform key for the running OS/arch,
// e.g. "linux-amd64", "macos-arm64", "windows-amd64".
func CurrentPlatformKey() string {
	return PlatformKey(runtime.GOOS, runtime.GOARCH, false)
}

// PlatformKey builds a canonical lockfile platform key from GOOS/GOARCH components.
//
//	PlatformKey("linux",   "amd64", false) → "linux-amd64"
//	PlatformKey("darwin",  "arm64", false) → "macos-arm64"
//	PlatformKey("windows", "amd64", false) → "windows-amd64"
//	PlatformKey("linux",   "amd64", true)  → "linux-amd64-musl"
func PlatformKey(goos, goarch string, musl bool) string {
	os, ok := osNames[goos]
	if !ok {
		os = goos // pass-through for unknown OS
	}
	arch, ok := archNames[goarch]
	if !ok {
		arch = goarch
	}
	key := os + "-" + arch
	if musl {
		key += "-musl"
	}
	return key
}

// ParsePlatformKey parses a canonical platform key back into (goos, goarch, musl).
// Returns an error for keys that do not follow the expected format.
func ParsePlatformKey(key string) (goos, goarch string, musl bool, err error) {
	parts := strings.Split(key, "-")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", false, fmt.Errorf("lockfile: invalid platform key %q (expected os-arch or os-arch-musl)", key)
	}

	osKey := parts[0]
	archKey := parts[1]

	if len(parts) == 3 {
		if parts[2] != "musl" {
			return "", "", false, fmt.Errorf("lockfile: invalid platform key %q (unknown suffix %q)", key, parts[2])
		}
		musl = true
	}

	// Reverse-map OS.
	for goosVal, lockOS := range osNames {
		if lockOS == osKey {
			goos = goosVal
			break
		}
	}
	if goos == "" {
		goos = osKey // pass-through
	}

	// Reverse-map arch.
	for goarchVal, lockArch := range archNames {
		if lockArch == archKey {
			goarch = goarchVal
			break
		}
	}
	if goarch == "" {
		goarch = archKey // pass-through
	}

	return goos, goarch, musl, nil
}

// IsValidPlatformKey reports whether key is a recognised standard platform key.
func IsValidPlatformKey(key string) bool {
	for _, p := range StandardPlatforms {
		if p == key {
			return true
		}
	}
	return false
}

// NormalizePlatformKey lower-cases and validates a user-supplied platform key.
// It accepts both "linux-x64" (mise-style) and "linux-amd64" (unirtm-style) forms.
func NormalizePlatformKey(raw string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(raw))

	// Accept mise-style aliases.
	aliases := map[string]string{
		"linux-x64":    "linux-amd64",
		"linux-x64-musl": "linux-amd64-musl",
		"macos-x64":    "macos-amd64",
		"windows-x64":  "windows-amd64",
		"linux-aarch64": "linux-arm64",
		"macos-aarch64": "macos-arm64",
	}
	if canonical, ok := aliases[key]; ok {
		key = canonical
	}

	if !IsValidPlatformKey(key) {
		return "", fmt.Errorf("lockfile: unknown platform key %q; supported: %v", raw, StandardPlatforms)
	}
	return key, nil
}

// ParsePlatformKeys parses a comma-separated list of platform keys (user input),
// normalising each one and deduplicating.
func ParsePlatformKeys(input string) ([]string, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var out []string

	for _, raw := range strings.Split(input, ",") {
		k, err := NormalizePlatformKey(raw)
		if err != nil {
			return nil, err
		}
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out, nil
}
