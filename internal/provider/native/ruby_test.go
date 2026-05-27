package native

import (
	"context"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestRubyHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			osName := runtime.GOOS
			if osName == "darwin" {
				osName = "macos"
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
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "3.2.0", versions[0].Version)
	assert.Len(t, versions[0].Assets, 1)

	// test isMatch
	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "macos"
	}
	filename := fmt.Sprintf("ruby-3.2.0-%s-%s.tar.gz", osName, runtime.GOARCH)
	assert.True(t, h.isMatch(filename))
	assert.False(t, h.isMatch("ruby-3.2.0-linux-amd64.zip"))
	assert.False(t, h.isMatch("ruby-3.2.0-preview1-linux-amd64.tar.gz"))
}
