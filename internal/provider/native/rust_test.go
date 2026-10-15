// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestRustHandler_Name(t *testing.T) {
	h := &RustHandler{}
	assert.Equal(t, "rust", h.Name())
}

func TestRustHandler_ResolveVersions_Mock(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			url := req.URL.String()
			if strings.Contains(url, "/repos/rust-lang/rust/tags") {
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(bytes.NewBufferString(`[
						{"name": "1.85.0"},
						{"name": "1.84.1"},
						{"name": "1.84.0"},
						{"name": "release-0.7"},
						{"name": "1.74.0"}
					]`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString("Not found"))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &RustHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, versions)

	// release-0.7 should be filtered out, 4 valid versions
	assert.Len(t, versions, 4)
	assert.Equal(t, "1.85.0", versions[0].Version)
	assert.Equal(t, "1.84.1", versions[1].Version)
	assert.Equal(t, "1.84.0", versions[2].Version)
	assert.Equal(t, "1.74.0", versions[3].Version)

	// Verify generated assets
	assert.NotEmpty(t, versions[0].Assets)
	assert.Contains(t, versions[0].Assets[0].URL, "https://static.rust-lang.org/dist/rust-1.85.0")
}

func TestRustHandler_ResolveVersions_Fallback(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString("Server Error"))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &RustHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, versions)
	// Should contain rich fallback versions
	assert.Greater(t, len(versions), 10)
}
