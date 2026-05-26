// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAquaBackend_Name(t *testing.T) {
	b := NewAquaBackend()
	if b.Name() != "aqua" {
		t.Errorf("expected aqua, got %s", b.Name())
	}
}

func TestAquaBackend_Properties(t *testing.T) {
	b := NewAquaBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if b.SupportsGPG() {
		t.Error("expected SupportsGPG to be false")
	}
	if b.AttestationType() != "SLSA" {
		t.Errorf("expected SLSA, got %s", b.AttestationType())
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

func TestAquaBackend_FetchPackageMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/aquaproj/aqua/pkg.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"type": "github_release",
				"repo_owner": "aquaproj",
				"repo_name": "aqua",
				"asset": "aqua_{{.OS}}_{{.Arch}}.tar.gz"
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewAquaBackend()
	b.registryURL = ts.URL // Point to our mock server

	ctx := context.Background()

	// Test successful fetch
	pkg, err := b.fetchPackageMetadata(ctx, "aquaproj/aqua")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Type != "github_release" {
		t.Errorf("expected github_release, got %s", pkg.Type)
	}

	// Test not found
	_, err = b.fetchPackageMetadata(ctx, "nonexistent/tool")
	if err == nil {
		t.Fatal("expected error for nonexistent tool, got nil")
	}
}
