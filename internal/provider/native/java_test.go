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

func TestJavaHandler_ResolveVersions(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"release_name":"jdk-17","version_data":{"openjdk_version":"17.0.2+8"},"binaries":[{"package":{"name":"file.tar.gz","link":"url"}}]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	// It will query 5 major versions, returning 5 results with version "17.0.2"
	assert.Len(t, versions, 5)
	assert.Equal(t, "17.0.2", versions[0].Version)
}
