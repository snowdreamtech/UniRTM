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

func TestJavaHandler_ResolveVersions(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

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

func TestJavaHandler_ResolveVersions_Failures(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewBufferString(`Error`))}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	h := &JavaHandler{}
	versions, err := h.ResolveVersions(context.Background(), "")
	// It continues on error so it shouldn't return error
	assert.NoError(t, err)
	assert.Len(t, versions, 0)
}

func TestJavaHandler_ConcurrentDeduplication(t *testing.T) {
	ClearJavaHandlerCache()
	defer ClearJavaHandlerCache()

	var reqCount int
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			reqCount++
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
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			versions, err := h.ResolveVersions(context.Background(), "")
			assert.NoError(t, err)
			assert.Len(t, versions, 5)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// 5 major versions queried in single execution = 5 requests total instead of 5 * 5 = 25
	assert.Equal(t, 5, reqCount, "Expected exactly 5 requests (one per major version) for all concurrent callers")
}

