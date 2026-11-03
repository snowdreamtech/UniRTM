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

func TestPhpHandler_Name(t *testing.T) {
	h := &PhpHandler{}
	assert.Equal(t, "php", h.Name())
}

func TestPhpHandler_ResolveVersions(t *testing.T) {
	ClearPhpHandlerCache()
	defer ClearPhpHandlerCache()

	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			jsonResp := `[
				{"is_dir":false,"name":"php-8.4.4-cli-linux-x86_64.tar.gz","full_path":"/static-php-cli/bulk/php-8.4.4-cli-linux-x86_64.tar.gz"},
				{"is_dir":false,"name":"php-8.4.4-cli-linux-aarch64.tar.gz","full_path":"/static-php-cli/bulk/php-8.4.4-cli-linux-aarch64.tar.gz"},
				{"is_dir":false,"name":"php-8.4.4-cli-macos-x86_64.tar.gz","full_path":"/static-php-cli/bulk/php-8.4.4-cli-macos-x86_64.tar.gz"},
				{"is_dir":false,"name":"php-8.4.4-cli-macos-aarch64.tar.gz","full_path":"/static-php-cli/bulk/php-8.4.4-cli-macos-aarch64.tar.gz"},
				{"is_dir":false,"name":"php-8.3.15-cli-linux-x86_64.tar.gz","full_path":"/static-php-cli/bulk/php-8.3.15-cli-linux-x86_64.tar.gz"},
				{"is_dir":true,"name":"license/","full_path":"/static-php-cli/bulk/license/"}
			]`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &PhpHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 2)

	// Semver descending order: 8.4.4 before 8.3.15
	assert.Equal(t, "8.4.4", versions[0].Version)
	assert.Equal(t, "8.3.15", versions[1].Version)

	// Check assets for 8.4.4: 4 static-php (linux x64/arm64, macos x64/arm64) + 2 windows (x64, 386)
	v844 := versions[0]
	assert.Len(t, v844.Assets, 6)

	hasLinuxAmd64 := false
	hasLinuxArm64 := false
	hasDarwinAmd64 := false
	hasDarwinArm64 := false
	hasWindowsAmd64 := false
	for _, a := range v844.Assets {
		if a.OS == "linux" && a.Arch == "amd64" {
			hasLinuxAmd64 = true
			assert.Contains(t, a.URL, "php-8.4.4-cli-linux-x86_64.tar.gz")
		}
		if a.OS == "linux" && a.Arch == "arm64" {
			hasLinuxArm64 = true
		}
		if a.OS == "darwin" && a.Arch == "amd64" {
			hasDarwinAmd64 = true
		}
		if a.OS == "darwin" && a.Arch == "arm64" {
			hasDarwinArm64 = true
		}
		if a.OS == "windows" && a.Arch == "amd64" {
			hasWindowsAmd64 = true
			assert.Contains(t, a.URL, "php-8.4.4-Win32-vs17-x64.zip")
		}
	}
	assert.True(t, hasLinuxAmd64)
	assert.True(t, hasLinuxArm64)
	assert.True(t, hasDarwinAmd64)
	assert.True(t, hasDarwinArm64)
	assert.True(t, hasWindowsAmd64)
}

func TestPhpHandler_ResolveTag(t *testing.T) {
	ClearPhpHandlerCache()
	defer ClearPhpHandlerCache()

	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			jsonResp := `[
				{"is_dir":false,"name":"php-8.4.4-cli-linux-x86_64.tar.gz","full_path":"/static-php-cli/bulk/php-8.4.4-cli-linux-x86_64.tar.gz"}
			]`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &PhpHandler{}

	// 1. Direct match
	info, err := h.ResolveTag(context.Background(), "8.4.4")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "8.4.4", info.Version)

	// 2. Fallback resolution for a patch not explicitly in list (e.g. 8.4.3)
	infoFallback, err := h.ResolveTag(context.Background(), "8.4.3")
	assert.NoError(t, err)
	assert.NotNil(t, infoFallback)
	assert.Equal(t, "8.4.3", infoFallback.Version)

	hasWin := false
	hasUnix := false
	for _, a := range infoFallback.Assets {
		if a.OS == "windows" {
			hasWin = true
		}
		if a.OS == "linux" || a.OS == "darwin" {
			hasUnix = true
		}
	}
	assert.True(t, hasWin)
	assert.True(t, hasUnix)
}

func TestPhpHandler_ResolveVersions_Failures(t *testing.T) {
	ClearPhpHandlerCache()
	defer ClearPhpHandlerCache()

	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	// 1. HTTP 500
	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("Internal Error")),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &PhpHandler{}
	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)

	// 2. Invalid JSON
	ClearPhpHandlerCache()
	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("not json")),
				Header:     make(http.Header),
			}, nil
		},
	}
	_, err = h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
}
