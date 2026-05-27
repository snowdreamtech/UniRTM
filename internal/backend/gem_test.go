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

func TestGemBackend_Name(t *testing.T) {
	b := NewGemBackend()
	if b.Name() != "gem" {
		t.Errorf("expected 'gem', got '%s'", b.Name())
	}
	if len(b.Dependencies()) != 1 || b.Dependencies()[0] != "ruby" {
		t.Errorf("expected [ruby] dependencies, got %v", b.Dependencies())
	}
	if !b.SupportsChecksum() || b.SupportsGPG() || b.AttestationType() != "" || !b.IsRecommended() || !b.IsScriptless() || b.GetReach() != "Large" || !b.IsStable() || !b.SupportsOffline() {
		t.Errorf("properties not returning expected values")
	}
}

func TestGemBackend_Interface(t *testing.T) {
	var _ Backend = (*GemBackend)(nil)
}

func TestGemBackend_ListVersions(t *testing.T) {
	b := NewGemBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "versions/rails.json") {
				body := `[{"number": "7.0.0"}, {"number": "7.0.1"}]`
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

	versions, err := b.ListVersions(ctx, "rails", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
}

func TestGemBackend_ResolveVersion(t *testing.T) {
	b := NewGemBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "gems/rails.json") {
				body := `{"version": "7.0.1"}`
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

	info, err := b.ResolveVersion(ctx, "rails", "7.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "7.0.0" {
		t.Errorf("expected version 7.0.0, got %s", info.Version)
	}

	infoLatest, err := b.ResolveVersion(ctx, "rails", "latest", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if infoLatest.Version != "7.0.1" {
		t.Errorf("expected latest version 7.0.1, got %s", infoLatest.Version)
	}
}

func TestGemBackend_GetDownloadInfo(t *testing.T) {
	b := NewGemBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "rails", "7.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "7.0.0" {
		t.Errorf("expected 7.0.0, got %s", info.Version)
	}
}
