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

func TestGithubHandler_Name(t *testing.T) {
	h := &GithubHandler{}
	assert.Equal(t, "github_release", h.Name())
}

func TestGithubHandler_ResolveVersions(t *testing.T) {
	// Setup mock transport using mockRoundTripper from recipes_test.go
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			url := req.URL.String()
			if url == "https://api.github.com/repos/owner/repo/releases?per_page=20" {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(`[
					{"tag_name":"v1.2.3", "assets": [{"name": "repo-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl"}]}, 
					{"tag_name":"v1.2.0", "assets": [{"name": "repo-darwin-amd64.tar.gz", "browser_download_url": "http://example.com/dl2"}]}
				]`))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(`Not found`))}, nil
		},
	}
	
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &GithubHandler{Owner: "owner", Repo: "repo"}

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "1.2.3", versions[0].Version)
	assert.Equal(t, "1.2.0", versions[1].Version)
}
