// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"os"
	"path/filepath"
	"testing"
)

// resetRegistry resets the sync.Once so tests are independent.
func resetRegistry() {
	InvalidateEffectiveRegistry()
}

func TestLoadLocalAssetRegistry_FileNotExist(t *testing.T) {
	result, err := LoadLocalAssetRegistry("/nonexistent/path/asset-registry.toml")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result for missing file, got %v", result)
	}
}

func TestLoadLocalAssetRegistry_ValidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "asset-registry.toml")

	// Write valid TOML content without shell-unsafe characters in the go string.
	content := "[tools]\n\n" +
		"[tools.\"my-org/my-tool\"]\n" +
		"description = \"Custom tool\"\n\n" +
		"[tools.\"my-org/my-tool\".patterns]\n" +
		"\"linux-amd64\" = \"my-tool-linux-x64\"\n" +
		"\"macos-arm64\" = \"my-tool-macos-arm64\"\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadLocalAssetRegistry(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	entry, ok := result["my-org/my-tool"]
	if !ok {
		t.Fatal("expected entry for my-org/my-tool")
	}
	if entry.Description != "Custom tool" {
		t.Errorf("expected description %q, got %q", "Custom tool", entry.Description)
	}
	if entry.Patterns["linux-amd64"] != "my-tool-linux-x64" {
		t.Errorf("expected linux-amd64 pattern my-tool-linux-x64, got %q", entry.Patterns["linux-amd64"])
	}
	if entry.Patterns["macos-arm64"] != "my-tool-macos-arm64" {
		t.Errorf("expected macos-arm64 pattern my-tool-macos-arm64, got %q", entry.Patterns["macos-arm64"])
	}
}

func TestLoadLocalAssetRegistry_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "asset-registry.toml")
	if err := os.WriteFile(path, []byte("this is not valid toml ["), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadLocalAssetRegistry(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestMergeAssetRegistries_LocalOverridesBuiltin(t *testing.T) {
	builtin := map[string]AssetRegistryEntry{
		"tool-a": {Description: "builtin-a", Patterns: map[string]string{"linux-amd64": "builtin.tar.gz"}},
		"tool-b": {Description: "builtin-b"},
	}
	local := map[string]AssetRegistryEntry{
		"tool-a": {Description: "local-a", Patterns: map[string]string{"linux-amd64": "local.tar.gz"}},
		"tool-c": {Description: "local-c"},
	}

	merged := MergeAssetRegistries(local, builtin)

	// local overrides builtin for tool-a
	if merged["tool-a"].Description != "local-a" {
		t.Errorf("expected local-a, got %q", merged["tool-a"].Description)
	}
	if merged["tool-a"].Patterns["linux-amd64"] != "local.tar.gz" {
		t.Errorf("expected local.tar.gz, got %q", merged["tool-a"].Patterns["linux-amd64"])
	}
	// builtin-only tool-b is preserved
	if merged["tool-b"].Description != "builtin-b" {
		t.Errorf("expected builtin-b, got %q", merged["tool-b"].Description)
	}
	// local-only tool-c is present
	if merged["tool-c"].Description != "local-c" {
		t.Errorf("expected local-c, got %q", merged["tool-c"].Description)
	}
	if len(merged) != 3 {
		t.Errorf("expected 3 entries, got %d", len(merged))
	}
}

func TestMergeAssetRegistries_NilLocal(t *testing.T) {
	builtin := map[string]AssetRegistryEntry{
		"tool-a": {Description: "builtin-a"},
	}
	merged := MergeAssetRegistries(nil, builtin)
	if merged["tool-a"].Description != "builtin-a" {
		t.Errorf("expected builtin-a, got %q", merged["tool-a"].Description)
	}
}

func TestMergeAssetRegistries_NilBuiltin(t *testing.T) {
	local := map[string]AssetRegistryEntry{
		"tool-a": {Description: "local-a"},
	}
	merged := MergeAssetRegistries(local, nil)
	if merged["tool-a"].Description != "local-a" {
		t.Errorf("expected local-a, got %q", merged["tool-a"].Description)
	}
}

func TestLookupRegistryPatterns_BuiltinFound(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	entry, ok := LookupRegistryPatterns("editorconfig-checker/editorconfig-checker")
	if !ok {
		t.Fatal("expected built-in entry for editorconfig-checker")
	}
	if entry.Patterns["macos-amd64"] != "editorconfig-checker-darwin-all.tar.gz" {
		t.Errorf("unexpected pattern: %q", entry.Patterns["macos-amd64"])
	}
}

func TestLookupRegistryPatterns_PrefixStripping(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	// Should match with "github:" prefix stripped.
	entry, ok := LookupRegistryPatterns("github:editorconfig-checker/editorconfig-checker")
	if !ok {
		t.Fatal("expected entry via github: prefix stripping")
	}
	if entry.Patterns["macos-arm64"] != "editorconfig-checker-darwin-all.tar.gz" {
		t.Errorf("unexpected pattern: %q", entry.Patterns["macos-arm64"])
	}
}

func TestLookupRegistryPatterns_NotFound(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	_, ok := LookupRegistryPatterns("nonexistent/tool")
	if ok {
		t.Fatal("expected not found for unknown tool")
	}
}

func TestEffectiveAssetRegistry_LocalFileOverridesBuiltin(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	// Directly test MergeAssetRegistries with a local override to verify
	// that local entries take precedence over built-in ones.
	local := map[string]AssetRegistryEntry{
		"editorconfig-checker/editorconfig-checker": {
			Description: "overridden by local",
			Patterns:    map[string]string{"macos-amd64": "overridden-asset.tar.gz"},
		},
	}
	merged := MergeAssetRegistries(local, AssetRegistry)

	entry := merged["editorconfig-checker/editorconfig-checker"]
	if entry.Patterns["macos-amd64"] != "overridden-asset.tar.gz" {
		t.Errorf("expected overridden-asset.tar.gz, got %q", entry.Patterns["macos-amd64"])
	}
	if entry.Description != "overridden by local" {
		t.Errorf("expected overridden description, got %q", entry.Description)
	}
}

func TestFindBestAssetWithOverride_ExactMatch(t *testing.T) {
	assets := []CommonAsset{
		{Name: "tool-darwin-all.tar.gz", URL: "https://example.com/darwin-all"},
		{Name: "tool-linux-amd64.tar.gz", URL: "https://example.com/linux-amd64"},
	}
	patterns := map[string]string{
		"macos-amd64": "tool-darwin-all.tar.gz",
		"macos-arm64": "tool-darwin-all.tar.gz",
	}
	platform := Platform{OS: "darwin", Arch: "amd64"}

	asset, score := FindBestAssetWithOverride(assets, platform, "my-tool", "macos-amd64", patterns)
	if asset == nil {
		t.Fatal("expected asset, got nil")
	}
	if asset.Name != "tool-darwin-all.tar.gz" {
		t.Errorf("expected tool-darwin-all.tar.gz, got %q", asset.Name)
	}
	if score != 1000 {
		t.Errorf("expected override score 1000, got %d", score)
	}
}

func TestFindBestAssetWithOverride_FallsBackToScoring(t *testing.T) {
	assets := []CommonAsset{
		{Name: "tool-linux-amd64.tar.gz", URL: "https://example.com/linux-amd64"},
	}
	platform := Platform{OS: "linux", Arch: "amd64"}

	// No patterns — should fall back to heuristic scoring.
	asset, _ := FindBestAssetWithOverride(assets, platform, "tool", "linux-amd64", nil)
	if asset == nil {
		t.Fatal("expected asset via heuristic fallback, got nil")
	}
	if asset.Name != "tool-linux-amd64.tar.gz" {
		t.Errorf("unexpected asset: %q", asset.Name)
	}
}

func TestFindBestAssetWithOverride_OverrideNotInRelease_FallsBack(t *testing.T) {
	assets := []CommonAsset{
		{Name: "tool-linux-amd64.tar.gz", URL: "https://example.com/linux-amd64"},
	}
	patterns := map[string]string{
		"linux-amd64": "nonexistent-asset.tar.gz", // specified but not in release
	}
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Override name not found in release; should fall back to heuristic.
	asset, _ := FindBestAssetWithOverride(assets, platform, "tool", "linux-amd64", patterns)
	if asset == nil {
		t.Fatal("expected heuristic fallback asset, got nil")
	}
}
