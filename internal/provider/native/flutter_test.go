// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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

func TestFlutterHandler_ResolveVersions(t *testing.T) {
	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			resp := `{
				"base_url": "https://storage.googleapis.com/flutter_infra_release/releases",
				"releases": [
					{"hash": "123", "channel": "stable", "version": "3.10.0", "archive": "flutter_macos_3.10.0-stable.zip", "dart_sdk_version": "3.0.0"}
				]
			}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(resp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &FlutterHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "3.10.0", versions[0].Version)
	assert.Len(t, versions[0].Assets, 1)

	// Flutter asset OS depends on the platform running the test
	osName := runtime.GOOS
	assert.Equal(t, osName, versions[0].Assets[0].OS)
}
