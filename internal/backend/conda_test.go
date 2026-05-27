// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCondaBackend_Name(t *testing.T) {
	b := NewCondaBackend()
	if b.Name() != "conda" {
		t.Errorf("expected name 'conda', got %s", b.Name())
	}
}

func TestCondaBackend_Properties(t *testing.T) {
	b := NewCondaBackend()

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

func TestCondaBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/package/anaconda/numpy" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"versions": ["1.23.0", "1.24.0", "1.22.0"]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewCondaBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "numpy", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// It should sort them newest first (lexicographical string sort in conda.go)
	if versions[0].Version != "1.24.0" {
		t.Errorf("expected 1.24.0, got %s", versions[0].Version)
	}

	// Test not found
	_, err = b.ListVersions(ctx, "nonexistent", platform)
	if err == nil {
		t.Error("expected error for nonexistent package, got nil")
	}
}

func TestCondaBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/package/anaconda/numpy" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"versions": ["1.23.0", "1.24.0"]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewCondaBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test latest
	vLatest, err := b.ResolveVersion(ctx, "numpy", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vLatest.Version != "1.24.0" {
		t.Errorf("expected 1.24.0, got %s", vLatest.Version)
	}

	// Test specific
	v, err := b.ResolveVersion(ctx, "numpy", "1.24.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "1.24.0" {
		t.Errorf("expected 1.24.0, got %s", v.Version)
	}
}

func TestCondaBackend_GetDownloadInfo(t *testing.T) {
	b := NewCondaBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "numpy", "1.24.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.24.0" {
		t.Errorf("expected 1.24.0, got %s", info.Version)
	}
}
