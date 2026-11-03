// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestPipxBackend_Name(t *testing.T) {
	b := NewPipxBackend()
	if b.Name() != "pipx" {
		t.Errorf("expected pipx, got %s", b.Name())
	}
}

func TestPipxBackend_Properties(t *testing.T) {
	b := NewPipxBackend()

	deps := b.Dependencies()
	if len(deps) != 1 || deps[0] != "python" {
		t.Errorf("expected [python], got %v", deps)
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
	if b.SupportsOffline() {
		t.Error("expected SupportsOffline to be false")
	}
}

func TestPipxBackend_ResolveVersion(t *testing.T) {
	b := NewPipxBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Note: We bypass pypi API testing since pypi test handles it,
	// we just test exact version resolution which doesn't hit the network.
	v, err := b.ResolveVersion(ctx, "cowsay", "5.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "5.0" {
		t.Errorf("expected 5.0, got %s", v.Version)
	}
}

func TestPipxBackend_GetDownloadInfo(t *testing.T) {
	b := NewPipxBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "cowsay", "5.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "5.0" {
		t.Errorf("expected 5.0, got %s", info.Version)
	}
}
