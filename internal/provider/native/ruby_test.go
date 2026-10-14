// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRubyHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			osName := env.RuntimeGOOS
			if osName == "darwin" {
				osName = "macos"
			} else if osName == "linux" {
				osName = "ubuntu"
			}
			filename := fmt.Sprintf("ruby-3.2.0-%s-%s.tar.gz", osName, runtime.GOARCH)
			resp := fmt.Sprintf(`[
				{
					"tag_name": "v3.2.0",
					"assets": [
						{"name": "%s", "browser_download_url": "https://example.com/%s"}
					]
				}
			]`, filename, filename)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(resp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &RubyHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, versions, 1)
	assert.Equal(t, "3.2.0", versions[0].Version)
	assert.Len(t, versions[0].Assets, 1)

	// test isMatch
	osName2 := env.RuntimeGOOS
	if osName2 == "darwin" {
		osName2 = "macos"
	} else if osName2 == "linux" {
		osName2 = "ubuntu"
	}
	filename := fmt.Sprintf("ruby-3.2.0-%s-%s.tar.gz", osName2, runtime.GOARCH)
	assert.True(t, h.isMatch(filename))
}

func TestRubyHandler_isMatch(t *testing.T) {
	h := &RubyHandler{}

	originalOS := env.RuntimeGOOS
	defer func() { env.RuntimeGOOS = originalOS }()

	assert.False(t, h.isMatch("python-3.0.0-ubuntu-amd64.tar.gz"))

	env.RuntimeGOOS = "darwin"
	assert.True(t, h.isMatch(fmt.Sprintf("ruby-3.0-macos-%s.tar.gz", runtime.GOARCH)))
	assert.False(t, h.isMatch(fmt.Sprintf("ruby-3.0-ubuntu-%s.tar.gz", runtime.GOARCH)))

	env.RuntimeGOOS = "linux"
	assert.True(t, h.isMatch(fmt.Sprintf("ruby-3.0-ubuntu-%s.tar.gz", runtime.GOARCH)))
	assert.False(t, h.isMatch(fmt.Sprintf("ruby-3.0-macos-%s.tar.gz", runtime.GOARCH)))

	env.RuntimeGOOS = "windows"
	assert.True(t, h.isMatch(fmt.Sprintf("ruby-3.0-windows-%s.tar.gz", runtime.GOARCH)))
	assert.False(t, h.isMatch(fmt.Sprintf("ruby-3.0-macos-%s.tar.gz", runtime.GOARCH)))

	env.RuntimeGOOS = "freebsd"
	assert.False(t, h.isMatch(fmt.Sprintf("ruby-3.0-freebsd-%s.tar.gz", env.RuntimeGOARCH)))

	// Test Arch
	env.RuntimeGOOS = "linux"
	originalArch := env.RuntimeGOARCH
	defer func() { env.RuntimeGOARCH = originalArch }()

	env.RuntimeGOARCH = "arm64"
	assert.True(t, h.isMatch("ruby-3.0-ubuntu-arm64.tar.gz"))
	assert.False(t, h.isMatch("ruby-3.0-ubuntu-amd64.tar.gz"))

	env.RuntimeGOARCH = "amd64"
	assert.True(t, h.isMatch("ruby-3.0-ubuntu-amd64.tar.gz"))
	assert.True(t, h.isMatch("ruby-3.0-ubuntu-x86_64.tar.gz"))
	assert.False(t, h.isMatch("ruby-3.0-ubuntu-arm64.tar.gz"))
	assert.False(t, h.isMatch("ruby-3.0-ubuntu.tar.gz")) // amd64 requires arch in name usually unless handled otherwise, wait, actually if not present it returns false
}

func TestRubyHandler_ResolveTag(t *testing.T) {
	ClearGithubHandlerCache()
	defer ClearGithubHandlerCache()

	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "ruby-3.4.1") {
				resp := `{
					"tag_name": "ruby-3.4.1",
					"assets": [
						{"name": "ruby-3.4.1-darwin-arm64.tar.gz", "browser_download_url": "https://example.com/darwin-arm64"},
						{"name": "ruby-3.4.1-darwin-x64.tar.gz", "browser_download_url": "https://example.com/darwin-x64"},
						{"name": "ruby-3.4.1-ubuntu-22.04-arm64.tar.gz", "browser_download_url": "https://example.com/ubuntu-arm64"},
						{"name": "ruby-3.4.1-ubuntu-22.04-x64.tar.gz", "browser_download_url": "https://example.com/ubuntu-x64"}
					]
				}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(resp)),
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString("Not Found"))}, nil
		},
	}

	h := &RubyHandler{}
	info, err := h.ResolveTag(context.Background(), "3.4.1")
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "3.4.1", info.Version)
	assert.Len(t, info.Assets, 4)

	// Verify all platforms were detected
	platforms := make(map[string]bool)
	for _, a := range info.Assets {
		platforms[a.OS+"-"+a.Arch] = true
	}
	assert.True(t, platforms["darwin-arm64"])
	assert.True(t, platforms["darwin-amd64"])
	assert.True(t, platforms["linux-arm64"])
	assert.True(t, platforms["linux-amd64"])
}

