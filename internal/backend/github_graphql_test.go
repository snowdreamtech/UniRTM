// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockGraphQLRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockGraphQLRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestBuildGraphQLQuery(t *testing.T) {
	specs := []GitHubReleaseQuerySpec{
		{Tool: "github:astral-sh/ruff", Tag: "0.16.6"},
		{Tool: "cli/cli", Tag: "v2.100.0"},
		{Tool: "invalid-spec", Tag: "1.0.0"},
	}

	query, aliasMap := buildGraphQLQuery(specs)

	if len(aliasMap) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(aliasMap))
	}

	if !strings.Contains(query, `query BatchReleases`) {
		t.Error("query missing BatchReleases header")
	}
	if !strings.Contains(query, `repository(owner: "astral-sh", name: "ruff")`) {
		t.Error("query missing astral-sh/ruff")
	}
	if !strings.Contains(query, `release(tagName: "0.16.6")`) {
		t.Error("query missing release tag 0.16.6")
	}
	if !strings.Contains(query, `release(tagName: "v0.16.6")`) {
		t.Error("query missing release tag v0.16.6")
	}
	if !strings.Contains(query, `repository(owner: "cli", name: "cli")`) {
		t.Error("query missing cli/cli")
	}
}

func TestBatchPrefetchReleases_Success(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-test-token")

	mockRespBody := `{
  "data": {
    "t0": {
      "r1": {
        "tagName": "0.16.6",
        "name": "Ruff 0.16.6",
        "isPrerelease": false,
        "createdAt": "2026-01-01T00:00:00Z",
        "releaseAssets": {
          "nodes": [
            {
              "name": "ruff-darwin-arm64.tar.gz",
              "downloadUrl": "https://github.com/astral-sh/ruff/releases/download/0.16.6/ruff-darwin-arm64.tar.gz",
              "size": 1024
            }
          ]
        }
      },
      "r2": null
    },
    "t1": {
      "r1": null,
      "r2": {
        "tagName": "v2.100.0",
        "name": "gh 2.100.0",
        "isPrerelease": false,
        "createdAt": "2026-01-02T00:00:00Z",
        "releaseAssets": {
          "nodes": [
            {
              "name": "gh_2.100.0_macOS_arm64.zip",
              "downloadUrl": "https://github.com/cli/cli/releases/download/v2.100.0/gh_2.100.0_macOS_arm64.zip",
              "size": 2048
            }
          ]
        }
      }
    }
  }
}`

	callCount := 0
	rt := &mockGraphQLRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if req.URL.String() != "https://api.github.com/graphql" {
				t.Fatalf("expected call to https://api.github.com/graphql, got %s", req.URL.String())
			}
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", req.Method)
			}
			auth := req.Header.Get("Authorization")
			if auth != "Bearer fake-test-token" {
				t.Fatalf("expected Bearer fake-test-token, got %s", auth)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(mockRespBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	backend := NewGitHubBackend()
	backend.client.Transport = rt

	ctx := context.Background()
	specs := []GitHubReleaseQuerySpec{
		{Tool: "astral-sh/ruff", Tag: "0.16.6"},
		{Tool: "cli/cli", Tag: "2.100.0"},
	}

	err := backend.BatchPrefetchReleases(ctx, specs)
	if err != nil {
		t.Fatalf("BatchPrefetchReleases failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("expected exactly 1 GraphQL request, got %d", callCount)
	}

	// Verify both tools are now in cache and require 0 additional network calls
	rel1, err := backend.FetchReleaseByTag(ctx, "astral-sh/ruff", "0.16.6")
	if err != nil {
		t.Fatalf("FetchReleaseByTag ruff failed: %v", err)
	}
	if rel1.Tag != "0.16.6" || len(rel1.Assets) != 1 || rel1.Assets[0].Name != "ruff-darwin-arm64.tar.gz" {
		t.Fatalf("unexpected cached release data for ruff: %+v", rel1)
	}

	rel2, err := backend.FetchReleaseByTag(ctx, "cli/cli", "2.100.0")
	if err != nil {
		t.Fatalf("FetchReleaseByTag cli failed: %v", err)
	}
	if rel2.Tag != "v2.100.0" || len(rel2.Assets) != 1 || rel2.Assets[0].Name != "gh_2.100.0_macOS_arm64.zip" {
		t.Fatalf("unexpected cached release data for cli: %+v", rel2)
	}

	// Still only 1 call made!
	if callCount != 1 {
		t.Fatalf("expected call count to remain 1 after FetchReleaseByTag hits, got %d", callCount)
	}
}

func TestBatchPrefetchReleases_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	backend := NewGitHubBackend()
	ctx := context.Background()
	specs := []GitHubReleaseQuerySpec{
		{Tool: "astral-sh/ruff", Tag: "0.16.6"},
	}

	// Should return nil without error and not panic
	err := backend.BatchPrefetchReleases(ctx, specs)
	if err != nil {
		t.Fatalf("expected nil when no token, got %v", err)
	}
}
