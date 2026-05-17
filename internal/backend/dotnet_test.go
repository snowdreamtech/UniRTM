// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDotnetBackend_Name(t *testing.T) {
	b := NewDotnetBackend()
	if b.Name() != "dotnet" {
		t.Errorf("expected name 'dotnet', got %s", b.Name())
	}
}

func TestDotnetBackend_ResolveVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := NewDotnetBackend()

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "dotnet-ef", "7.0.5", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "7.0.5" {
		t.Errorf("expected version '7.0.5', got %s", info.Version)
	}
}
