// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPBackend_Name(t *testing.T) {
	b := NewHTTPBackend()
	if b.Name() != "http" {
		t.Errorf("expected http, got %s", b.Name())
	}
}

func TestHTTPBackend_Properties(t *testing.T) {
	b := NewHTTPBackend()

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
	if b.GetReach() != "Small" {
		t.Errorf("expected Small, got %s", b.GetReach())
	}
	if !b.IsStable() {
		t.Error("expected IsStable to be true")
	}
	if b.SupportsOffline() {
		t.Error("expected SupportsOffline to be false")
	}
}

func TestHTTPBackend_ListVersions(t *testing.T) {
	b := NewHTTPBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	_, err := b.ListVersions(ctx, "tool", platform)
	if err == nil {
		t.Error("expected error for ListVersions, got nil")
	}
}

func TestHTTPBackend_ResolveVersion(t *testing.T) {
	b := NewHTTPBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	_, err := b.ResolveVersion(ctx, "tool", "1.0.0", platform)
	if err == nil {
		t.Error("expected error for ResolveVersion without config, got nil")
	}
}

func TestHTTPBackend_GetDownloadInfo(t *testing.T) {
	b := NewHTTPBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	_, err := b.GetDownloadInfo(ctx, "tool", "1.0.0", platform)
	if err == nil {
		t.Error("expected error for GetDownloadInfo without config, got nil")
	}
}

func TestHTTPBackend_GetDownloadInfoWithConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tool-1.0.0-linux-amd64.tar.gz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/tool-1.0.0-linux-amd64.tar.gz.sha256" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("abcd1234efgh5678  tool-1.0.0-linux-amd64.tar.gz\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	b := NewHTTPBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	config := HTTPConfig{
		URLTemplate:      ts.URL + "/tool-{{.version}}-{{.os}}-{{.arch}}.tar.gz",
		ChecksumTemplate: ts.URL + "/tool-{{.version}}-{{.os}}-{{.arch}}.tar.gz.sha256",
	}

	info, err := b.GetDownloadInfoWithConfig(ctx, "tool", "1.0.0", platform, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := ts.URL + "/tool-1.0.0-linux-amd64.tar.gz"
	if info.DownloadURL != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, info.DownloadURL)
	}
	if info.Checksum != "abcd1234efgh5678" {
		t.Errorf("expected checksum abcd1234efgh5678, got %s", info.Checksum)
	}
}

func TestHTTPBackend_BuildURL(t *testing.T) {
	b := NewHTTPBackend()
	platform := Platform{OS: "darwin", Arch: "amd64"}

	template := "https://example.com/{{.version}}/{{.os_alt}}/{{.arch_alt}}/file-{{.OS_ALT}}-{{.ARCH_ALT}}.zip"
	url := b.buildURL(template, "1.2.3", platform, nil)

	expected := "https://example.com/1.2.3/macos/x86_64/file-macOS-x86_64.zip"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}
