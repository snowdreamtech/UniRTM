package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	unirtmhttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("mocked response")),
		Header:     make(http.Header),
	}, nil
}

func TestGetBuiltinRecipes(t *testing.T) {
	recipes := GetBuiltinRecipes()
	if len(recipes) == 0 {
		t.Error("Expected built-in recipes to be populated")
	}

	tests := []struct {
		toolName    string
		expectedID  string
		expectedURL string
		hasAliases  bool
		hasGPG      bool
	}{
		{"go", "go", "https://go.dev/dl", true, true},
		{"nodejs", "node", "https://nodejs.org/dist", false, true},
		{"python", "python", "https://github.com/astral-sh/python-build-standalone/releases", true, false},
		{"rust", "rust", "https://static.rust-lang.org/dist", true, true},
		{"java", "java", "https://api.adoptium.net", false, true},
		{"ruby", "ruby", "https://github.com/ruby/ruby-builder/releases", false, false},
		{"bun", "bun", "https://github.com/oven-sh/bun/releases", false, false},
		{"cmake", "cmake", "https://github.com/Kitware/CMake/releases", false, true},
		{"helm", "helm", "https://github.com/helm/helm/releases", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			recipe, ok := recipes[tt.toolName]
			if !ok {
				t.Fatalf("Expected recipe for %s", tt.toolName)
			}
			assert.Equal(t, tt.expectedID, recipe.ID)
			assert.Equal(t, tt.expectedURL, recipe.BaseURL)
			if tt.hasAliases {
				assert.NotEmpty(t, recipe.Aliases)
			}
			if tt.hasGPG {
				assert.NotEmpty(t, recipe.GPGKeys)
			}
		})
	}
}

func TestHandlerNames(t *testing.T) {
	tests := []struct {
		handler interface{ Name() string }
		name    string
	}{
		{&GolangHandler{}, "golang"},
		{&NodeJSHandler{}, "nodejs"},
		{&PythonHandler{}, "python_standalone"},
		{&RustHandler{}, "rust"},
		{&GithubHandler{Owner: "oven-sh", Repo: "bun"}, "github_release"},
		{&JavaHandler{}, "java"},
		{&RubyHandler{}, "ruby"},
		{&MavenHandler{}, "maven"},
		{&GradleHandler{}, "gradle"},
		{&CMakeHandler{}, "cmake"},
		{&NinjaHandler{}, "ninja"},
		{&ElixirHandler{}, "elixir"},
		{&ErlangHandler{}, "erlang"},
		{&FlutterHandler{}, "flutter"},
		{&JuliaHandler{}, "julia"},
		{&HelmHandler{}, "helm"},
		{&KubectlHandler{}, "kubectl"},
		{&ZigHandler{}, "zig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.handler.Name())
		})
	}
}

func TestGolangHandler_ResolveVersions(t *testing.T) {
	oldMock := unirtmhttp.MockTransport
	defer func() { unirtmhttp.MockTransport = oldMock }()

	unirtmhttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `[
				{
					"version": "go1.20",
					"stable": true,
					"files": [
						{"filename": "go1.20.linux-amd64.tar.gz", "os": "linux", "arch": "amd64", "version": "go1.20", "sha256": "abcdef", "size": 123456, "kind": "archive"}
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

	handler := &GolangHandler{}
	versions, err := handler.ResolveVersions(context.Background(), "https://go.dev/dl")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "1.20", versions[0].Version)
	assert.True(t, versions[0].IsLTS)
	assert.Len(t, versions[0].Assets, 1)
	assert.Equal(t, "go1.20.linux-amd64.tar.gz", versions[0].Assets[0].Filename)
}

func TestJuliaHandler_ResolveVersions(t *testing.T) {
	oldMock := unirtmhttp.MockTransport
	defer func() { unirtmhttp.MockTransport = oldMock }()

	unirtmhttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `{
				"1.9.0": {
					"version": "1.9.0",
					"files": [
						{"os": "mac", "arch": "aarch64", "kind": "archive", "url": "https://example.com/julia.tar.gz", "sha256": "abcdef", "asc": "signed"}
					]
				}
			}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(resp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	handler := &JuliaHandler{}
	versions, err := handler.ResolveVersions(context.Background(), "https://julialang-s3.julialang.org/bin")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "1.9.0", versions[0].Version)
	assert.Len(t, versions[0].Assets, 1)
	assert.Equal(t, "darwin", versions[0].Assets[0].OS)
	assert.Equal(t, "arm64", versions[0].Assets[0].Arch)
}

func TestMapPlatform(t *testing.T) {
	os, arch := mapPlatform("mac", "aarch64")
	assert.Equal(t, "darwin", os)
	assert.Equal(t, "arm64", arch)

	os, arch = mapPlatform("win", "x86_64")
	assert.Equal(t, "windows", os)
	assert.Equal(t, "amd64", arch)

	os, arch = mapPlatform("linux", "i686")
	assert.Equal(t, "linux", os)
	assert.Equal(t, "386", arch)
}
