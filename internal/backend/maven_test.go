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

func TestMavenBackend_Name(t *testing.T) {
	b := NewMavenBackend()
	if b.Name() != "maven" {
		t.Errorf("expected maven, got %s", b.Name())
	}
}

func TestMavenBackend_Properties(t *testing.T) {
	b := NewMavenBackend()

	if deps := b.Dependencies(); deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
	if !b.SupportsChecksum() {
		t.Error("expected SupportsChecksum to be true")
	}
	if !b.SupportsGPG() {
		t.Error("expected SupportsGPG to be true")
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
	if b.GetReach() != "Huge" {
		t.Errorf("expected Huge, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if !b.SupportsOffline() {
		t.Error("expected SupportsOffline to be true")
	}
}

func TestMavenBackend_ListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/solrsearch/select" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"response": {
					"docs": [
						{"v": "3.8.1"},
						{"v": "3.8.0"}
					]
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewMavenBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test success
	versions, err := b.ListVersions(ctx, "org.apache.maven:maven-core", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != "3.8.1" {
		t.Errorf("expected 3.8.1, got %s", versions[0].Version)
	}

	// Test invalid tool name
	_, err = b.ListVersions(ctx, "invalid-tool-format", platform)
	if err == nil {
		t.Error("expected error for invalid tool name format, got nil")
	}
}

func TestMavenBackend_ResolveVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"response": {
				"docs": [
					{"v": "4.0.0"}
				]
			}
		}`))
	}))
	defer ts.Close()

	b := NewMavenBackend()
	b.client.Transport = &mockTransport{
		rt:  http.DefaultTransport,
		url: ts.URL,
	}

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "g:a", "latest", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "4.0.0" {
		t.Errorf("expected 4.0.0, got %s", v.Version)
	}

	v2, err := b.ResolveVersion(ctx, "g:a", "3.0.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.Version != "3.0.0" {
		t.Errorf("expected 3.0.0, got %s", v2.Version)
	}
}

func TestMavenBackend_GetDownloadInfo(t *testing.T) {
	b := NewMavenBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "g:a", "1.0", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0" {
		t.Errorf("expected 1.0, got %s", info.Version)
	}
}

func TestMavenBackend_ConcurrentDeduplication(t *testing.T) {
	ClearMavenCache()
	defer ClearMavenCache()

	var requestCount int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"response": {
				"docs": [
					{"v": "3.8.1"},
					{"v": "3.8.0"}
				]
			}
		}`))
	}))
	defer ts.Close()

	b := NewMavenBackend()
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
		res, err := b.ListVersions(context.Background(), "org.apache.maven:maven-core", p)
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

func TestMavenBackend_DiskCache(t *testing.T) {
	ClearMavenCache()
	defer ClearMavenCache()

	tool := "com.example:my-lib"
	dummyVersions := []string{"1.0.0", "1.1.0"}
	writeEcosystemMetadataDiskCache("maven", tool, dummyVersions)

	b := NewMavenBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, tool, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 || versions[0].Version != "1.0.0" {
		t.Fatalf("expected cached versions, got %v", versions)
	}
}
