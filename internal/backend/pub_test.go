// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPubBackend_Name(t *testing.T) {
	b := NewPubBackend()
	if b.Name() != "pub" {
		t.Errorf("expected pub, got %s", b.Name())
	}
}

func TestPubBackend_Properties(t *testing.T) {
	b := NewPubBackend()

	if deps := b.Dependencies(); len(deps) != 1 || deps[0] != "dart" {
		t.Errorf("expected [dart] dependencies, got %v", deps)
	}
	if b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be false")
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
	if b.GetReach() != "Medium" {
		t.Errorf("expected Medium, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestPubBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/packages/shelf" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"name": "shelf",
				"versions": [
					{"version": "1.4.0"},
					{"version": "1.4.1"}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewPubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test nil context
	_, err := b.ListVersions(nil, "shelf", platform)
	if err == nil {
		t.Error("expected error with nil context, got nil")
	}

	// Test success
	versions, err := b.ListVersions(context.Background(), "shelf", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	// Verify reversed order
	if versions[0].Version != "1.4.1" {
		t.Errorf("expected 1.4.1 as first version, got %s", versions[0].Version)
	}

	// Test not found
	_, err = b.ListVersions(context.Background(), "notfound", platform)
	if err == nil {
		t.Error("expected error for not found, got nil")
	}
}

func TestPubBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/packages/shelf" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"name": "shelf",
				"versions": [
					{"version": "1.4.0"},
					{"version": "1.4.1"}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewPubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	vLatest, err := b.ResolveVersion(ctx, "shelf", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vLatest.Version != "1.4.1" {
		t.Errorf("expected 1.4.1, got %s", vLatest.Version)
	}

	vExact, err := b.ResolveVersion(ctx, "shelf", "1.4.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vExact.Version != "1.4.0" {
		t.Errorf("expected 1.4.0, got %s", vExact.Version)
	}
}

func TestPubBackend_GetDownloadInfo(t *testing.T) {
	b := NewPubBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "shelf", "1.4.1", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.4.1" {
		t.Errorf("expected 1.4.1, got %s", info.Version)
	}
}

func TestPubBackend_ConcurrentDeduplication(t *testing.T) {
	var requestCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"name": "shelf",
			"versions": [
				{"version": "1.4.0"},
				{"version": "1.4.1"}
			]
		}`))
	}))
	defer ts.Close()

	b := NewPubBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	platforms := []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
	}

	for _, p := range platforms {
		res, err := b.ListVersions(context.Background(), "shelf", p)
		if err != nil {
			t.Fatalf("ListVersions error: %v", err)
		}
		if len(res) != 2 || res[0].Platform != p {
			t.Errorf("unexpected result for platform %v", p)
		}
	}

	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("expected exactly 1 request due to cache, got %d", requestCount)
	}
}
