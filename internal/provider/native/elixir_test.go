package native

import (
	"context"
	"bytes"
	"io"
	"net/http"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestElixirHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `[
				{
					"tag_name": "v1.15.0",
					"assets": [
						{"name": "Precompiled.zip", "browser_download_url": "https://example.com/Precompiled.zip"},
						{"name": "elixir-otp-25.zip", "browser_download_url": "https://example.com/elixir-otp-25.zip"}
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

	h := &ElixirHandler{}
	assert.Equal(t, "elixir", h.Name())

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "1.15.0", versions[0].Version)
	
	// Each asset matching adds 4 universal OS/arch combinations
	assert.Len(t, versions[0].Assets, 8)
}
