// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGemBackend_Name(t *testing.T) {
	b := NewGemBackend()
	if b.Name() != "gem" {
		t.Errorf("expected name 'gem', got %s", b.Name())
	}
}

func TestGemBackend_ResolveVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/gems/bundler.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"bundler","version":"2.4.22"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := NewGemBackend()
	// Override the client's transport if we wanted to mock the URL, but here we just
	// test the hardcoded explicit version logic since we can't easily inject the URL into NewGemBackend.

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	// Test explicit version (does not make HTTP request)
	info, err := b.ResolveVersion(ctx, "bundler", "2.4.22", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "2.4.22" {
		t.Errorf("expected version '2.4.22', got %s", info.Version)
	}
}
