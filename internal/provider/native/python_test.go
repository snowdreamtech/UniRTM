// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"runtime"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestPythonHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `[
				{
					"tag_name": "20230507",
					"assets": [
						{"name": "cpython-3.11.3+20230507-x86_64-unknown-linux-gnu-install_only.tar.gz", "browser_download_url": "https://example.com/python.tar.gz"}
					]
				}
			]`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(resp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &PythonHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "3.11.3", versions[0].Version)

	// Test detectPlatform locally
	osName, arch := h.detectPlatform("cpython-3.11.3+20230507-x86_64-unknown-linux-gnu-install_only.tar.gz")
	assert.Equal(t, "linux", osName)
	assert.Equal(t, "amd64", arch)

	osName, arch = h.detectPlatform("cpython-3.11.3-aarch64-apple-darwin-install_only.tar.gz")
	assert.Equal(t, "darwin", osName)
	assert.Equal(t, "arm64", arch)

	// Ensure the returned assets match current os/arch if they happen to match it
	// Ensure the returned assets match current os/arch if they happen to match it
	if env.RuntimeGOOS == "linux" && runtime.GOARCH == "amd64" {
		assert.Len(t, versions[0].Assets, 1)
	}
}

func TestPythonHandler_detectPlatform(t *testing.T) {
	h := &PythonHandler{}

	tests := []struct {
		filename string
		wantOS   string
		wantArch string
	}{
		{"app.dmg", "", ""},
		{"app.pkg", "", ""},
		{"app.msi", "", ""},
		{"app.deb", "", ""},
		{"app.rpm", "", ""},
		{"app-linux-amd64.tar.gz", "linux", "amd64"},
		{"app-darwin-aarch64.tar.gz", "darwin", "arm64"},
		{"app-macos-arm64.tar.gz", "darwin", "arm64"},
		{"app-apple-m1.tar.gz", "darwin", ""},
		{"app-mac-x64.tar.gz", "darwin", "amd64"},
		{"app-windows-x86_64.zip", "windows", "amd64"},
		{"app-win-386.zip", "windows", "386"},
		{"app-linux-i686.tar.gz", "linux", "386"},
		{"app-linux-x86.tar.gz", "linux", "386"},
	}

	for _, tt := range tests {
		os, arch := h.detectPlatform(tt.filename)
		assert.Equal(t, tt.wantOS, os, "filename: %s", tt.filename)
		assert.Equal(t, tt.wantArch, arch, "filename: %s", tt.filename)
	}
}

func TestPythonHandler_ResolveVersions_Failures(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Internal Error`))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &PythonHandler{Owner: "owner", Repo: "repo"}

	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "github api call failed after 3 attempts")
}

func TestPythonHandler_ResolveVersions_InvalidJSON(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`invalid json`))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &PythonHandler{Owner: "owner", Repo: "repo"}

	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
}

func TestPythonHandler_ResolveVersions_FilterAssets(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `[
				{
					"tag_name": "20230507",
					"assets": [
						{"name": "cpython-3.11.3+20230507-x86_64-unknown-linux-gnu-debug.tar.gz", "browser_download_url": "https://example.com/python-debug.tar.gz"},
						{"name": "cpython-3.11.3+20230507-x86_64-unknown-linux-gnu-pgo+lto.tar.gz", "browser_download_url": "https://example.com/python-pgo.tar.gz"},
						{"name": "cpython-3.11.3+20230507-x86_64-unknown-linux-gnu.tar.gz.asc", "browser_download_url": "https://example.com/python.tar.gz.asc"}
					]
				}
			]`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(resp))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &PythonHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	// Debug asset must be filtered out, leaving only the pgo+lto asset for linux/amd64
	assert.Len(t, versions[0].Assets, 1)
	assert.Equal(t, "cpython-3.11.3+20230507-x86_64-unknown-linux-gnu-pgo+lto.tar.gz", versions[0].Assets[0].Filename)
}

func TestPythonHandler_ResolveVersions_WindowsDebugRejection(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Debug comes alphabetically before install_only and pgo
			resp := `[
				{
					"tag_name": "20260929",
					"assets": [
						{"name": "cpython-3.14.7+20260929-x86_64-pc-windows-msvc-debug-full.tar.zst", "browser_download_url": "https://example.com/python-debug.tar.zst"},
						{"name": "cpython-3.14.7+20260929-x86_64-pc-windows-msvc-install_only.tar.gz", "browser_download_url": "https://example.com/python-install-only.tar.gz"},
						{"name": "cpython-3.14.7+20260929-x86_64-pc-windows-msvc-pgo-full.tar.zst", "browser_download_url": "https://example.com/python-pgo.tar.zst"}
					]
				}
			]`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(resp))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &PythonHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	// Must pick install_only over debug-full and pgo-full for windows/amd64
	assert.Len(t, versions[0].Assets, 1)
	assert.Equal(t, "windows", versions[0].Assets[0].OS)
	assert.Equal(t, "amd64", versions[0].Assets[0].Arch)
	assert.Equal(t, "cpython-3.14.7+20260929-x86_64-pc-windows-msvc-install_only.tar.gz", versions[0].Assets[0].Filename)
}

func TestPythonAssetPreference(t *testing.T) {
	assert.Equal(t, -1, pythonAssetPreference("cpython-3.14.7-windows-msvc-debug-full.tar.zst"))
	assert.Equal(t, 100, pythonAssetPreference("cpython-3.14.7-windows-msvc-install_only.tar.gz"))
	assert.Equal(t, 95, pythonAssetPreference("cpython-3.14.7-windows-msvc-install_only_stripped.tar.gz"))
	assert.Equal(t, 85, pythonAssetPreference("cpython-3.14.7-windows-msvc-pgo-full.tar.zst"))
	assert.Equal(t, 70, pythonAssetPreference("cpython-3.14.7-windows-msvc-shared-install_only.tar.gz"))
}
