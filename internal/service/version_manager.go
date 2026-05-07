// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/backend"
)

// VersionManager handles version constraint parsing and resolution.
// It enforces explicit version requirements and delegates to backends for resolution.
type VersionManager struct {
	backends map[string]backend.Backend
}

// NewVersionManager creates a new VersionManager with the given backends.
func NewVersionManager(backends map[string]backend.Backend) *VersionManager {
	return &VersionManager{
		backends: backends,
	}
}

// ParseVersionConstraint parses a version constraint string into a Version object.
// This is a convenience wrapper around ParseVersion that enforces explicit version requirements.
//
// Returns an error if:
// - The version string is empty (explicit version required)
// - The version string is invalid
func (vm *VersionManager) ParseVersionConstraint(versionStr string) (*Version, error) {
	if versionStr == "" {
		return nil, fmt.Errorf("version specification is required: must be an exact version (e.g., 1.20.0), range (e.g., >=1.20.0, ^3.11, ~2.7.0), or alias (latest, lts, stable)")
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("parse version constraint: %w", err)
	}

	return version, nil
}

// ResolveVersion resolves a version specification to a concrete version using the specified backend.
//
// Parameters:
// - ctx: Context for cancellation and timeouts
// - backendName: Name of the backend to use for resolution (e.g., "github", "aqua", "http")
// - tool: Name of the tool to resolve the version for
// - versionSpec: Version specification (exact version, range, or alias)
// - platform: Target platform for the resolution
//
// Returns:
// - The resolved VersionInfo with concrete version and download information
// - An error if resolution fails
//
// Resolution behavior:
// - Exact versions (e.g., "1.20.0"): Delegates to backend.GetDownloadInfo
// - Aliases (e.g., "latest", "lts", "stable"): Delegates to backend.ResolveVersion
// - Ranges (e.g., ">=1.20.0", "^3.11", "~2.7.0"): Delegates to backend.ResolveVersion
func (vm *VersionManager) ResolveVersion(ctx context.Context, backendName, tool, versionSpec string, platform backend.Platform) (*backend.VersionInfo, error) {
	// Enforce explicit version requirement
	if versionSpec == "" {
		return nil, fmt.Errorf("explicit version specification required for tool '%s': must specify an exact version (e.g., 1.20.0), range (e.g., >=1.20.0, ^3.11, ~2.7.0), or alias (latest, lts, stable)", tool)
	}

	// Parse the version specification
	version, err := vm.ParseVersionConstraint(versionSpec)
	if err != nil {
		return nil, fmt.Errorf("resolve version for tool '%s': %w", tool, err)
	}

	// Get the backend
	b, ok := vm.backends[backendName]
	if !ok {
		return nil, fmt.Errorf("backend '%s' not found for tool '%s'", backendName, tool)
	}

	// Resolve based on version type
	switch version.Type {
	case VersionTypeExact:
		// For exact versions, get download info directly
		versionInfo, err := b.GetDownloadInfo(ctx, tool, version.String(), platform)
		if err != nil {
			return nil, fmt.Errorf("get download info for tool '%s' version '%s': %w", tool, version.String(), err)
		}
		return versionInfo, nil

	case VersionTypeAlias, VersionTypeRange:
		// For aliases and ranges, delegate to backend resolution
		versionInfo, err := b.ResolveVersion(ctx, tool, versionSpec, platform)
		if err != nil {
			return nil, fmt.Errorf("resolve version '%s' for tool '%s': %w", versionSpec, tool, err)
		}
		return versionInfo, nil

	default:
		return nil, fmt.Errorf("unknown version type %d for tool '%s'", version.Type, tool)
	}
}

// ValidateVersionConstraint validates a version constraint string without resolving it.
// This is useful for configuration validation.
//
// Returns an error if the version constraint is invalid.
func (vm *VersionManager) ValidateVersionConstraint(versionStr string) error {
	if versionStr == "" {
		return fmt.Errorf("version specification cannot be empty")
	}

	_, err := ParseVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version constraint: %w", err)
	}

	return nil
}

// ListAvailableVersions lists all available versions for a tool from the specified backend.
// This is useful for displaying available versions to users.
//
// Returns a list of VersionInfo objects in descending order (newest first).
func (vm *VersionManager) ListAvailableVersions(ctx context.Context, backendName, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	b, ok := vm.backends[backendName]
	if !ok {
		return nil, fmt.Errorf("backend '%s' not found", backendName)
	}

	versions, err := b.ListVersions(ctx, tool, platform)
	if err != nil {
		return nil, fmt.Errorf("list versions for tool '%s': %w", tool, err)
	}

	return versions, nil
}

// SupportsChecksum checks if the specified backend supports checksum verification.
func (vm *VersionManager) SupportsChecksum(backendName string) (bool, error) {
	b, ok := vm.backends[backendName]
	if !ok {
		return false, fmt.Errorf("backend '%s' not found", backendName)
	}

	return b.SupportsChecksum(), nil
}

// SupportsGPG checks if the specified backend supports GPG signature verification.
func (vm *VersionManager) SupportsGPG(backendName string) (bool, error) {
	b, ok := vm.backends[backendName]
	if !ok {
		return false, fmt.Errorf("backend '%s' not found", backendName)
	}

	return b.SupportsGPG(), nil
}
