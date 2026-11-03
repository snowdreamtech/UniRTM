// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"strconv"
	"strings"
)

// CompareVersions robustly compares two version strings.
// It leverages UniRTM's comprehensive internal Version parser to handle semver, date-versions,
// and aliases (like "latest" vs "lts").
// If one or both versions fail to parse (e.g. non-standard alphanumeric directories like "tip"),
// it falls back to a segment-based numerical splitting algorithm to gracefully handle any string format.
// Returns:
//
//	1 if v1 > v2
//
// -1 if v1 < v2
//
//	0 if v1 == v2
func CompareVersions(v1, v2 string) int {
	ver1, err1 := ParseVersion(v1)
	ver2, err2 := ParseVersion(v2)

	// If both parsed successfully, use internal standard comparators
	if err1 == nil && err2 == nil {
		// If both are exact versions, use SemVer comparator
		if ver1.Type == VersionTypeExact && ver2.Type == VersionTypeExact {
			return ver1.Exact.Compare(ver2.Exact)
		}

		// If both are aliases, we could compare them or just fall back to numerical/string comparison
		if ver1.Type == VersionTypeAlias && ver2.Type == VersionTypeAlias {
			if ver1.Alias == ver2.Alias {
				return 0
			}
			// Let it fall through to fallback comparator
		}
	}

	// Graceful fallback for non-standard unparsable directories (e.g. "tip") or mixed types
	return compareVersionsFallback(v1, v2)
}

// compareVersionsFallback splits strings by '.' and compares each segment numerically.
// If a segment is completely non-numeric (e.g. "tip"), it falls back to string comparison.
func compareVersionsFallback(v1, v2 string) int {
	// Strip any 'v' prefix
	v1 = strings.TrimPrefix(strings.TrimPrefix(v1, "v"), "V")
	v2 = strings.TrimPrefix(strings.TrimPrefix(v2, "v"), "V")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		p1 := "0"
		if i < len(parts1) {
			p1 = parts1[i]
		}
		p2 := "0"
		if i < len(parts2) {
			p2 = parts2[i]
		}

		n1, err1 := strconv.Atoi(p1)
		n2, err2 := strconv.Atoi(p2)

		if err1 == nil && err2 == nil {
			if n1 != n2 {
				if n1 > n2 {
					return 1
				}
				return -1
			}
		} else {
			if p1 != p2 {
				if p1 > p2 {
					return 1
				}
				return -1
			}
		}
	}
	return 0
}
