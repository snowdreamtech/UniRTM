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

func TestNodeJSHandler_ResolveVersions(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"version":"v20.5.0","files":["linux-x64","osx-x64-tar","win-x64-zip"]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &NodeJSHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	if len(versions) > 0 {
		assert.Equal(t, "20.5.0", versions[0].Version)
	}
}

func TestNodeJSHandler_parseNodeFile(t *testing.T) {
	tests := []struct {
		file     string
		wantOS   string
		wantArch string
		wantExt  string
	}{
		{"linux-x64", "linux", "amd64", ".tar.gz"},
		{"linux-arm64", "linux", "arm64", ".tar.gz"},
		{"linux-x86", "linux", "386", ".tar.gz"},
		{"osx-x64-tar", "darwin", "amd64", ".tar.gz"},
		{"osx-arm64-tar", "darwin", "arm64", ".tar.gz"},
		{"win-x64-zip", "win", "amd64", ".zip"},
		{"win-x86-zip", "win", "386", ".zip"},
		{"src", "", "", ""},
		{"headers", "", "", ""},
		{"unknown", "", "", ""},
	}

	for _, tt := range tests {
		os, arch, _, ext, ok := parseNodeFile(tt.file)
		if tt.wantOS == "" {
			assert.False(t, ok)
			continue
		}
		assert.True(t, ok)
		assert.Equal(t, tt.wantOS, os, "file: %s", tt.file)
		assert.Equal(t, tt.wantArch, arch, "file: %s", tt.file)
		assert.Equal(t, tt.wantExt, ext, "file: %s", tt.file)
	}
}

func TestNodeJSHandler_ResolveVersions_Failures(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Internal Error`))}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &NodeJSHandler{}
	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)

	mockRt.roundTripFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`invalid json`))}, nil
	}
	_, err = h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
}
