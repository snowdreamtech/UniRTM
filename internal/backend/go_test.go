// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGoBackend_Name(t *testing.T) {
	b := NewGoBackend()
	if b.Name() != "go" {
		t.Errorf("expected go, got %s", b.Name())
	}
}

func TestGoBackend_Properties(t *testing.T) {
	b := NewGoBackend()

	deps := b.Dependencies()
	if len(deps) != 1 || deps[0] != "go" {
		t.Errorf("expected [go], got %v", deps)
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
	if b.GetReach() != "Huge" {
		t.Errorf("expected Huge, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestGetGoProxyBase(t *testing.T) {
	original := os.Getenv("GOPROXY")
	defer os.Setenv("GOPROXY", original)

	os.Setenv("GOPROXY", "")
	if getGoProxyBase() != "https://proxy.golang.org" {
		t.Errorf("expected default proxy, got %s", getGoProxyBase())
	}

	os.Setenv("GOPROXY", "direct")
	if getGoProxyBase() != "https://proxy.golang.org" {
		t.Errorf("expected fallback proxy, got %s", getGoProxyBase())
	}

	os.Setenv("GOPROXY", "https://goproxy.cn,direct")
	if getGoProxyBase() != "https://goproxy.cn" {
		t.Errorf("expected goproxy.cn, got %s", getGoProxyBase())
	}
}

func TestGoBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/golang.org/x/tools/@v/list" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("v0.1.0\nv0.1.1\nv0.2.0\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Intercept transport to redirect to test server
	b := NewGoBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "golang.org/x/tools", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// Sorted descending
	if versions[0].Version != "v0.2.0" {
		t.Errorf("expected v0.2.0, got %s", versions[0].Version)
	}

	// Test not found
	_, err = b.ListVersions(ctx, "nonexistent", platform)
	if err == nil {
		t.Error("expected error for nonexistent package, got nil")
	}
}

func TestGoBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("v1.0.0\nv2.0.0\n"))
	}))
	defer ts.Close()

	b := NewGoBackend()
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
	if v.Version != "v2.0.0" {
		t.Errorf("expected v2.0.0, got %s", v.Version)
	}

	v2, err := b.ResolveVersion(ctx, "dummy", "v1.5.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.Version != "v1.5.0" {
		t.Errorf("expected v1.5.0, got %s", v2.Version)
	}
}

func TestGoBackend_GetDownloadInfo(t *testing.T) {
	b := NewGoBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "golang.org/x/tools", "v0.2.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "v0.2.0" {
		t.Errorf("expected v0.2.0, got %s", info.Version)
	}
}
