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

func TestGradleHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"version":"1.0.0", "downloadUrl": "http://example.com", "broken": false, "snapshot": false}]`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &GradleHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "1.0.0", versions[0].Version)
	assert.NotEmpty(t, versions[0].Assets)

	assert.Equal(t, "gradle", h.Name())
	assert.True(t, h.IsMatch("gradle-1.0.0-src.zip", "linux", "amd64"))
	assert.True(t, h.IsMatch("gradle-1.0.0-bin.zip", "linux", "amd64"))
}

func TestMavenHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`<metadata><versioning><versions><version>3.9.5</version></versions></versioning></metadata>`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &MavenHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "3.9.6", versions[0].Version)
	assert.NotEmpty(t, versions[0].Assets)

	assert.Equal(t, "maven", h.Name())
	assert.False(t, h.IsMatch("apache-maven-3.9.6-src.zip", "linux", "amd64"))
	assert.True(t, h.IsMatch("apache-maven-3.9.6-bin.zip", "linux", "amd64"))
}
