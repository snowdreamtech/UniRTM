// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service_test

import (
	"context"
	"fmt"
	"log"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// mockBackendForExample is a simple mock backend for examples
type mockBackendForExample struct {
	name string
}

func (m *mockBackendForExample) Name() string {
	return m.name
}

func (m *mockBackendForExample) AttestationType() string {
	return ""
}

func (m *mockBackendForExample) Dependencies() []string {
	return nil
}

func (m *mockBackendForExample) GetReach() string {
	return ""
}

func (m *mockBackendForExample) IsRecommended() bool {
	return false
}

func (m *mockBackendForExample) IsScriptless() bool {
	return false
}

func (m *mockBackendForExample) IsStable() bool {
	return true
}

func (m *mockBackendForExample) SupportsOffline() bool {
	return false
}

func (m *mockBackendForExample) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	// Return mock version list
	return []backend.VersionInfo{
		{Version: "20.11.0", DownloadURL: "https://example.com/node-20.11.0.tar.gz"},
		{Version: "20.10.0", DownloadURL: "https://example.com/node-20.10.0.tar.gz"},
		{Version: "18.19.0", DownloadURL: "https://example.com/node-18.19.0.tar.gz"},
	}, nil
}

func (m *mockBackendForExample) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	// Simulate resolving "latest" to the newest version
	if versionRequest == "latest" {
		return &backend.VersionInfo{
			Version:     "20.11.0",
			DownloadURL: "https://example.com/node-20.11.0.tar.gz",
			Checksum:    "abc123def456",
			Platform:    platform,
		}, nil
	}

	// Simulate resolving "^20.0.0" to the latest 20.x version
	if versionRequest == "^20.0.0" {
		return &backend.VersionInfo{
			Version:     "20.11.0",
			DownloadURL: "https://example.com/node-20.11.0.tar.gz",
			Checksum:    "abc123def456",
			Platform:    platform,
		}, nil
	}

	return nil, fmt.Errorf("unsupported version request: %s", versionRequest)
}

func (m *mockBackendForExample) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	// Return download info for the exact version
	return &backend.VersionInfo{
		Version:     version,
		DownloadURL: fmt.Sprintf("https://example.com/%s-%s.tar.gz", tool, version),
		Checksum:    "abc123def456",
		Platform:    platform,
	}, nil
}

func (m *mockBackendForExample) SupportsChecksum() bool {
	return true
}

func (m *mockBackendForExample) SupportsGPG() bool {
	return false
}

// Example_versionManager_basicUsage demonstrates basic Version Manager usage
func Example_versionManager_basicUsage() {
	// Create backends
	backends := map[string]backend.Backend{
		"github": &mockBackendForExample{name: "github"},
	}

	// Create Version Manager
	vm := service.NewVersionManager(backends)

	// Resolve exact version
	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "20.0.0", platform)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Version: %s\n", versionInfo.Version)
	fmt.Printf("Download URL: %s\n", versionInfo.DownloadURL)

	// Output:
	// Version: 20.0.0
	// Download URL: https://example.com/node-20.0.0.tar.gz
}

// Example_versionManager_resolveAlias demonstrates resolving version aliases
func Example_versionManager_resolveAlias() {
	backends := map[string]backend.Backend{
		"github": &mockBackendForExample{name: "github"},
	}

	vm := service.NewVersionManager(backends)

	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	// Resolve "latest" alias
	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "latest", platform)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Latest version: %s\n", versionInfo.Version)

	// Output:
	// Latest version: 20.11.0
}

// Example_versionManager_resolveRange demonstrates resolving version ranges
func Example_versionManager_resolveRange() {
	backends := map[string]backend.Backend{
		"github": &mockBackendForExample{name: "github"},
	}

	vm := service.NewVersionManager(backends)

	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	// Resolve caret range
	versionInfo, err := vm.ResolveVersion(ctx, "github", "node", "^20.0.0", platform)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Resolved ^20.0.0 to: %s\n", versionInfo.Version)

	// Output:
	// Resolved ^20.0.0 to: 20.11.0
}

// Example_versionManager_validateConstraint demonstrates version constraint validation
func Example_versionManager_validateConstraint() {
	vm := service.NewVersionManager(nil)

	// Validate various version constraints
	constraints := []string{
		"20.0.0",
		"^20.0.0",
		">=18.0.0",
		"latest",
	}

	for _, constraint := range constraints {
		if err := vm.ValidateVersionConstraint(constraint); err != nil {
			fmt.Printf("Invalid: %s - %v\n", constraint, err)
		} else {
			fmt.Printf("Valid: %s\n", constraint)
		}
	}

	// Output:
	// Valid: 20.0.0
	// Valid: ^20.0.0
	// Valid: >=18.0.0
	// Valid: latest
}

// Example_versionManager_listVersions demonstrates listing available versions
func Example_versionManager_listVersions() {
	backends := map[string]backend.Backend{
		"github": &mockBackendForExample{name: "github"},
	}

	vm := service.NewVersionManager(backends)

	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	// List available versions
	versions, err := vm.ListAvailableVersions(ctx, "github", "node", platform)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available versions:")
	for _, v := range versions {
		fmt.Printf("  - %s\n", v.Version)
	}

	// Output:
	// Available versions:
	//   - 20.11.0
	//   - 20.10.0
	//   - 18.19.0
}

// Example_versionManager_explicitRequirement demonstrates explicit version requirement enforcement
func Example_versionManager_explicitRequirement() {
	backends := map[string]backend.Backend{
		"github": &mockBackendForExample{name: "github"},
	}

	vm := service.NewVersionManager(backends)

	ctx := context.Background()
	platform := backend.Platform{OS: "linux", Arch: "amd64"}

	// Try to resolve without specifying a version
	_, err := vm.ResolveVersion(ctx, "github", "node", "", platform)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Output:
	// Error: explicit version specification required for tool 'node': must specify an exact version (e.g., 1.20.0), range (e.g., >=1.20.0, ^3.11, ~2.7.0), or alias (latest, lts, stable)
}
