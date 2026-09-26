// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// GitHubReleaseQuerySpec defines a tool and version tag to query via GraphQL.
type GitHubReleaseQuerySpec struct {
	Tool string // e.g. "astral-sh/ruff" or "github:astral-sh/ruff"
	Tag  string // e.g. "0.16.6" or "v0.16.6"
}

type graphQLRequest struct {
	Query string `json:"query"`
}

type graphQLResponse struct {
	Data   map[string]*graphQLRepo `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLRepo struct {
	R1 *graphQLRelease `json:"r1"`
	R2 *graphQLRelease `json:"r2"`
}

type graphQLRelease struct {
	TagName       string `json:"tagName"`
	Name          string `json:"name"`
	IsPrerelease  bool   `json:"isPrerelease"`
	CreatedAt     string `json:"createdAt"`
	ReleaseAssets struct {
		Nodes []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"downloadUrl"`
			Size        int64  `json:"size"`
		} `json:"nodes"`
	} `json:"releaseAssets"`
}

// buildGraphQLQuery constructs a GraphQL query batch for up to N tools.
func buildGraphQLQuery(specs []GitHubReleaseQuerySpec) (string, map[string]GitHubReleaseQuerySpec) {
	aliasMap := make(map[string]GitHubReleaseQuerySpec)
	var sb strings.Builder
	sb.WriteString("query BatchReleases {\n")

	idx := 0
	seen := make(map[string]bool)
	for _, spec := range specs {
		tool := strings.TrimPrefix(spec.Tool, "github:")
		parts := strings.Split(tool, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		tag := strings.TrimSpace(spec.Tag)
		if tag == "" {
			continue
		}

		key := fmt.Sprintf("%s@%s", tool, tag)
		if seen[key] {
			continue
		}
		seen[key] = true

		alias := fmt.Sprintf("t%d", idx)
		aliasMap[alias] = GitHubReleaseQuerySpec{Tool: tool, Tag: tag}
		idx++

		tag1 := tag
		var tag2 string
		if strings.HasPrefix(tag, "v") {
			tag2 = strings.TrimPrefix(tag, "v")
		} else {
			tag2 = "v" + tag
		}

		owner := parts[0]
		repo := parts[1]

		fmt.Fprintf(&sb, `  %s: repository(owner: %q, name: %q) {
    r1: release(tagName: %q) {
      tagName name isPrerelease createdAt
      releaseAssets(first: 100) { nodes { name downloadUrl size } }
    }
    r2: release(tagName: %q) {
      tagName name isPrerelease createdAt
      releaseAssets(first: 100) { nodes { name downloadUrl size } }
    }
  }
`, alias, owner, repo, tag1, tag2)
	}

	sb.WriteString("}\n")
	return sb.String(), aliasMap
}

// BatchPrefetchReleases fetches multiple GitHub releases using GitHub GraphQL API
// in a single HTTP request and populates the releaseCache and disk cache.
// If no token is configured or GraphQL fails, it safely logs and returns nil so callers
// can fall back to normal on-demand resolution.
func (g *GitHubBackend) BatchPrefetchReleases(ctx context.Context, specs []GitHubReleaseQuerySpec) error {
	if len(specs) == 0 {
		return nil
	}

	token := resolveGitHubToken("github.com")
	if token == "" {
		logger.Debug("BatchPrefetchReleases: skipping GraphQL batch prefetch because no token is available", nil)
		return nil
	}

	// Filter out specs that are already cached in memory or on disk
	var needed []GitHubReleaseQuerySpec
	for _, s := range specs {
		tool := strings.TrimPrefix(s.Tool, "github:")
		tag := strings.TrimSpace(s.Tag)
		cacheKey := tool + "@" + tag
		if _, ok := g.releaseCache.Load(cacheKey); ok {
			continue
		}
		if g.client.Transport == nil {
			if rel := readReleaseDiskCache(tool, tag); rel != nil {
				g.releaseCache.Store(cacheKey, rel)
				continue
			}
		}
		needed = append(needed, GitHubReleaseQuerySpec{Tool: tool, Tag: tag})
	}

	if len(needed) == 0 {
		return nil
	}

	// Process in batches of 40 to avoid GraphQL complexity limits
	const batchSize = 40
	for i := 0; i < len(needed); i += batchSize {
		end := i + batchSize
		if end > len(needed) {
			end = len(needed)
		}
		batch := needed[i:end]
		if err := g.fetchGraphQLBatch(ctx, token, batch); err != nil {
			logger.Warn("BatchPrefetchReleases: GraphQL batch fetch had errors, continuing with fallback", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	return nil
}

func (g *GitHubBackend) fetchGraphQLBatch(ctx context.Context, token string, batch []GitHubReleaseQuerySpec) error {
	query, aliasMap := buildGraphQLQuery(batch)
	if len(aliasMap) == 0 {
		return nil
	}

	reqBody, err := json.Marshal(graphQLRequest{Query: query})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GraphQL HTTP %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return err
	}

	for alias, repoData := range gqlResp.Data {
		spec, ok := aliasMap[alias]
		if !ok || repoData == nil {
			continue
		}

		relData := repoData.R1
		if relData == nil {
			relData = repoData.R2
		}
		if relData == nil {
			continue
		}

		var assets []CommonAsset
		for _, a := range relData.ReleaseAssets.Nodes {
			assets = append(assets, CommonAsset{
				Name: a.Name,
				URL:  a.DownloadURL,
				Size: a.Size,
			})
		}

		var publishedAt time.Time
		if relData.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, relData.CreatedAt); err == nil {
				publishedAt = t
			}
		}

		commonRel := &CommonRelease{
			Tag:         relData.TagName,
			Prerelease:  relData.IsPrerelease,
			Assets:      assets,
			PublishedAt: publishedAt,
		}

		tool := spec.Tool
		tag := spec.Tag

		// Store under requested tool@tag
		g.releaseCache.Store(tool+"@"+tag, commonRel)
		if g.client.Transport == nil {
			writeReleaseDiskCache(tool, tag, commonRel)
		}

		// Also store under actual relData.TagName if different
		if relData.TagName != "" && relData.TagName != tag {
			g.releaseCache.Store(tool+"@"+relData.TagName, commonRel)
			if g.client.Transport == nil {
				writeReleaseDiskCache(tool, relData.TagName, commonRel)
			}
		}

		logger.Debug("GraphQL batch prefetch stored release", map[string]interface{}{
			"tool":   tool,
			"tag":    tag,
			"assets": len(assets),
		})
	}

	return nil
}
