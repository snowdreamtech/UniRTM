// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestDartHandler_ResolveVersions_Mock(t *testing.T) {
	ClearDartCache()
	defer ClearDartCache()

	oldMock := pkgHttp.MockTransport
	defer func() { pkgHttp.MockTransport = oldMock }()

	pkgHttp.MockTransport = &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			urlStr := req.URL.String()
			if strings.Contains(urlStr, "latest/VERSION") {
				respJSON := `{"version": "3.4.0"}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(respJSON)),
					Header:     make(http.Header),
				}, nil
			}

			// XML bucket listing
			respXML := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://doc.s3.amazonaws.com/2006-03-01">
  <CommonPrefixes><Prefix>channels/stable/release/3.4.0/</Prefix></CommonPrefixes>
  <CommonPrefixes><Prefix>channels/stable/release/3.3.0/</Prefix></CommonPrefixes>
</ListBucketResult>`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(respXML)),
				Header:     make(http.Header),
			}, nil
		},
	}

	h := &DartHandler{}
	assert.Equal(t, "dart", h.Name())

	versions, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.NotEmpty(t, versions)

	foundLatest := false
	found340 := false
	for _, v := range versions {
		if v.Version == "latest" {
			foundLatest = true
		}
		if v.Version == "3.4.0" {
			found340 = true
		}
		assert.NotEmpty(t, v.Assets, "version %s should have assets", v.Version)
	}

	assert.True(t, foundLatest, "expected 'latest' alias to be present")
	assert.True(t, found340, "expected '3.4.0' version to be present")

	// Verify in-memory cache hit
	cached, err := h.ResolveVersions(context.Background(), "")
	assert.NoError(t, err)
	assert.Equal(t, len(versions), len(cached))
}

func TestDartHandler_Live(t *testing.T) {
	if os.Getenv("TEST_NETWORK") == "" {
		t.Skip("Skipping live network test for dart. Set TEST_NETWORK=1 to enable.")
	}

	ClearDartCache()
	defer ClearDartCache()

	h := &DartHandler{}
	assert.Equal(t, "dart", h.Name())

	versions, err := h.ResolveVersions(context.Background(), "")
	if err != nil {
		t.Skipf("skipping live network test for dart due to network failure: %v", err)
		return
	}

	assert.NotEmpty(t, versions)
}
