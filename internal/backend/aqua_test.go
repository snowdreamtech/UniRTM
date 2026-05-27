// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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

func TestAquaBackend_ResolveVersion(t *testing.T) {
	b := NewAquaBackend()
	b.registryURL = "https://raw.githubusercontent.com/aquaproj/aqua-registry/main/pkgs"
	
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "aquaproj/aqua/pkg.yaml") {
				body := `{"type": "github_release", "repo_owner": "aquaproj", "repo_name": "aqua", "asset": "aqua_{{.OS}}_{{.Arch}}.tar.gz"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
			} else if strings.Contains(req.URL.Path, "repos/aquaproj/aqua/releases") {
				body := `[{"tag_name": "v2.0.1", "name": "v2.0.1"}, {"tag_name": "v2.0.0", "name": "v2.0.0"}]`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// Test exact match
	info, err := b.ResolveVersion(ctx, "aquaproj/aqua", "2.0.0", platform)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if info.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", info.Version)
	}

	// Test latest
	infoLatest, err := b.ResolveVersion(ctx, "aquaproj/aqua", "latest", platform)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if infoLatest.Version != "2.0.1" {
		t.Errorf("expected 2.0.1, got %s", infoLatest.Version)
	}
}

func TestAquaBackend_GetDownloadInfo(t *testing.T) {
	b := NewAquaBackend()
	b.registryURL = "https://raw.githubusercontent.com/aquaproj/aqua-registry/main/pkgs"
	
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "aquaproj/aqua/pkg.yaml") {
				body := `{"type": "github_release", "repo_owner": "aquaproj", "repo_name": "aqua", "asset": "aqua_{{.OS}}_{{.Arch}}.tar.gz"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
			} else if strings.Contains(req.URL.Path, "repos/aquaproj/aqua/releases") {
				body := `[{"tag_name": "v2.0.0", "name": "v2.0.0"}]`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(body))}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "aquaproj/aqua", "2.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", info.Version)
	}
}
