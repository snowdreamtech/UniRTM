// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestNativeBackend_Name(t *testing.T) {
	b := NewNativeBackend()
	if b.Name() != "native" {
		t.Errorf("expected native, got %s", b.Name())
	}
}

func TestNativeBackend_Properties(t *testing.T) {
	b := NewNativeBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if !b.SupportsGPG() {
		t.Error("expected SupportsGPG to be true")
	}
	if b.AttestationType() != "Native" {
		t.Errorf("expected Native, got %s", b.AttestationType())
	}
	if !b.IsRecommended() {
		t.Error("expected IsRecommended to be true")
	}
	if !b.IsScriptless() {
		t.Error("expected IsScriptless to be true")
	}
	if b.GetReach() != "Small" {
		t.Errorf("expected Small, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestNativeBackend_ListVersions(t *testing.T) {
	b := NewNativeBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test missing recipe
	_, err := b.ListVersions(ctx, "nonexistent-tool", platform)
	if err == nil {
		t.Error("expected error for nonexistent tool, got nil")
	}
}

func TestNativeBackend_ResolveVersion(t *testing.T) {
	b := NewNativeBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test missing recipe
	_, err := b.ResolveVersion(ctx, "nonexistent-tool", "latest", platform)
	if err == nil {
		t.Error("expected error for nonexistent tool, got nil")
	}
}

func TestNativeBackend_GetDownloadInfo(t *testing.T) {
	b := NewNativeBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test missing recipe
	_, err := b.GetDownloadInfo(ctx, "nonexistent-tool", "1.0", platform)
	if err == nil {
		t.Error("expected error for nonexistent tool, got nil")
	}
}
