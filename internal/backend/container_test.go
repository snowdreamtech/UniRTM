// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestContainerBackend_Name(t *testing.T) {
	b := NewContainerBackend("docker")
	if b.Name() != "docker" {
		t.Errorf("expected docker, got %s", b.Name())
	}
}

func TestContainerBackend_Dependencies(t *testing.T) {
	b := NewContainerBackend("container")
	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
}

func TestContainerBackend_ListVersions(t *testing.T) {
	b := NewContainerBackend("container")
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "ghcr.io/aquasec/trivy", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(versions) != 1 || versions[0].Version != "latest" {
		t.Errorf("expected [latest], got %v", versions)
	}
}

func TestContainerBackend_ResolveVersion(t *testing.T) {
	b := NewContainerBackend("container")
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test empty version
	v, err := b.ResolveVersion(ctx, "alpine", "", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "latest" {
		t.Errorf("expected version to resolve to latest, got %s", v.Version)
	}

	// Test specific version
	v, err = b.ResolveVersion(ctx, "alpine", "3.18", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "3.18" {
		t.Errorf("expected version 3.18, got %s", v.Version)
	}
}

func TestContainerBackend_GetDownloadInfo(t *testing.T) {
	b := NewContainerBackend("container")
	ctx := context.Background()
	platform := Platform{OS: "darwin", Arch: "arm64"}

	// Test with tag
	info, err := b.GetDownloadInfo(ctx, "ghcr.io/aquasec/trivy", "0.48.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.DownloadURL != "" {
		t.Errorf("expected empty DownloadURL, got %s", info.DownloadURL)
	}
	if info.Checksum != "" {
		t.Errorf("expected empty Checksum, got %s", info.Checksum)
	}
	if info.Metadata["image"] != "ghcr.io/aquasec/trivy" {
		t.Errorf("expected metadata image to be ghcr.io/aquasec/trivy, got %s", info.Metadata["image"])
	}

	// Test with digest (sha256)
	digest := "sha256:1234567890abcdef"
	infoDigest, err := b.GetDownloadInfo(ctx, "ubuntu", digest, platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if infoDigest.Checksum != digest {
		t.Errorf("expected Checksum to be %s, got %s", digest, infoDigest.Checksum)
	}
	if infoDigest.Metadata["tag"] != digest {
		t.Errorf("expected metadata tag to be %s, got %s", digest, infoDigest.Metadata["tag"])
	}
}

func TestContainerBackend_Properties(t *testing.T) {
	b := NewContainerBackend("container")

	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if b.SupportsGPG() {
		t.Error("expected SupportsGPG to be false")
	}
	if b.AttestationType() != "" {
		t.Errorf("expected empty AttestationType, got %s", b.AttestationType())
	}
	if !b.IsRecommended() {
		t.Error("expected IsRecommended to be true")
	}
	if b.IsScriptless() {
		t.Error("expected IsScriptless to be false")
	}
	if b.GetReach() != "Large" {
		t.Errorf("expected reach Large, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if b.SupportsOffline() {
		t.Error("expected SupportsOffline to be false")
	}
}
