// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	mavenFlight        singleflight.Group
	mavenVersionsCache sync.Map // map[string][]string (tool -> version strings)
)

// ClearMavenCache clears the in-memory cache for Maven artifacts. Mainly used for testing.
func ClearMavenCache() {
	mavenVersionsCache.Range(func(key, _ interface{}) bool {
		mavenVersionsCache.Delete(key)
		return true
	})
}

type MavenBackend struct {
	client        *http.Client
	requestGroup  *singleflight.Group
	versionsCache *sync.Map
}

func NewMavenBackend() *MavenBackend {
	return &MavenBackend{
		client:        pkgHttp.NewClientWithTimeout(10 * time.Second),
		requestGroup:  &mavenFlight,
		versionsCache: &mavenVersionsCache,
	}
}

func (b *MavenBackend) Name() string {
	return "maven"
}

func (b *MavenBackend) Dependencies() []string {
	return nil
}

type mavenSearchResponse struct {
	Response struct {
		Docs []struct {
			Version string `json:"v"`
		} `json:"docs"`
	} `json:"response"`
}

func (b *MavenBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
	}

	// tool is expected to be group:artifact
	parts := strings.Split(tool, ":")
	if len(parts) != 2 {
		return nil, NewBackendError(b.Name(), tool, "invalid tool name format, expected group:artifact", nil)
	}

	if val, ok := b.versionsCache.Load(tool); ok {
		versionStrs := val.([]string)
		versions := make([]VersionInfo, len(versionStrs))
		for i, v := range versionStrs {
			versions[i] = VersionInfo{
				Version:  v,
				Platform: platform,
			}
		}
		return versions, nil
	}

	result, err, _ := b.requestGroup.Do("versions:"+tool, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(tool); ok {
			return val, nil
		}

		url := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=g:%s+AND+a:%s&rows=50&core=gav", parts[0], parts[1])

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}
		defer resp.Body.Close()

		var data mavenSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for _, doc := range data.Response.Docs {
			versionStrs = append(versionStrs, doc.Version)
		}

		b.versionsCache.Store(tool, versionStrs)
		return versionStrs, nil
	})

	if err != nil {
		return nil, err
	}

	versionStrs := result.([]string)
	versions := make([]VersionInfo, len(versionStrs))
	for i, v := range versionStrs {
		versions[i] = VersionInfo{
			Version:  v,
			Platform: platform,
		}
	}

	return versions, nil
}

func (b *MavenBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		return &versions[0], nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *MavenBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *MavenBackend) SupportsChecksum() bool {
	return true
}

func (b *MavenBackend) SupportsGPG() bool {
	return true
}

func (b *MavenBackend) AttestationType() string {
	return ""
}

func (b *MavenBackend) IsRecommended() bool {
	return true
}

func (b *MavenBackend) IsScriptless() bool {
	return true
}

func (b *MavenBackend) GetReach() string {
	return "Huge"
}

func (b *MavenBackend) IsStable() bool {
	return true
}

func (b *MavenBackend) SupportsOffline() bool {
	return true
}
