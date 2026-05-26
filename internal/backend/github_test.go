// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubBackend_Name(t *testing.T) {
	b := NewGitHubBackend()
	if b.Name() != "github" {
		t.Errorf("expected github, got %s", b.Name())
	}
}

func TestGitHubBackend_Properties(t *testing.T) {
	b := NewGitHubBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if b.GetClient() == nil {
		t.Error("expected non-nil client")
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if !b.SupportsGPG() {
		t.Error("expected SupportsGPG to be true")
	}
	if b.AttestationType() != "GitHub Attestation" {
		t.Errorf("expected GitHub Attestation, got %s", b.AttestationType())
	}
	if b.GetAttestationType() != "GitHub Attestation" {
		t.Errorf("expected GitHub Attestation, got %s", b.GetAttestationType())
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

func TestGitHubBackend_FetchReleases(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/releases" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"tag_name": "v1.2.0",
					"prerelease": false,
					"created_at": "2023-01-01T00:00:00Z",
					"assets": [
						{
							"name": "asset-linux-amd64.tar.gz",
							"browser_download_url": "https://example.com/asset.tar.gz",
							"size": 12345
						}
					]
				}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewGitHubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()

	releases, err := b.FetchReleases(ctx, "owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].Tag != "v1.2.0" {
		t.Errorf("expected v1.2.0, got %s", releases[0].Tag)
	}
	if len(releases[0].Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(releases[0].Assets))
	}
	if releases[0].Assets[0].Name != "asset-linux-amd64.tar.gz" {
		t.Errorf("expected asset-linux-amd64.tar.gz, got %s", releases[0].Assets[0].Name)
	}
}

func TestGitHubBackend_FetchReleaseByTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/releases/tags/v1.2.0" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"tag_name": "v1.2.0",
				"prerelease": false,
				"created_at": "2023-01-01T00:00:00Z",
				"assets": []
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewGitHubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()

	release, err := b.FetchReleaseByTag(ctx, "owner/repo", "v1.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.Tag != "v1.2.0" {
		t.Errorf("expected v1.2.0, got %s", release.Tag)
	}
}

func TestGitHubBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/releases" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"tag_name": "v1.2.0",
					"assets": [
						{"name": "tool-linux-amd64.tar.gz", "browser_download_url": "https://example.com/asset.tar.gz"}
					]
				}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewGitHubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "owner/repo", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	// Note: ListVersions trims "v" prefix in this implementation
	if versions[0].Version != "1.2.0" {
		t.Errorf("expected 1.2.0, got %s", versions[0].Version)
	}
}
