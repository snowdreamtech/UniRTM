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

func TestGithubHandler_Name(t *testing.T) {
	h := &GithubHandler{}
	assert.Equal(t, "github_release", h.Name())
}

func TestGithubHandler_ResolveVersions(t *testing.T) {
	// Setup mock transport using mockRoundTripper from recipes_test.go
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			url := req.URL.String()
			if url == "https://api.github.com/repos/owner/repo/releases?per_page=20" {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`[
					{"tag_name":"v1.2.3", "assets": [{"name": "repo-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl"}]},
					{"tag_name":"v1.2.0", "assets": [{"name": "repo-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl2"}]}
				]`))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(`Not found`))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &GithubHandler{Owner: "owner", Repo: "repo"}

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "1.2.3", versions[0].Version)
	assert.Equal(t, "1.2.0", versions[1].Version)
}

func TestGithubHandler_detectPlatform(t *testing.T) {
	h := &GithubHandler{}

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
		{"app-darwin-amd64.tar.gz", "darwin", "amd64"},
		{"app-macos-aarch64.zip", "darwin", "arm64"},
		{"app-apple-m1.tar.gz", "darwin", ""},
		{"app-mac-x64.tar.gz", "darwin", "amd64"},
		{"app-windows-x86_64.zip", "windows", "amd64"},
		{"app-win-386.zip", "windows", "386"},
		{"app-linux-amd64.tar.gz", "linux", "amd64"},
		{"app-ubuntu-arm64.tar.gz", "linux", "arm64"},
		{"app-centos-i686.tar.gz", "linux", "386"},
		{"app-x86.tar.gz", "linux", "386"},
		{"app-universal.tar.gz", "linux", "universal"},
		{"app-all.zip", "linux", "universal"},
	}

	for _, tt := range tests {
		os, arch := h.detectPlatform(tt.filename)
		assert.Equal(t, tt.wantOS, os, "filename: %s", tt.filename)
		assert.Equal(t, tt.wantArch, arch, "filename: %s", tt.filename)
	}
}

func TestGithubHandler_ResolveVersions_Failures(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Internal Error`))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &GithubHandler{Owner: "owner", Repo: "repo"}

	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "github api call failed after 3 attempts")
}

func TestGithubHandler_ResolveVersions_Signatures(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`[
				{"tag_name":"v1.2.3", "assets": [
					{"name": "app-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl"},
					{"name": "app-darwin-amd64.tar.gz.asc", "browser_download_url": "http://example.com/dl.asc"},
					{"name": "app-linux-amd64.tar.gz", "browser_download_url": "http://example.com/dl2"},
					{"name": "app-linux-amd64.tar.gz.sig", "browser_download_url": "http://example.com/dl2.sig"},
					{"name": "app-windows-amd64.zip", "browser_download_url": "http://example.com/dl3"},
					{"name": "app-windows-amd64.zip.sha256", "browser_download_url": "http://example.com/dl3.sha256"}
				]}
			]`))}, nil
		},
	}

	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &GithubHandler{Owner: "owner", Repo: "repo"}

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Len(t, versions[0].Assets, 3)

	for _, a := range versions[0].Assets {
		if a.OS == "darwin" {
			assert.Equal(t, "http://example.com/dl.asc", a.SignatureURL)
		} else if a.OS == "linux" {
			assert.Equal(t, "http://example.com/dl2.sig", a.SignatureURL)
		} else if a.OS == "windows" {
			assert.Empty(t, a.SignatureURL)
		}
	}
}
