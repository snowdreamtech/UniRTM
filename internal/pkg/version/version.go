// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionType represents the type of version specification
type VersionType int

const (
	// VersionTypeExact represents an exact semver version (e.g., "1.20.0")
	VersionTypeExact VersionType = iota
	// VersionTypeRange represents a version range (e.g., ">=1.20.0", "^3.11", "~2.7.0")
	VersionTypeRange
	// VersionTypeAlias represents a version alias (e.g., "latest", "lts", "stable")
	VersionTypeAlias
)

// RangeOperator represents the operator in a version range
type RangeOperator string

const (
	// RangeOperatorGTE represents ">=" (greater than or equal)
	RangeOperatorGTE RangeOperator = ">="
	// RangeOperatorGT represents ">" (greater than)
	RangeOperatorGT RangeOperator = ">"
	// RangeOperatorLTE represents "<=" (less than or equal)
	RangeOperatorLTE RangeOperator = "<="
	// RangeOperatorLT represents "<" (less than)
	RangeOperatorLT RangeOperator = "<"
	// RangeOperatorEQ represents "=" (equal)
	RangeOperatorEQ RangeOperator = "="
	// RangeOperatorCaret represents "^" (compatible with)
	RangeOperatorCaret RangeOperator = "^"
	// RangeOperatorTilde represents "~" (approximately equivalent to)
	RangeOperatorTilde RangeOperator = "~"
)

// VersionAlias represents known version aliases
type VersionAlias string

const (
	// VersionAliasLatest represents the latest version
	VersionAliasLatest VersionAlias = "latest"
	// VersionAliasLTS represents the latest LTS version
	VersionAliasLTS VersionAlias = "lts"
	// VersionAliasStable represents the latest stable version
	VersionAliasStable VersionAlias = "stable"
)

// SemVer represents a semantic version
type SemVer struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string // e.g., "alpha.1", "beta.2", "rc.1"
	Build      string // e.g., "20130313144700"
}

// Version represents a version specification
type Version struct {
	Type VersionType

	// For VersionTypeExact
	Exact *SemVer

	// For VersionTypeRange
	RangeOp  RangeOperator
	RangeVer *SemVer

	// For VersionTypeAlias
	Alias VersionAlias
}

// Regular expressions for version parsing
var (
	// semverRegex matches semantic versions with optional prerelease and build metadata
	// Examples: 1.2.3, 1.2.3-alpha.1, 1.2.3+build.123, 1.2.3-beta.2+build.456
	semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)

	// rangeRegex matches version ranges with operators
	// Examples: >=1.2.3, >1.0.0, <=2.0.0, <3.0.0, =1.2.3
	rangeRegex = regexp.MustCompile(`^(>=|>|<=|<|=)v?(\d+\.\d+\.\d+(?:-[0-9A-Za-z\-\.]+)?(?:\+[0-9A-Za-z\-\.]+)?)$`)

	// caretRegex matches caret ranges
	// Examples: ^1.2.3, ^0.2.5, ^0.0.4
	caretRegex = regexp.MustCompile(`^\^v?(\d+\.\d+\.\d+(?:-[0-9A-Za-z\-\.]+)?(?:\+[0-9A-Za-z\-\.]+)?)$`)

	// tildeRegex matches tilde ranges
	// Examples: ~1.2.3, ~1.2, ~1
	tildeRegex = regexp.MustCompile(`^~v?(\d+(?:\.\d+)?(?:\.\d+)?(?:-[0-9A-Za-z\-\.]+)?(?:\+[0-9A-Za-z\-\.]+)?)$`)
)

// ParseVersion parses a version string into a Version object
func ParseVersion(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version string cannot be empty")
	}

	// Trim whitespace
	versionStr = strings.TrimSpace(versionStr)

	// Check for aliases first
	lowerVersion := strings.ToLower(versionStr)
	switch lowerVersion {
	case "latest":
		return &Version{
			Type:  VersionTypeAlias,
			Alias: VersionAliasLatest,
		}, nil
	case "lts":
		return &Version{
			Type:  VersionTypeAlias,
			Alias: VersionAliasLTS,
		}, nil
	case "stable":
		return &Version{
			Type:  VersionTypeAlias,
			Alias: VersionAliasStable,
		}, nil
	}

	// Check for caret range (^)
	if matches := caretRegex.FindStringSubmatch(versionStr); matches != nil {
		semver, err := ParseSemVer(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid caret range version: %w", err)
		}
		return &Version{
			Type:     VersionTypeRange,
			RangeOp:  RangeOperatorCaret,
			RangeVer: semver,
		}, nil
	}

	// Check for tilde range (~)
	if matches := tildeRegex.FindStringSubmatch(versionStr); matches != nil {
		// Try full semver parser first (handles prerelease and build metadata)
		semver, err := ParseSemVer(matches[1])
		if err == nil {
			// Full semver parse succeeded
			return &Version{
				Type:     VersionTypeRange,
				RangeOp:  RangeOperatorTilde,
				RangeVer: semver,
			}, nil
		}

		// Fall back to partial parser for cases like ~1 or ~1.2
		semver, err = ParseSemVerPartial(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid tilde range version: %w", err)
		}
		return &Version{
			Type:     VersionTypeRange,
			RangeOp:  RangeOperatorTilde,
			RangeVer: semver,
		}, nil
	}

	// Check for comparison range (>=, >, <=, <, =)
	if matches := rangeRegex.FindStringSubmatch(versionStr); matches != nil {
		semver, err := ParseSemVer(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid range version: %w", err)
		}
		return &Version{
			Type:     VersionTypeRange,
			RangeOp:  RangeOperator(matches[1]),
			RangeVer: semver,
		}, nil
	}

	// Try to parse as exact semver
	semver, err := ParseSemVer(versionStr)
	if err != nil {
		// Fallback to partial semver for robustness (e.g. "18.2" or "1")
		semver, err = ParseSemVerPartial(versionStr)
		if err != nil {
			return nil, fmt.Errorf("invalid version string '%s': must be a valid semver (e.g., 1.2.3), partial semver (e.g., 1.2), range (e.g., >=1.2.0, ^1.2.3, ~1.2.0), or alias (latest, lts, stable): %w", versionStr, err)
		}
	}

	return &Version{
		Type:  VersionTypeExact,
		Exact: semver,
	}, nil
}

// ParseSemVer parses a semantic version string into a SemVer struct
func ParseSemVer(versionStr string) (*SemVer, error) {
	// Remove optional 'v' prefix
	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimPrefix(versionStr, "V")

	matches := semverRegex.FindStringSubmatch("v" + versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid semver format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return &SemVer{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// ParseSemVerPartial parses a partial semantic version string (for tilde ranges)
// Supports: 1, 1.2, 1.2.3, 1.2.3-alpha, etc.
func ParseSemVerPartial(versionStr string) (*SemVer, error) {
	// Remove optional 'v' prefix
	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimPrefix(versionStr, "V")

	// Split by dots
	parts := strings.Split(versionStr, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid partial semver format: %s", versionStr)
	}

	// Parse major
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	semver := &SemVer{
		Major: major,
		Minor: 0,
		Patch: 0,
	}

	// Parse minor if present
	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %w", err)
		}
		semver.Minor = minor
	}

	// Parse patch if present (may include prerelease and build)
	if len(parts) >= 3 {
		patchPart := parts[2]

		// Check for prerelease
		if idx := strings.Index(patchPart, "-"); idx != -1 {
			patch, err := strconv.Atoi(patchPart[:idx])
			if err != nil {
				return nil, fmt.Errorf("invalid patch version: %w", err)
			}
			semver.Patch = patch

			// Extract prerelease and build
			remainder := patchPart[idx+1:]
			if buildIdx := strings.Index(remainder, "+"); buildIdx != -1 {
				semver.Prerelease = remainder[:buildIdx]
				semver.Build = remainder[buildIdx+1:]
			} else {
				semver.Prerelease = remainder
			}
		} else if idx := strings.Index(patchPart, "+"); idx != -1 {
			// Check for build metadata
			patch, err := strconv.Atoi(patchPart[:idx])
			if err != nil {
				return nil, fmt.Errorf("invalid patch version: %w", err)
			}
			semver.Patch = patch
			semver.Build = patchPart[idx+1:]
		} else {
			// Just patch number
			patch, err := strconv.Atoi(patchPart)
			if err != nil {
				return nil, fmt.Errorf("invalid patch version: %w", err)
			}
			semver.Patch = patch
		}
	}

	return semver, nil
}

// FormatVersion formats a Version object back into a version string
func FormatVersion(v *Version) (string, error) {
	if v == nil {
		return "", fmt.Errorf("version cannot be nil")
	}

	switch v.Type {
	case VersionTypeExact:
		if v.Exact == nil {
			return "", fmt.Errorf("exact version is nil")
		}
		return formatSemVer(v.Exact), nil

	case VersionTypeRange:
		if v.RangeVer == nil {
			return "", fmt.Errorf("range version is nil")
		}
		switch v.RangeOp {
		case RangeOperatorCaret:
			return "^" + formatSemVer(v.RangeVer), nil
		case RangeOperatorTilde:
			return "~" + formatSemVer(v.RangeVer), nil
		case RangeOperatorGTE, RangeOperatorGT, RangeOperatorLTE, RangeOperatorLT, RangeOperatorEQ:
			return string(v.RangeOp) + formatSemVer(v.RangeVer), nil
		default:
			return "", fmt.Errorf("unknown range operator: %s", v.RangeOp)
		}

	case VersionTypeAlias:
		return string(v.Alias), nil

	default:
		return "", fmt.Errorf("unknown version type: %d", v.Type)
	}
}

// formatSemVer formats a SemVer struct into a version string
func formatSemVer(v *SemVer) string {
	if v == nil {
		return ""
	}

	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}

	if v.Build != "" {
		result += "+" + v.Build
	}

	return result
}

// String returns the string representation of a Version
func (v *Version) String() string {
	str, _ := FormatVersion(v)
	return str
}

// Equal checks if two Version objects are equivalent
func (v *Version) Equal(other *Version) bool {
	if v == nil || other == nil {
		return v == other
	}

	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case VersionTypeExact:
		return v.Exact.Equal(other.Exact)
	case VersionTypeRange:
		return v.RangeOp == other.RangeOp && v.RangeVer.Equal(other.RangeVer)
	case VersionTypeAlias:
		return v.Alias == other.Alias
	default:
		return false
	}
}

// Equal checks if two SemVer objects are equivalent
func (s *SemVer) Equal(other *SemVer) bool {
	if s == nil || other == nil {
		return s == other
	}

	return s.Major == other.Major &&
		s.Minor == other.Minor &&
		s.Patch == other.Patch &&
		s.Prerelease == other.Prerelease &&
		s.Build == other.Build
}

// Compare compares s to other.
// Returns:
//
//	-1 if s < other
//	 0 if s == other
//	 1 if s > other
func (s *SemVer) Compare(other *SemVer) int {
	if s == nil && other == nil {
		return 0
	}
	if s == nil {
		return -1
	}
	if other == nil {
		return 1
	}

	if s.Major != other.Major {
		if s.Major < other.Major {
			return -1
		}
		return 1
	}

	if s.Minor != other.Minor {
		if s.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if s.Patch != other.Patch {
		if s.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Simple prerelease comparison:
	// A release version is greater than a prerelease version.
	if s.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if s.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if s.Prerelease != "" && other.Prerelease != "" {
		if s.Prerelease < other.Prerelease {
			return -1
		}
		if s.Prerelease > other.Prerelease {
			return 1
		}
	}

	return 0
}
