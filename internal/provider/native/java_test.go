// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestJavaHandler_ResolveVersions(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"release_name":"jdk-17","version_data":{"openjdk_version":"17.0.2+8"},"binaries":[{"package":{"name":"file.tar.gz","link":"url"}}]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	// It will query 5 major versions, returning 5 results with version "17.0.2"
	assert.Len(t, versions, 5)
	assert.Equal(t, "17.0.2", versions[0].Version)
}

func TestJavaHandler_ResolveVersions_Failures(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Error`))}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	// It continues on error so it shouldn't return error
	assert.NoError(t, err)
	assert.Len(t, versions, 0)
}

func TestJavaHandler_ConcurrentDeduplication(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	var reqCount int32
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&reqCount, 1)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"release_name":"jdk-17","version_data":{"openjdk_version":"17.0.2+8"},"binaries":[{"package":{"name":"file.tar.gz","link":"url"}}]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			versions, err := h.ResolveVersions(context.Background(), "")
			assert.NoError(t, err)
			assert.Len(t, versions, 5)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// 5 major versions queried in single execution = 5 requests total instead of 5 * 5 = 25
	assert.Equal(t, int32(5), atomic.LoadInt32(&reqCount), "Expected exactly 5 requests (one per major version) for all concurrent callers")
}

func TestJavaHandler_ResolveVersions_MultiPlatform(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	mockJSON := `[{
		"release_name": "jdk-21.0.2+13",
		"version_data": {"openjdk_version": "21.0.2+13-LTS"},
		"binaries": [
			{"os": "mac", "architecture": "x64", "package": {"name": "mac-x64.tar.gz", "link": "https://download/mac-x64.tar.gz", "checksum": "abc1"}},
			{"os": "mac", "architecture": "aarch64", "package": {"name": "mac-arm64.tar.gz", "link": "https://download/mac-arm64.tar.gz", "checksum": "abc2"}},
			{"os": "linux", "architecture": "x64", "package": {"name": "linux-x64.tar.gz", "link": "https://download/linux-x64.tar.gz", "checksum": "abc3"}},
			{"os": "alpine-linux", "architecture": "x64", "package": {"name": "alpine-x64.tar.gz", "link": "https://download/alpine-x64.tar.gz", "checksum": "abc4"}},
			{"os": "windows", "architecture": "x64", "package": {"name": "win-x64.zip", "link": "https://download/win-x64.zip", "checksum": "abc5"}}
		]
	}]`

	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			assert.Contains(t, req.URL.RawQuery, "page_size=50")
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(mockJSON)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, versions)

	v := versions[0]
	assert.Equal(t, "21.0.2", v.Version)
	assert.Len(t, v.Assets, 5)

	// Check platform mappings
	assert.Equal(t, "darwin", v.Assets[0].OS)
	assert.Equal(t, "amd64", v.Assets[0].Arch)
	assert.Equal(t, "abc1", v.Assets[0].Checksum)
	assert.Equal(t, "sha256", v.Assets[0].Algo)

	assert.Equal(t, "darwin", v.Assets[1].OS)
	assert.Equal(t, "arm64", v.Assets[1].Arch)

	assert.Equal(t, "linux", v.Assets[2].OS)
	assert.Equal(t, "amd64", v.Assets[2].Arch)

	assert.Equal(t, "linux", v.Assets[3].OS)
	assert.Equal(t, "amd64", v.Assets[3].Arch)

	assert.Equal(t, "windows", v.Assets[4].OS)
	assert.Equal(t, "amd64", v.Assets[4].Arch)
}
