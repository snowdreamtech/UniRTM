// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestCranBackend_Name(t *testing.T) {
	b := NewCranBackend()
	if b.Name() != "cran" {
		t.Errorf("expected cran, got %s", b.Name())
	}
}

func TestCranBackend_Properties(t *testing.T) {
	b := NewCranBackend()

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
	if b.GetReach() != "Medium" {
		t.Errorf("expected Medium, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestCranBackend_ListVersions(t *testing.T) {
	b := NewCranBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	_, err := b.ListVersions(ctx, "ggplot2", platform)
	if err == nil {
		t.Error("expected error for ListVersions as it is not implemented, got nil")
	}
}

func TestCranBackend_ResolveVersion(t *testing.T) {
	b := NewCranBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "ggplot2", "3.4.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "3.4.0" {
		t.Errorf("expected 3.4.0, got %s", v.Version)
	}
}

func TestCranBackend_GetDownloadInfo(t *testing.T) {
	b := NewCranBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "ggplot2", "3.4.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "3.4.0" {
		t.Errorf("expected 3.4.0, got %s", info.Version)
	}
}
