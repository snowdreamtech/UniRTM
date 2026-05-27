// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestDotnetBackend_Interface(t *testing.T) {
	var _ Backend = (*DotnetBackend)(nil)
}

func TestDotnetBackend_Properties(t *testing.T) {
	b := NewDotnetBackend()
	if b.Name() != "dotnet" {
		t.Errorf("expected name dotnet, got %s", b.Name())
	}
	if b.Dependencies() != nil {
		t.Errorf("expected no dependencies")
	}
	if !b.SupportsChecksum() || b.SupportsGPG() {
		t.Errorf("expected checksum true and gpg support to be false")
	}
}

func TestDotnetBackend_ListVersions(t *testing.T) {
	mockTransport := &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://api.nuget.org/v3-flatcontainer/tool/index.json" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"versions": ["1.0.0", "1.1.0", "2.0.0"]}`)),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"versions": []}`)),
			}, nil
		},
	}
	client := &http.Client{Transport: mockTransport}
	b := NewDotnetBackend()
	b.client = client

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "tool", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// Note: It reverses the versions!
	if versions[0].Version != "2.0.0" {
		t.Errorf("expected highest version first: %s", versions[0].Version)
	}

	_, err = b.ListVersions(ctx, "notfound", p)
	if err == nil {
		t.Errorf("expected error for notfound")
	}
}

func TestDotnetBackend_ResolveVersion(t *testing.T) {
	mockTransport := &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://api.nuget.org/v3-flatcontainer/tool/index.json" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"versions": ["1.0.0", "1.1.0", "2.0.0"]}`)),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"versions": []}`)),
			}, nil
		},
	}
	client := &http.Client{Transport: mockTransport}
	b := NewDotnetBackend()
	b.client = client

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	v, err := b.ResolveVersion(ctx, "tool", "1.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", v.Version)
	}

	v2, err := b.ResolveVersion(ctx, "tool", "latest", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", v2.Version)
	}

	_, err = b.ResolveVersion(ctx, "empty", "latest", p)
	if err == nil {
		t.Errorf("expected error for empty latest")
	}
}

func TestDotnetBackend_GetDownloadInfo(t *testing.T) {
	b := NewDotnetBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "tool", "1.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", info.Version)
	}
}
