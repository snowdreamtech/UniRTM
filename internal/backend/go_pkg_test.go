// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoPkgBackend_Name(t *testing.T) {
	b := NewGoPkgBackend()
	if b.Name() != "go-pkg" {
		t.Errorf("expected go-pkg, got %s", b.Name())
	}
}

func TestGoPkgBackend_Methods(t *testing.T) {
	b := NewGoPkgBackend()

	deps := b.Dependencies()
	if len(deps) != 1 || deps[0] != "go" {
		t.Errorf("expected [go], got %v", deps)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("v1.0.0\nv2.0.0\n"))
	}))
	defer ts.Close()

	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "golang.org/x/tools", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	v, err := b.ResolveVersion(ctx, "golang.org/x/tools", "v1.0.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", v.Version)
	}
}
