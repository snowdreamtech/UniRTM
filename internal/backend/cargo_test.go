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

type mockCargoTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockCargoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestCargoBackend_Name(t *testing.T) {
	b := NewCargoBackend()
	if b.Name() != "cargo" {
		t.Errorf("expected 'cargo', got '%s'", b.Name())
	}
	if len(b.Dependencies()) != 1 || b.Dependencies()[0] != "rust" {
		t.Errorf("expected rust dependency")
	}
	if b.AttestationType() != "" {
		t.Errorf("expected empty attestation type")
	}
	if !b.IsRecommended() || !b.IsScriptless() || !b.IsStable() || !b.SupportsOffline() || !b.SupportsChecksum() || b.SupportsGPG() {
		t.Errorf("boolean flag methods not returning expected values")
	}
	if b.GetReach() != "Huge" {
		t.Errorf("expected Huge reach")
	}
}

func TestCargoBackend_Interface(t *testing.T) {
	var _ Backend = (*CargoBackend)(nil)
}

func TestCargoBackend_ResolveVersion_Latest(t *testing.T) {
	b := NewCargoBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "crates/ripgrep") {
				body := `{"crate": {"max_version": "13.0.0"}}`
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

	info, err := b.ResolveVersion(ctx, "ripgrep", "latest", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "13.0.0" {
		t.Errorf("expected version 13.0.0, got %s", info.Version)
	}
}

func TestCargoBackend_ResolveVersion_Specific(t *testing.T) {
	b := NewCargoBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "ripgrep", "12.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "12.0.0" {
		t.Errorf("expected version 12.0.0, got %s", info.Version)
	}
}

func TestCargoBackend_ListVersions(t *testing.T) {
	b := NewCargoBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "crates/ripgrep") {
				body := `{"versions": [{"num": "13.0.0", "created_at": "2021-06-12T10:00:00Z"}, {"num": "12.1.1", "created_at": "2020-01-01T10:00:00Z"}]}`
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

	versions, err := b.ListVersions(ctx, "ripgrep", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != "13.0.0" {
		t.Errorf("expected version 13.0.0")
	}
	if versions[0].PublishedAt.IsZero() {
		t.Errorf("expected non-zero published at")
	}
}

func TestCargoBackend_GetDownloadInfo(t *testing.T) {
	b := NewCargoBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "ripgrep", "13.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "13.0.0" {
		t.Errorf("expected 13.0.0, got %s", info.Version)
	}
}
