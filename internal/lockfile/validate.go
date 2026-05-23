// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"fmt"
	"strings"
)

// ValidationError accumulates multiple lockfile validation issues.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("lockfile validation failed (%d issue(s)):\n  - %s",
		len(e.Errors), strings.Join(e.Errors, "\n  - "))
}

func (e *ValidationError) add(msg string, args ...any) {
	e.Errors = append(e.Errors, fmt.Sprintf(msg, args...))
}

func (e *ValidationError) ok() bool { return len(e.Errors) == 0 }

// Validate checks the LockFile for structural and semantic correctness.
// It returns *ValidationError (with all discovered issues) or nil.
func (lf *LockFile) Validate() error {
	ve := &ValidationError{}

	for key, entries := range lf.Tools {
		if len(entries) == 0 {
			ve.add("tool %q: empty entry list", key)
			continue
		}

		// Check for duplicate versions within the same tool key.
		seen := make(map[string]bool)
		for _, e := range entries {
			if e.Version == "" {
				ve.add("tool %q: entry has empty version", key)
			}
			if seen[e.Version] {
				ve.add("tool %q: duplicate version %q", key, e.Version)
			}
			seen[e.Version] = true

			// Validate platform entries.
			for pk, pe := range e.Platforms {
				if !IsValidPlatformKey(pk) {
					ve.add("tool %q version %q: unknown platform key %q", key, e.Version, pk)
				}
				if pe == nil {
					ve.add("tool %q version %q platform %q: nil entry", key, e.Version, pk)
					continue
				}
				if err := validatePlatformEntry(key, e.Version, pk, pe); err != nil {
					ve.add("%s", err)
				}
			}
		}
	}

	if !ve.ok() {
		return ve
	}
	return nil
}

// validatePlatformEntry validates a single PlatformEntry.
func validatePlatformEntry(toolKey, version, platformKey string, pe *PlatformEntry) error {
	ve := &ValidationError{}
	ctx := fmt.Sprintf("tool %q version %q platform %q", toolKey, version, platformKey)

	if pe.Checksum != "" && !isValidChecksumFormat(pe.Checksum) {
		ve.add("%s: checksum %q must start with 'sha256:' or 'blake3:'", ctx, pe.Checksum)
	}
	if pe.Size < 0 {
		ve.add("%s: size must be ≥ 0, got %d", ctx, pe.Size)
	}
	if pe.URL != "" && !strings.HasPrefix(pe.URL, "http") {
		ve.add("%s: url %q does not look like a valid HTTP URL", ctx, pe.URL)
	}

	if !ve.ok() {
		return ve
	}
	return nil
}

// isValidChecksumFormat verifies that s begins with a known algorithm prefix.
func isValidChecksumFormat(s string) bool {
	return strings.HasPrefix(s, "sha256:") ||
		strings.HasPrefix(s, "blake3:") ||
		strings.HasPrefix(s, "sha512:")
}

// CheckStrict verifies that the lockfile contains a valid entry for each tool/platform
// pair in the required set. For URL-based backends, it ensures a URL is present.
// Used to enforce UNIRTM_LOCKED=1 / settings.locked=true.
//
// required is a slice of (toolKey, version, platformKey) tuples.
func (lf *LockFile) CheckStrict(required []LockRequirement) error {
	ve := &ValidationError{}

	for _, r := range required {
		entry := lf.GetEntry(r.ToolKey, r.Version)
		if entry == nil {
			ve.add(
				"strict mode: no locked entry for tool=%q version=%q — run `unirtm lock` first",
				r.ToolKey, r.Version,
			)
			continue
		}

		pe := lf.GetPlatform(r.ToolKey, r.Version, r.PlatformKey)
		if pe == nil {
			ve.add(
				"strict mode: no locked platform %q for tool=%q version=%q — run `unirtm lock` first",
				r.PlatformKey, r.ToolKey, r.Version,
			)
			continue
		}

		// Backends that download from explicit URLs must have the URL locked.
		// Package manager backends (npm, pipx, asdf, cargo, go) delegate resolution
		// natively so they legitimately have an empty URL.
		needsURL := true
		switch entry.Backend {
		case "npm", "pipx", "asdf", "cargo", "go":
			needsURL = false
		}

		if needsURL && pe.URL == "" {
			ve.add(
				"strict mode: no locked URL for tool=%q version=%q platform=%q — run `unirtm lock` first",
				r.ToolKey, r.Version, r.PlatformKey,
			)
		}
	}

	if !ve.ok() {
		return ve
	}
	return nil
}

// LockRequirement describes a (tool, version, platform) combination that must
// be present in the lockfile under strict mode.
type LockRequirement struct {
	ToolKey     string // e.g. "github:cli/cli"
	Version     string // e.g. "2.72.0"
	PlatformKey string // e.g. "linux-amd64"
}
