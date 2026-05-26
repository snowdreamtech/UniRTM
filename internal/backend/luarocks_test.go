// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestLuarocksBackend_Name(t *testing.T) {
	b := NewLuarocksBackend()
	if b.Name() != "luarocks" {
		t.Errorf("expected luarocks, got %s", b.Name())
	}
}

func TestLuarocksBackend_Properties(t *testing.T) {
	b := NewLuarocksBackend()

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

func TestLuarocksBackend_ListVersions(t *testing.T) {
	b := NewLuarocksBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	_, err := b.ListVersions(ctx, "lapis", platform)
	if err == nil {
		t.Error("expected error for ListVersions as it is not implemented, got nil")
	}
}

func TestLuarocksBackend_ResolveVersion(t *testing.T) {
	b := NewLuarocksBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "lapis", "1.9.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "1.9.0" {
		t.Errorf("expected 1.9.0, got %s", v.Version)
	}
}

func TestLuarocksBackend_GetDownloadInfo(t *testing.T) {
	b := NewLuarocksBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "lapis", "1.9.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.9.0" {
		t.Errorf("expected 1.9.0, got %s", info.Version)
	}
}
