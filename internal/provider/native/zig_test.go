// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestZigHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"0.11.0": {"x86_64-linux": {"tarball": "https://ziglang.org/download/0.11.0/zig-linux-x86_64-0.11.0.tar.xz"}}}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &ZigHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "0.11.0", versions[0].Version)
	assert.NotEmpty(t, versions[0].Assets)
}

func TestZigHandler_parsePlatform(t *testing.T) {
	h := &ZigHandler{}

	tests := []struct {
		platform string
		wantOS   string
		wantArch string
	}{
		{"x86_64-linux", "linux", "amd64"},
		{"aarch64-linux", "linux", "arm64"},
		{"x86-linux", "linux", "386"},
		{"x86_64-macos", "darwin", "amd64"},
		{"aarch64-macos", "darwin", "arm64"},
		{"x86_64-windows", "windows", "amd64"},
		{"aarch64-windows", "windows", "arm64"},
		{"x86-windows", "windows", "386"},
		{"unknown-platform", "", ""},
	}

	for _, tt := range tests {
		os, arch := h.parsePlatform(tt.platform)
		assert.Equal(t, tt.wantOS, os, "platform: %s", tt.platform)
		assert.Equal(t, tt.wantArch, arch, "platform: %s", tt.platform)
	}
}

func TestZigHandler_ResolveVersions_Failures(t *testing.T) {
	// Test HTTP error
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Internal Error`))}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &ZigHandler{}
	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)

	// Test Invalid JSON
	mockRt.roundTripFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`invalid json`))}, nil
	}
	_, err = h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
}
