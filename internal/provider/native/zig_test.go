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

func TestZigHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"0.11.0": {"x86_64-linux": {"tarball": "https://ziglang.org/download/0.11.0/zig-linux-x86_64-0.11.0.tar.xz"}}}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &ZigHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "0.11.0", versions[0].Version)
	assert.NotEmpty(t, versions[0].Assets)
}
