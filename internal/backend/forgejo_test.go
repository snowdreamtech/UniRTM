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
