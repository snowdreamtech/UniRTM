// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoBackend_Properties(t *testing.T) {
	b := NewGoBackend()
	if b.Name() != "go" {
		t.Errorf("expected go, got %s", b.Name())
	}
	deps := b.Dependencies()
	if len(deps) != 1 || deps[0] != "go" {
		t.Errorf("expected [go], got %v", deps)
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum true")
	}
	if b.SupportsGPG() {
		t.Error("expected SupportsGPG false")
	}
	if b.AttestationType() != "" {
		t.Errorf("expected empty attestation, got %s", b.AttestationType())
	}
	if !b.IsRecommended() || !b.IsScriptless() || b.GetReach() != "Huge" || !b.IsStable() || !b.SupportsOffline() {
		t.Error("properties mismatch")
	}
}

func TestGoBackend_GetGoProxyBase(t *testing.T) {
	t.Setenv("GOPROXY", "https://proxy.example.com,direct")
	base := getGoProxyBase()
	if base != "https://proxy.example.com" {
		t.Errorf("expected https://proxy.example.com, got %s", base)
	}

	t.Setenv("GOPROXY", "")
	base = getGoProxyBase()
	if base != "https://proxy.golang.org" {
		t.Errorf("expected https://proxy.golang.org, got %s", base)
	}
}

func TestGoBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tool/@v/list" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("v1.0.0\nv1.1.0\nv1.2.0"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	t.Setenv("GOPROXY", ts.URL)

	b := NewGoBackend()
	b.client.Transport = http.DefaultTransport

	ctx := context.Background()
	plat := Platform{OS: "linux", Arch: "amd64"}
	versions, err := b.ListVersions(ctx, "tool", plat)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// Note: sorted descending
	if versions[0].Version != "v1.2.0" {
		t.Errorf("expected v1.2.0, got %s", versions[0].Version)
	}

	// 404 test
	_, err = b.ListVersions(ctx, "not-found", plat)
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestGoBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tool/@v/list" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("v1.0.0\nv1.1.0\nv1.2.0"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	t.Setenv("GOPROXY", ts.URL)

	b := NewGoBackend()
	ctx := context.Background()
	plat := Platform{}

	info, err := b.ResolveVersion(ctx, "tool", "latest", plat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "v1.2.0" {
		t.Errorf("expected v1.2.0, got %s", info.Version)
	}

	info, err = b.ResolveVersion(ctx, "tool", "1.1.0", plat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", info.Version)
	}
}

func TestGoBackend_GetDownloadInfo(t *testing.T) {
	b := NewGoBackend()
	info, err := b.GetDownloadInfo(context.Background(), "tool", "1.0.0", Platform{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected 1.0.0")
	}
}
