// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os/exec"
	"testing"
)

func TestSpmBackend_Name(t *testing.T) {
	b := NewSpmBackend()
	if b.Name() != "spm" {
		t.Errorf("expected name 'spm', got %s", b.Name())
	}
}

func TestSpmBackend_Properties(t *testing.T) {
	b := NewSpmBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if b.SupportsGPG() {
		t.Error("expected SupportsGPG to be false")
	}
	if b.AttestationType() != "" {
		t.Errorf("expected empty string, got %s", b.AttestationType())
	}
	if !b.IsRecommended() {
		t.Error("expected IsRecommended to be true")
	}
	if !b.IsScriptless() {
		t.Error("expected IsScriptless to be true")
	}
	if b.GetReach() != "Large" {
		t.Errorf("expected Large, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestSpmBackend_ListVersions(t *testing.T) {
	// Create a dummy local git repository with tags
	tmpDir := t.TempDir()

	// Create git repo
	if out, err := exec.Command("git", "-C", tmpDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").CombinedOutput(); err != nil {
		t.Fatalf("git config email failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").CombinedOutput(); err != nil {
		t.Fatalf("git config name failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-m", "init").CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "tag", "v1.0.0").CombinedOutput(); err != nil {
		t.Fatalf("git tag v1.0.0 failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "tag", "v1.1.0").CombinedOutput(); err != nil {
		t.Fatalf("git tag v1.1.0 failed: %v, output: %s", err, string(out))
	}

	b := NewSpmBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Note: use local file URL
	repoURL := "file://" + tmpDir

	versions, err := b.ListVersions(ctx, repoURL, platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	// "git ls-remote --tags" output might be sorted alphabetically
	// v1.0.0, v1.1.0
	found100 := false
	found110 := false
	for _, v := range versions {
		if v.Version == "v1.0.0" {
			found100 = true
		}
		if v.Version == "v1.1.0" {
			found110 = true
		}
	}
	if !found100 || !found110 {
		t.Errorf("expected tags v1.0.0 and v1.1.0, got %v", versions)
	}
}

func TestSpmBackend_ResolveVersion(t *testing.T) {
	b := NewSpmBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	// Test latest with fake repo
	tmpDir := t.TempDir()
	if out, err := exec.Command("git", "-C", tmpDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").CombinedOutput(); err != nil {
		t.Fatalf("git config email failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").CombinedOutput(); err != nil {
		t.Fatalf("git config name failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-m", "init").CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v, output: %s", err, string(out))
	}
	if out, err := exec.Command("git", "-C", tmpDir, "tag", "v2.0.0").CombinedOutput(); err != nil {
		t.Fatalf("git tag failed: %v, output: %s", err, string(out))
	}

	repoURL := "file://" + tmpDir

	vLatest, err := b.ResolveVersion(ctx, repoURL, "latest", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vLatest.Version != "v2.0.0" {
		t.Errorf("expected v2.0.0, got %s", vLatest.Version)
	}

	// Test specific
	info, err := b.ResolveVersion(ctx, "apple/swift-format", "509.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "509.0.0" {
		t.Errorf("expected version '509.0.0', got %s", info.Version)
	}
}

func TestSpmBackend_GetDownloadInfo(t *testing.T) {
	b := NewSpmBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "apple/swift-format", "509.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "509.0.0" {
		t.Errorf("expected 509.0.0, got %s", info.Version)
	}
}
