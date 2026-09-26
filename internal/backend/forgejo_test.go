// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForgejoBackend_Name(t *testing.T) {
	b := NewForgejoBackend()
	if b.Name() != "forgejo" {
		t.Errorf("expected name 'forgejo', got %s", b.Name())
	}
}

func TestForgejoBackend_ResolveVersion(t *testing.T) {
	ClearForgejoCache()
	defer ClearForgejoCache()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/releases" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"tag_name":"v1.0.0","assets":[{"name":"tool-linux-amd64.tar.gz","browser_download_url":"https://example.com/download"}]}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := &ForgejoBackend{
		client:  server.Client(),
		baseURL: server.URL,
	}

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "owner/repo", "latest", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", info.Version)
	}
}

func TestForgejoBackend_ListVersions(t *testing.T) {
	ClearForgejoCache()
	defer ClearForgejoCache()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"tag_name":"v2.0.0","assets":[{"name":"tool-linux-amd64.tar.gz","browser_download_url":"https://example.com/download2"}]},
			{"tag_name":"v1.0.0","assets":[{"name":"tool-linux-amd64.tar.gz","browser_download_url":"https://example.com/download"}]}
		]`))
	}))
	defer server.Close()

	b := &ForgejoBackend{
		client:  server.Client(),
		baseURL: server.URL,
	}

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "owner/repo", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", versions[0].Version)
	}
}

func TestForgejoBackend_Properties(t *testing.T) {
	b := NewForgejoBackend()
	if b.Dependencies() != nil {
		t.Errorf("expected nil dependencies")
	}
	if !b.SupportsChecksum() || !b.SupportsGPG() {
		t.Errorf("expected checksum and gpg support")
	}
	if b.AttestationType() != "SLSA" || b.GetAttestationType() != "SLSA" {
		t.Errorf("expected SLSA attestation")
	}
	if b.GetClient() == nil {
		t.Errorf("expected non-nil client")
	}
	if !b.IsRecommended() || !b.IsScriptless() || !b.IsStable() {
		t.Errorf("expected true for boolean properties")
	}
	if b.SupportsOffline() {
		t.Errorf("expected false for SupportsOffline")
	}
	if b.GetReach() != "Large" {
		t.Errorf("expected Large for Reach")
	}
}

func TestForgejoFetchReleaseByTag(t *testing.T) {
	ClearForgejoCache()
	defer ClearForgejoCache()

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v1.0.0",
			"assets": [
				{"name": "app-darwin-amd64", "browser_download_url": "http://example.com/app"}
			]
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_FORGEJO_API_URL", server.URL)

	backend := NewForgejoBackend()
	release, err := backend.FetchReleaseByTag(context.Background(), "owner/repo", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.Tag != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", release.Tag)
	}
	if len(release.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(release.Assets))
	}
}

func TestForgejoFetchReleaseByTag_NotFound(t *testing.T) {
	ClearForgejoCache()
	defer ClearForgejoCache()

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_FORGEJO_API_URL", server.URL)

	backend := NewForgejoBackend()
	_, err := backend.FetchReleaseByTag(context.Background(), "owner/repo", "v1.0.0")
	if err == nil {
		t.Errorf("expected error for not found")
	}
}

func TestForgejoBackend_DeduplicationAndCache(t *testing.T) {
	ClearForgejoCache()
	defer ClearForgejoCache()
	var requestCount int
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"tag_name": "v1.0.0", "created_at": "2023-01-01T00:00:00Z", "assets": [{"name": "app-linux-amd64", "browser_download_url": "https://example.com/app"}, {"name": "app-darwin-arm64", "browser_download_url": "https://example.com/app-mac"}]}]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_FORGEJO_API_URL", server.URL)

	b := NewForgejoBackend()
	ctx := context.Background()
	p1 := Platform{OS: "linux", Arch: "amd64"}
	p2 := Platform{OS: "darwin", Arch: "arm64"}

	// 1. First call fetches from REST
	v1, err := b.ListVersions(ctx, "owner/repo", p1)
	if err != nil || len(v1) != 1 {
		t.Fatalf("first call failed: %v", err)
	}

	// 2. Second call across another platform should hit memory cache without extra network request
	v2, err := b.ListVersions(ctx, "owner/repo", p2)
	if err != nil || len(v2) != 1 {
		t.Fatalf("second call failed: %v", err)
	}

	// 3. FetchReleaseByTag for v1.0.0 should also hit cache
	rel, err := b.FetchReleaseByTag(ctx, "owner/repo", "v1.0.0")
	if err != nil || rel == nil || rel.Tag != "v1.0.0" {
		t.Fatalf("FetchReleaseByTag failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 REST request due to caching/deduplication, got %d", requestCount)
	}
}

