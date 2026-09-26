// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

type ComposerBackend struct {
	client        *http.Client
	requestGroup  singleflight.Group
	versionsCache sync.Map // map[string][]string (tool -> sorted version strings)
}

func NewComposerBackend() *ComposerBackend {
	return &ComposerBackend{
		client: pkgHttp.NewClientWithTimeout(15 * time.Second),
	}
}

func (b *ComposerBackend) Name() string {
	return "composer"
}

func (b *ComposerBackend) Dependencies() []string {
	return []string{"php"}
}

type packagistResponse struct {
	Package struct {
		Versions map[string]interface{} `json:"versions"`
	} `json:"package"`
}

func (b *ComposerBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
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

		url := fmt.Sprintf("https://packagist.org/packages/%s.json", tool)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, NewBackendError(b.Name(), tool, "package not found on packagist", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var data packagistResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for v := range data.Package.Versions {
			versionStrs = append(versionStrs, v)
		}

		// Sort versions (roughly newest first)
		sort.Slice(versionStrs, func(i, j int) bool {
			return versionStrs[i] > versionStrs[j]
		})

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

func (b *ComposerBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *ComposerBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *ComposerBackend) SupportsChecksum() bool {
	return true
}

func (b *ComposerBackend) SupportsGPG() bool {
	return false
}

func (b *ComposerBackend) AttestationType() string {
	return ""
}

func (b *ComposerBackend) IsRecommended() bool {
	return true
}

func (b *ComposerBackend) IsScriptless() bool {
	return true
}

func (b *ComposerBackend) GetReach() string {
	return "Large"
}

func (b *ComposerBackend) IsStable() bool {
	return true
}

func (b *ComposerBackend) SupportsOffline() bool {
	return true
}
