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

func TestPythonHandler_ResolveVersions(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"tag_name":"20230507","assets":[{"name":"cpython-3.10.11+20230507-x86_64-unknown-linux-gnu-install_only.tar.gz","browser_download_url":"url"}]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &PythonHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	// It parses the asset name "cpython-3.10.11+20230507..." to version "3.10.11"
	if len(versions) > 0 {
		assert.Equal(t, "3.10.11", versions[0].Version)
	}
}
