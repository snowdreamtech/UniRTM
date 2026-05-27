// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestVfoxBackend_Name(t *testing.T) {
	b := NewVfoxBackend()
	if b.Name() != "vfox" {
		t.Errorf("expected name 'vfox', got %s", b.Name())
	}
}

func TestVfoxBackend_Properties(t *testing.T) {
	b := NewVfoxBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be false")
	}
	if b.SupportsGPG() {
		t.Error("expected SupportsGPG to be false")
	}
	if b.AttestationType() != "" {
		t.Errorf("expected empty string, got %s", b.AttestationType())
	}
	if b.IsRecommended() {
		t.Error("expected IsRecommended to be false")
	}
	if b.IsScriptless() {
		t.Error("expected IsScriptless to be false")
	}
	if b.GetReach() != "Huge" {
		t.Errorf("expected Huge, got %s", b.GetReach())
	}
	if b.IsStable() {
		t.Error("expected IsStable to be false")
	}
	if b.SupportsOffline() {
		t.Error("expected SupportsOffline to be false")
	}
}

func TestVfoxBackend_ListVersions(t *testing.T) {
	// Create a fake vfox executable
	tmpDir := t.TempDir()
	vfoxPath := filepath.Join(tmpDir, "vfox")
	if runtime.GOOS == "windows" {
		vfoxPath += ".bat"
		os.WriteFile(vfoxPath, []byte("@echo off\nif \"%1\"==\"list\" if \"%2\"==\"all\" (\n  echo Available versions\n  echo 20.0.1\n  echo 21.0.0\n  echo 21.0.1\n)\n"), 0755)
	} else {
		os.WriteFile(vfoxPath, []byte("#!/bin/sh\nif [ \"$1\" = \"list\" ] && [ \"$2\" = \"all\" ]; then\n  echo \"Available versions\"\n  echo \"20.0.1\"\n  echo \"21.0.0\"\n  echo \"21.0.1\"\nfi\n"), 0755)
	}

	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	b := NewVfoxBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "java", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].Version != "20.0.1" || versions[2].Version != "21.0.1" {
		t.Errorf("unexpected versions: %v", versions)
	}
}

func TestVfoxBackend_ResolveVersion(t *testing.T) {
	b := NewVfoxBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	// Create a fake vfox executable for testing latest
	tmpDir := t.TempDir()
	vfoxPath := filepath.Join(tmpDir, "vfox")
	if runtime.GOOS == "windows" {
		vfoxPath += ".bat"
		os.WriteFile(vfoxPath, []byte("@echo off\nif \"%1\"==\"list\" if \"%2\"==\"all\" (\n  echo Available versions\n  echo 20.0.1\n  echo 21.0.0\n  echo 22.0.1\n)\n"), 0755)
	} else {
		os.WriteFile(vfoxPath, []byte("#!/bin/sh\nif [ \"$1\" = \"list\" ] && [ \"$2\" = \"all\" ]; then\n  echo \"Available versions\"\n  echo \"20.0.1\"\n  echo \"21.0.0\"\n  echo \"22.0.1\"\nfi\n"), 0755)
	}

	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	vLatest, err := b.ResolveVersion(ctx, "java", "latest", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vLatest.Version != "22.0.1" {
		t.Errorf("expected 22.0.1, got %s", vLatest.Version)
	}

	// Test specific
	info, err := b.ResolveVersion(ctx, "java", "21.0.1", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "21.0.1" {
		t.Errorf("expected version '21.0.1', got %s", info.Version)
	}
}

func TestVfoxBackend_GetDownloadInfo(t *testing.T) {
	b := NewVfoxBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "java", "21.0.1", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "21.0.1" {
		t.Errorf("expected 21.0.1, got %s", info.Version)
	}
}
