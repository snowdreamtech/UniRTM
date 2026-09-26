// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGitLabGraphQL_BatchPrefetchReleases(t *testing.T) {
	b := NewGitlabBackend()
	var requestCount int

	b.client.Transport = &mockCargoTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			requestCount++
			if !strings.Contains(req.URL.Path, "graphql") {
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			}

			// Simulate GraphQL batch response for two projects
			body := `{
				"data": {
					"t0": {
						"releases": {
							"nodes": [
								{
									"tagName": "v1.2.3",
									"releasedAt": "2024-01-01T00:00:00Z",
									"assets": {
										"links": {
											"nodes": [
												{
													"name": "tool_linux_amd64.tar.gz",
													"url": "https://gitlab.com/download/tool_linux_amd64.tar.gz"
												}
											]
										}
									}
								}
							]
						}
					},
					"t1": {
						"releases": {
							"nodes": [
								{
									"tagName": "v2.0.0",
									"releasedAt": "2024-02-01T00:00:00Z",
									"assets": {
										"links": {
											"nodes": [
												{
													"name": "cli_darwin_arm64.tar.gz",
													"url": "https://gitlab.com/download/cli_darwin_arm64.tar.gz"
												}
											]
										}
									}
								}
							]
						}
					}
				}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	ctx := context.Background()
	specs := []GitLabReleaseQuerySpec{
		{Tool: "group1/proj1", Tag: "1.2.3"},
		{Tool: "group2/proj2", Tag: "2.0.0"},
	}

	// 1. Batch prefetch
	err := b.BatchPrefetchReleases(ctx, specs)
	if err != nil {
		t.Fatalf("unexpected error during batch prefetch: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 GraphQL request, got %d", requestCount)
	}

	// 2. FetchReleases should now hit cache without extra HTTP requests
	rels1, err := b.FetchReleases(ctx, "group1/proj1")
	if err != nil || len(rels1) != 1 {
		t.Fatalf("FetchReleases failed for group1/proj1: %v", err)
	}
	if rels1[0].Tag != "v1.2.3" {
		t.Errorf("expected tag v1.2.3, got %s", rels1[0].Tag)
	}

	// 3. FetchReleaseByTag should also hit cache
	rel2, err := b.FetchReleaseByTag(ctx, "group2/proj2", "2.0.0")
	if err != nil || rel2 == nil {
		t.Fatalf("FetchReleaseByTag failed for group2/proj2: %v", err)
	}
	if rel2.Tag != "v2.0.0" {
		t.Errorf("expected tag v2.0.0, got %s", rel2.Tag)
	}

	// Still only 1 HTTP request had been made!
	if requestCount != 1 {
		t.Errorf("expected still 1 request, got %d", requestCount)
	}
}
