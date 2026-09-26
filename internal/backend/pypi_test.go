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

func TestPypiBackend_Name(t *testing.T) {
	b := NewPypiBackend()
	if b.Name() != "pypi" {
		t.Errorf("expected 'pypi', got '%s'", b.Name())
	}
	if b.Dependencies() != nil {
		t.Errorf("expected nil dependencies")
	}
	if !b.SupportsChecksum() || b.SupportsGPG() || b.AttestationType() != "" || !b.IsRecommended() || !b.IsScriptless() || b.GetReach() != "Huge" || !b.IsStable() || !b.SupportsOffline() {
		t.Errorf("properties not returning expected values")
	}
}

func TestPypiBackend_Interface(t *testing.T) {
	var _ Backend = (*PypiBackend)(nil)
}

func TestPypiBackend_ListVersions(t *testing.T) {
	b := NewPypiBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "pypi/black/json") {
				body := `{"releases": {"23.3.0": [], "22.1.0": []}}`
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

	versions, err := b.ListVersions(ctx, "black", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	// Not found
	_, err = b.ListVersions(ctx, "notfound", platform)
	if err == nil {
		t.Error("expected error for not found")
	}

	// execution error
	ClearPypiMetadataCache()
	bErr := NewPypiBackend()
	bErr.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	}
	_, err = bErr.ListVersions(ctx, "black", platform)
	if err == nil {
		t.Error("expected execution error")
	}

	// internal error
	ClearPypiMetadataCache()
	bInternal := NewPypiBackend()
	bInternal.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	_, err = bInternal.ListVersions(ctx, "black", platform)
	if err == nil {
		t.Error("expected status code error")
	}

	// bad json
	ClearPypiMetadataCache()
	bJSON := NewPypiBackend()
	bJSON.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		},
	}
	_, err = bJSON.ListVersions(ctx, "black", platform)
	if err == nil {
		t.Error("expected bad json error")
	}

	// bad request
	ClearPypiMetadataCache()
	_, err = b.ListVersions(nil, "black", platform)
	if err == nil {
		t.Error("expected request creation error with nil context")
	}
}

func TestPypiBackend_ResolveVersion(t *testing.T) {
	ClearPypiMetadataCache()
	defer ClearPypiMetadataCache()

	b := NewPypiBackend()
	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "pypi/black/json") {
				body := `{"info": {"version": "23.3.0"}}`
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

	info, err := b.ResolveVersion(ctx, "black", "23.3.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "23.3.0" {
		t.Errorf("expected version 23.3.0, got %s", info.Version)
	}

	infoLatest, err := b.ResolveVersion(ctx, "black", "latest", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if infoLatest.Version != "23.3.0" {
		t.Errorf("expected latest version 23.3.0, got %s", infoLatest.Version)
	}

	// execution error
	ClearPypiMetadataCache()
	bErr := NewPypiBackend()
	bErr.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	}
	_, err = bErr.ResolveVersion(ctx, "black", "latest", platform)
	if err == nil {
		t.Error("expected execution error")
	}

	// bad json
	ClearPypiMetadataCache()
	bJSON := NewPypiBackend()
	bJSON.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		},
	}
	_, err = bJSON.ResolveVersion(ctx, "black", "latest", platform)
	if err == nil {
		t.Error("expected bad json error")
	}

	// not found latest
	ClearPypiMetadataCache()
	_, err = b.ResolveVersion(ctx, "notfound", "latest", platform)
	if err == nil {
		t.Error("expected not found latest error")
	}

	// bad request latest
	ClearPypiMetadataCache()
	_, err = b.ResolveVersion(nil, "black", "latest", platform)
	if err == nil {
		t.Error("expected request creation error with nil context")
	}
}


func TestPypiBackend_GetDownloadInfo(t *testing.T) {
	b := NewPypiBackend()
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "black", "23.3.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "23.3.0" {
		t.Errorf("expected 23.3.0, got %s", info.Version)
	}
}

func TestPypiBackend_DeduplicationAndCache(t *testing.T) {
	ClearPypiMetadataCache()
	defer ClearPypiMetadataCache()

	b1 := NewPypiBackend()
	var requestCount int
	mockTr := &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			requestCount++
			body := `{"info": {"version": "23.3.0"}, "releases": {"23.3.0": [], "22.1.0": []}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}
	b1.client.Transport = mockTr

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	// 1. First call on b1 populates cache
	versions, err := b1.ListVersions(ctx, "concurrent-pkg", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	// 2. Second call across another instance b2 and platform should hit global memory cache
	b2 := NewPypiBackend()
	b2.client.Transport = mockTr
	platform2 := Platform{OS: "darwin", Arch: "arm64"}
	versions2, err := b2.ListVersions(ctx, "concurrent-pkg", platform2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions2) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions2))
	}

	// 3. ResolveVersion("latest") on b2 should also hit cache
	latest, err := b2.ResolveVersion(ctx, "concurrent-pkg", "latest", platform2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Version != "23.3.0" {
		t.Fatalf("expected 23.3.0, got %s", latest.Version)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 network request due to caching/deduplication, got %d", requestCount)
	}
}


