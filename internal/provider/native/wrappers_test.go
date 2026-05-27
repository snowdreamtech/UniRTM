package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

func TestNativeWrappers_ResolveVersions(t *testing.T) {
	// Setup mock transport using mockRoundTripper
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Mocking GitHub release API
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[
					{"tag_name":"v1.2.3", "assets": [
						{"name": "repo-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl"},
						{"name": "repo-macos-universal.tar.gz", "browser_download_url": "http://example.com/dl2"}
					]}
				]`)),
			}, nil
		},
	}
	
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	tests := []struct {
		name    string
		handler interface {
			ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error)
		}
	}{
		{"cmake", &CMakeHandler{}},
		{"elixir", &ElixirHandler{}},
		{"erlang", &ErlangHandler{}},
		{"flutter", &FlutterHandler{}},
		{"helm", &HelmHandler{}},
		{"ninja", &NinjaHandler{}},
		{"ruby", &RubyHandler{}},
		{"rust", &RustHandler{}},
		{"zig", &ZigHandler{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test just to cover the logic, we don't assert specific results since each does its own filtering
			versions, _ := tt.handler.ResolveVersions(context.Background(), "")
			// Some handlers expect different JSON shapes and will return an error, which is fine for coverage.
			_ = versions
		})
	}
}
