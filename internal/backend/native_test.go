// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
)

type mockRecipeHandler struct{}

func (m *mockRecipeHandler) Name() string { return "mock" }
func (m *mockRecipeHandler) ResolveVersions(ctx context.Context, baseURL string) ([]native.VersionInfo, error) {
	return []native.VersionInfo{
		{
			Version: "1.0.0",
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", Filename: "tool-linux-amd64"},
			},
		},
		{
			Version: "1.1.0",
			Assets: []native.Asset{
				{OS: "linux", Arch: "amd64", Filename: "tool-linux-amd64"},
			},
		},
	}, nil
}

func (m *mockRecipeHandler) BuildURL(version, os, arch, baseURL string) string {
	return "http://test/" + version
}
func (m *mockRecipeHandler) SupportedOS() []string   { return []string{"linux"} }
func (m *mockRecipeHandler) SupportedArch() []string { return []string{"amd64"} }

func TestNativeBackend_ResolveVersion(t *testing.T) {
	b := NewNativeBackend()
	// Inject a mock recipe
	b.recipes["test-tool"] = native.Recipe{
		BaseURL: "http://test",
		Handler: &mockRecipeHandler{},
		Aliases: map[string]string{
			"stable": "1.1.0",
		},
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test alias
	vi, err := b.ResolveVersion(ctx, "test-tool", "stable", platform)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if vi.Version != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", vi.Version)
	}

	// Test latest
	vi, err = b.ResolveVersion(ctx, "test-tool", "latest", platform)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if vi.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", vi.Version)
	} // first in array is 1.0.0

	// Test exact version
	vi, err = b.ResolveVersion(ctx, "test-tool", "1.0.0", platform)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if vi.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", vi.Version)
	}

	// Test GetDownloadInfo
	_, err = b.GetDownloadInfo(ctx, "test-tool", "1.0.0", platform)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	// Test ListVersions
	versions, err := b.ListVersions(ctx, "test-tool", platform)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestNativeBackend_MuslFallback(t *testing.T) {
	b := NewNativeBackend()
	b.recipes["test-tool"] = native.Recipe{
		BaseURL: "http://test",
		Handler: &mockRecipeHandler{}, // Returns asset "tool-linux-amd64" (glibc)
	}

	ctx := context.Background()
	// Simulate running on a Musl platform (e.g. Alpine)
	platform := Platform{OS: "linux", Arch: "amd64", Musl: true}

	vi, err := b.GetDownloadInfo(ctx, "test-tool", "1.0.0", platform)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if vi.Metadata == nil || vi.Metadata["IsGlibcFallback"] != "true" {
		t.Errorf("expected IsGlibcFallback to be 'true', got %v", vi.Metadata["IsGlibcFallback"])
	}
}

type mockDebugRecipeHandler struct{}

func (m *mockDebugRecipeHandler) Name() string { return "mock_debug" }
func (m *mockDebugRecipeHandler) ResolveVersions(ctx context.Context, baseURL string) ([]native.VersionInfo, error) {
	return []native.VersionInfo{
		{
			Version: "2.0.0",
			Assets: []native.Asset{
				{OS: "windows", Arch: "amd64", Filename: "tool-2.0.0-windows-amd64-debug.zip", URL: "http://test/debug.zip"},
				{OS: "windows", Arch: "amd64", Filename: "tool-2.0.0-windows-amd64.zip", URL: "http://test/release.zip"},
			},
		},
	}, nil
}
func (m *mockDebugRecipeHandler) BuildURL(version, os, arch, baseURL string) string {
	return "http://test/" + version
}
func (m *mockDebugRecipeHandler) SupportedOS() []string   { return []string{"windows"} }
func (m *mockDebugRecipeHandler) SupportedArch() []string { return []string{"amd64"} }

func TestNativeBackend_DebugAssetRejection(t *testing.T) {
	b := NewNativeBackend()
	b.recipes["mytool"] = native.Recipe{
		BaseURL: "http://test",
		Handler: &mockDebugRecipeHandler{},
	}

	ctx := context.Background()
	platform := Platform{OS: "windows", Arch: "amd64"}

	vi, err := b.GetDownloadInfo(ctx, "mytool", "2.0.0", platform)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if vi.DownloadURL != "http://test/release.zip" {
		t.Errorf("expected release URL, got %s", vi.DownloadURL)
	}
}
