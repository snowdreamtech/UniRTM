package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"runtime"
	"testing"

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
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		assert.Len(t, versions[0].Assets, 1)
	}
}
