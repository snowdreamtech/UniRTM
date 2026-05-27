// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCabalBackend_Name(t *testing.T) {
	b := NewCabalBackend()
	if b.Name() != "cabal" {
		t.Errorf("expected cabal, got %s", b.Name())
	}
}

func TestCabalBackend_Properties(t *testing.T) {
	b := NewCabalBackend()

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
	if b.GetReach() != "Large" {
		t.Errorf("expected Large, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestCabalBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/package/pandoc.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"version": "2.19"},
				{"version": "3.0"},
				{"version": "3.1"}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Override URL creation in test by mocking client Transport
	b := NewCabalBackend()
	
	// Temporarily hijack http request using a custom transport for testing
	b.client.Transport = &mockTransport{
		rt: http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "pandoc", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].Version != "2.19" {
		t.Errorf("expected 2.19, got %s", versions[0].Version)
	}

	// Test not found
	_, err = b.ListVersions(ctx, "nonexistent", platform)
	if err == nil {
		t.Error("expected error for nonexistent package, got nil")
	}
}

func TestCabalBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/package/pandoc.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"version": "2.19"},
				{"version": "3.0"},
				{"version": "3.1"}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewCabalBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test latest
	vLatest, err := b.ResolveVersion(ctx, "pandoc", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vLatest.Version != "3.1" {
		t.Errorf("expected 3.1, got %s", vLatest.Version)
	}

	// Test specific
	v, err := b.ResolveVersion(ctx, "pandoc", "3.1", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "3.1" {
		t.Errorf("expected 3.1, got %s", v.Version)
	}
}

func TestCabalBackend_GetDownloadInfo(t *testing.T) {
	b := NewCabalBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "pandoc", "3.1", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "3.1" {
		t.Errorf("expected 3.1, got %s", info.Version)
	}
}

// mockTransport intercepts requests and routes them to the test server
type mockTransport struct {
	rt  http.RoundTripper
	url string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Re-route to our test server
	req.URL.Scheme = "http"
	req.URL.Host = m.url[7:] // strip "http://"
	return m.rt.RoundTrip(req)
}
