// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDenoBackend_Name(t *testing.T) {
	b := NewDenoBackend()
	if b.Name() != "deno" {
		t.Errorf("expected deno, got %s", b.Name())
	}
}

func TestDenoBackend_Properties(t *testing.T) {
	b := NewDenoBackend()

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

func TestDenoBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/std/meta/versions.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"latest": "0.200.0",
				"versions": [
					"0.200.0",
					"0.199.0",
					"0.198.0"
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewDenoBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "std", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].Version != "0.200.0" {
		t.Errorf("expected 0.200.0, got %s", versions[0].Version)
	}

	// Test not found
	_, err = b.ListVersions(ctx, "nonexistent", platform)
	if err == nil {
		t.Error("expected error for nonexistent package, got nil")
	}
}

func TestDenoBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"latest": "1.5.0",
			"versions": ["1.5.0", "1.4.0"]
		}`))
	}))
	defer ts.Close()

	b := NewDenoBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "dummy", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", v.Version)
	}

	v2, err := b.ResolveVersion(ctx, "dummy", "0.9.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.Version != "0.9.0" {
		t.Errorf("expected 0.9.0, got %s", v2.Version)
	}
}

func TestDenoBackend_GetDownloadInfo(t *testing.T) {
	b := NewDenoBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "std", "0.200.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "0.200.0" {
		t.Errorf("expected 0.200.0, got %s", info.Version)
	}
}
