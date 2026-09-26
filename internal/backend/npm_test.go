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

	// Not found
	_, err = b.ListVersions(ctx, "notfound", platform)
	if err == nil {
		t.Error("expected error for not found")
	}

	// execution error
	ClearNpmMetadataCache()
	bErr := NewNpmBackend()
	bErr.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	}
	_, err = bErr.ListVersions(ctx, "typescript", platform)
	if err == nil {
		t.Error("expected execution error")
	}

	// internal error
	ClearNpmMetadataCache()
	bInternal := NewNpmBackend()
	bInternal.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
		},
	}
	_, err = bInternal.ListVersions(ctx, "typescript", platform)
	if err == nil {
		t.Error("expected status code error")
	}

	// bad json
	ClearNpmMetadataCache()
	bJSON := NewNpmBackend()
	bJSON.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		},
	}
	_, err = bJSON.ListVersions(ctx, "typescript", platform)
	if err == nil {
		t.Error("expected bad json error")
	}

	// time parse
	ClearNpmMetadataCache()
	bTime := NewNpmBackend()
	bTime.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"versions":{"1.0":{}},"time":{"1.0":"2023-01-01T00:00:00Z"}}`))}, nil
		},
	}
	vs, err := bTime.ListVersions(ctx, "typescript", platform)
	if err == nil && len(vs) == 1 {
		if vs[0].PublishedAt.IsZero() {
			t.Error("expected parsed time")
		}
	}

	// bad request
	ClearNpmMetadataCache()
	_, err = b.ListVersions(nil, "typescript", platform)
	if err == nil {
		t.Error("expected request creation error with nil context")
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

	// execution error
	ClearNpmMetadataCache()
	bErr := NewNpmBackend()
	bErr.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		},
	}
	_, err = bErr.ResolveVersion(ctx, "typescript", "latest", platform)
	if err == nil {
		t.Error("expected execution error")
	}

	// bad json
	ClearNpmMetadataCache()
	bJSON := NewNpmBackend()
	bJSON.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		},
	}
	_, err = bJSON.ResolveVersion(ctx, "typescript", "latest", platform)
	if err == nil {
		t.Error("expected bad json error")
	}

	// not found latest
	ClearNpmMetadataCache()
	_, err = b.ResolveVersion(ctx, "notfound", "latest", platform)
	if err == nil {
		t.Error("expected not found latest error")
	}

	// bad request latest
	ClearNpmMetadataCache()
	_, err = b.ResolveVersion(nil, "typescript", "latest", platform)
	if err == nil {
		t.Error("expected request creation error with nil context")
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

func TestNpmBackend_DeduplicationAndCache(t *testing.T) {
	ClearNpmMetadataCache()
	defer ClearNpmMetadataCache()

	b1 := NewNpmBackend()
	var requestCount int
	mockTr := &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			requestCount++
			body := `{"dist-tags": {"latest": "5.1.0"}, "versions": {"5.0.0": {}, "5.1.0": {}}}`
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
	b2 := NewNpmBackend()
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
	if latest.Version != "5.1.0" {
		t.Fatalf("expected 5.1.0, got %s", latest.Version)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 network request due to caching/deduplication, got %d", requestCount)
	}
}


