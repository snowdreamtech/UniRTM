// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestZigBackend_Name(t *testing.T) {
	b := NewZigBackend()
	if b.Name() != "zig" {
		t.Errorf("expected zig, got %s", b.Name())
	}
}

func TestZigBackend_Properties(t *testing.T) {
	b := NewZigBackend()

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

func TestZigBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/index.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"master": {},
				"0.11.0": {},
				"0.10.1": {}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewZigBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "zig", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
}

func TestZigBackend_ResolveVersion(t *testing.T) {
	b := NewZigBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "zig", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "master" {
		t.Errorf("expected master, got %s", v.Version)
	}

	v2, err := b.ResolveVersion(ctx, "zig", "0.11.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.Version != "0.11.0" {
		t.Errorf("expected 0.11.0, got %s", v2.Version)
	}
}

func TestZigBackend_GetDownloadInfo(t *testing.T) {
	b := NewZigBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "zig", "0.11.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "0.11.0" {
		t.Errorf("expected 0.11.0, got %s", info.Version)
	}
}
