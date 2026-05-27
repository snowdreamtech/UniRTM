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

func TestNpmBackend_Name(t *testing.T) {
	b := NewNpmBackend()
	if b.Name() != "npm" {
		t.Errorf("expected 'npm', got '%s'", b.Name())
	}
	if len(b.Dependencies()) != 1 || b.Dependencies()[0] != "node" {
		t.Errorf("expected [node] dependencies, got %v", b.Dependencies())
	}
	if !b.SupportsChecksum() || b.SupportsGPG() || b.AttestationType() != "" || !b.IsRecommended() || !b.IsScriptless() || b.GetReach() != "Huge" || !b.IsStable() || !b.SupportsOffline() {
		t.Errorf("properties not returning expected values")
	}
}

func TestNpmBackend_Interface(t *testing.T) {
	var _ Backend = (*NpmBackend)(nil)
}

func TestNpmBackend_ListVersions(t *testing.T) {
	b := NewNpmBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "typescript") {
				body := `{"versions": {"5.0.0": {}, "5.1.0": {}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "typescript", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
}

func TestNpmBackend_ResolveVersion(t *testing.T) {
	b := NewNpmBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "typescript/latest") {
				body := `{"version": "5.1.0"}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "typescript", "5.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "5.0.0" {
		t.Errorf("expected version 5.0.0, got %s", info.Version)
	}

	infoLatest, err := b.ResolveVersion(ctx, "typescript", "latest", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if infoLatest.Version != "5.1.0" {
		t.Errorf("expected latest version 5.1.0, got %s", infoLatest.Version)
	}
}

func TestNpmBackend_GetDownloadInfo(t *testing.T) {
	b := NewNpmBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "typescript", "5.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "5.0.0" {
		t.Errorf("expected 5.0.0, got %s", info.Version)
	}
}
