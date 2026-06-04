// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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

func TestKubectlHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("v1.28.0\n")),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &KubectlHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "v1.28.0", versions[0].Version)
	assert.NotEmpty(t, versions[0].Assets)
}

func TestKubectlHandler_ResolveVersions_Failures(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Internal Error`))}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &KubectlHandler{}
	_, err := h.ResolveVersions(context.Background(), "")
	assert.Error(t, err)

	// Invalid body isn't explicitly tested for kubectl because it's just strings.Split, so it's a valid empty list.
}
