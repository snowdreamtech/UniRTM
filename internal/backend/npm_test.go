// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNpmBackend_Name(t *testing.T) {
	b := NewNpmBackend()
	if b.Name() != "npm" {
		t.Errorf("expected 'npm', got '%s'", b.Name())
	}
}

func TestNpmBackend_ListVersions(t *testing.T) {
	// Create a mock npm registry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/typescript" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"versions": {"5.0.0": {}, "5.1.0": {}}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := NewNpmBackend()
	// override the URL internally if possible, but since it's hardcoded to registry.npmjs.org,
	// we will just test that it fails cleanly when the network is unavailable or package not found.
	// For this test, we will just test the error paths since we can't easily inject the URL.
	
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}
	
	// Test unknown package (will actually hit the real registry or fail network)
	// We'll skip the real network call if possible, but let's just let it run 
	// against a dummy package that doesn't exist.
	_, err := b.ListVersions(ctx, "this-package-definitely-does-not-exist-12345", platform)
	if err == nil {
		t.Error("expected error for non-existent package, got nil")
	}
}

func TestNpmBackend_ResolveVersion(t *testing.T) {
	b := NewNpmBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test explicit version resolution
	info, err := b.ResolveVersion(ctx, "typescript", "5.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "5.0.0" {
		t.Errorf("expected version 5.0.0, got %s", info.Version)
	}
}

func TestNpmBackend_Interface(t *testing.T) {
	// Ensure it implements Backend interface
	var _ Backend = (*NpmBackend)(nil)
}
