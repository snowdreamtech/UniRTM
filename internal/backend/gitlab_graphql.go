// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// GitLabReleaseQuerySpec defines a tool and version tag to query via GraphQL.
type GitLabReleaseQuerySpec struct {
	Tool string // e.g. "inkscape/inkscape" or "gitlab:inkscape/inkscape"
	Tag  string // e.g. "1.0.0" or "v1.0.0"
}

type gitlabGraphQLRequest struct {
	Query string `json:"query"`
}

type gitlabGraphQLResponse struct {
	Data   map[string]*gitlabGraphQLProject `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type gitlabGraphQLProject struct {
	Releases struct {
		Nodes []gitlabGraphQLRelease `json:"nodes"`
	} `json:"releases"`
}

type gitlabGraphQLRelease struct {
	TagName    string `json:"tagName"`
	ReleasedAt string `json:"releasedAt"`
	Assets     struct {
		Links struct {
			Nodes []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"nodes"`
		} `json:"links"`
	} `json:"assets"`
}

func buildGitLabGraphQLQuery(specs []GitLabReleaseQuerySpec) (string, map[string]string) {
	aliasMap := make(map[string]string)
	seen := make(map[string]bool)

	var sb strings.Builder
	sb.WriteString("query BatchGitLabReleases {\n")

	idx := 0
	for _, spec := range specs {
		tool := strings.TrimPrefix(spec.Tool, "gitlab:")
		tool = strings.TrimSpace(tool)
		if tool == "" || seen[tool] {
			continue
		}
		seen[tool] = true

		alias := fmt.Sprintf("t%d", idx)
		aliasMap[alias] = tool
		idx++

		fmt.Fprintf(&sb, `  %s: project(fullPath: %q) {
    releases(first: 10) {
      nodes {
        tagName
        releasedAt
        assets {
          links {
            nodes {
              name
              url
            }
          }
        }
      }
    }
  }
`, alias, tool)
	}

	sb.WriteString("}\n")
	return sb.String(), aliasMap
}

// BatchPrefetchReleases fetches multiple GitLab releases using GitLab's GraphQL API
// in a single HTTP request and populates the releasesCache.
func (b *GitlabBackend) BatchPrefetchReleases(ctx context.Context, specs []GitLabReleaseQuerySpec) error {
	if len(specs) == 0 {
		return nil
	}

	var needed []GitLabReleaseQuerySpec
	for _, s := range specs {
		tool := strings.TrimPrefix(s.Tool, "gitlab:")
		if _, ok := b.releasesCache.Load(tool); !ok {
			needed = append(needed, GitLabReleaseQuerySpec{Tool: tool, Tag: s.Tag})
		}
	}

	if len(needed) == 0 {
		return nil
	}

	query, aliasMap := buildGitLabGraphQLQuery(needed)
	if len(aliasMap) == 0 {
		return nil
	}

	reqBody, err := json.Marshal(gitlabGraphQLRequest{Query: query})
	if err != nil {
		return err
	}

	// Determine GraphQL endpoint from baseURL (e.g. https://gitlab.com/api/v4 -> https://gitlab.com/api/graphql)
	graphqlURL := strings.TrimSuffix(b.baseURL, "/v4") + "/graphql"
	if !strings.HasSuffix(graphqlURL, "/api/graphql") && strings.HasPrefix(b.baseURL, "https://gitlab.com") {
		graphqlURL = "https://gitlab.com/api/graphql"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := env.Get("GITLAB_TOKEN"); token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		logger.Debug("BatchPrefetchReleases: GitLab GraphQL request failed, fallback to REST", map[string]interface{}{"error": err.Error()})
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Debug("BatchPrefetchReleases: GitLab GraphQL status not OK", map[string]interface{}{"status": resp.StatusCode})
		return nil
	}

	var gqlResp gitlabGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil
	}

	for alias, tool := range aliasMap {
		proj, ok := gqlResp.Data[alias]
		if !ok || proj == nil {
			continue
		}

		var releases []CommonRelease
		for _, r := range proj.Releases.Nodes {
			var publishedAt time.Time
			if r.ReleasedAt != "" {
				if t, err := time.Parse(time.RFC3339, r.ReleasedAt); err == nil {
					publishedAt = t
				}
			}

			var assets []CommonAsset
			for _, l := range r.Assets.Links.Nodes {
				assets = append(assets, CommonAsset{
					Name: l.Name,
					URL:  l.URL,
				})
			}

			cr := CommonRelease{
				Tag:         r.TagName,
				Assets:      assets,
				PublishedAt: publishedAt,
			}
			releases = append(releases, cr)
			b.releaseByTagCache.Store(tool+"@"+r.TagName, &cr)
			cleanTag := strings.TrimPrefix(r.TagName, "v")
			b.releaseByTagCache.Store(tool+"@"+cleanTag, &cr)
			b.releaseByTagCache.Store(tool+"@v"+cleanTag, &cr)
		}

		if len(releases) > 0 {
			b.releasesCache.Store(tool, releases)
			logger.Debug("GitLab GraphQL batch prefetch stored releases", map[string]interface{}{
				"tool":  tool,
				"count": len(releases),
			})
		}
	}

	return nil
}
