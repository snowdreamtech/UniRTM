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

func TestGitlabBackend_Name(t *testing.T) {
	b := NewGitlabBackend()
	if b.Name() != "gitlab" {
		t.Errorf("expected name 'gitlab', got %s", b.Name())
	}
}

func TestGitlabBackend_ResolveVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/owner/repo/releases" || r.URL.Path == "/projects/owner%2Frepo/releases" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"tag_name":"v1.0.0","assets":{"links":[{"name":"tool-linux-amd64.tar.gz","url":"https://example.com/download"}]}}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := &GitlabBackend{
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

func TestGitlabBackend_ListVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"tag_name":"v2.0.0","assets":{"links":[{"name":"tool-linux-amd64.tar.gz","url":"https://example.com/download2"}]}},
			{"tag_name":"v1.0.0","assets":{"links":[{"name":"tool-linux-amd64.tar.gz","url":"https://example.com/download"}]}}
		]`))
	}))
	defer server.Close()

	b := &GitlabBackend{
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

func TestGitlabBackend_Properties(t *testing.T) {
	b := NewGitlabBackend()
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

func TestGitlabFetchReleaseByTag(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/releases/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v1.0.0",
			"assets": {
				"links": [
					{"name": "app-darwin-amd64", "url": "http://example.com/app", "direct_asset_url": "http://example.com/app"}
				]
			}
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	os.Setenv("UNIRTM_GITLAB_API_URL", server.URL)
	defer os.Unsetenv("UNIRTM_GITLAB_API_URL")

	backend := NewGitlabBackend()
	release, err := backend.FetchReleaseByTag(context.Background(), "owner/repo", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.Tag != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", release.Tag)
	}
	if len(release.Assets) != 1 {
		t.Errorf("expected 1 asset link, got %d", len(release.Assets))
	}
}

func TestGitlabFetchReleaseByTag_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/releases/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	os.Setenv("UNIRTM_GITLAB_API_URL", server.URL)
	defer os.Unsetenv("UNIRTM_GITLAB_API_URL")

	backend := NewGitlabBackend()
	_, err := backend.FetchReleaseByTag(context.Background(), "owner/repo", "v1.0.0")
	if err == nil {
		t.Errorf("expected error for not found")
	}
}
