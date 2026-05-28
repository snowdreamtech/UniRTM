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

func TestNinjaHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `[
				{
					"tag_name": "v1.11.1",
					"assets": [
						{"name": "ninja-linux.zip", "browser_download_url": "https://example.com/ninja-linux.zip"},
						{"name": "ninja-mac.zip", "browser_download_url": "https://example.com/ninja-mac.zip"},
						{"name": "ninja-win.zip", "browser_download_url": "https://example.com/ninja-win.zip"}
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

	h := &NinjaHandler{}
	assert.Equal(t, "ninja", h.Name())

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "1.11.1", versions[0].Version)

	// ninja creates an asset for each major platform
	assert.Len(t, versions[0].Assets, 3)
}
