// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"golang.org/x/sync/singleflight"
)

var (
	pubFlight        singleflight.Group
	pubVersionsCache sync.Map // map[string][]string (tool -> sorted version strings)
)

// ClearPubCache clears the in-memory cache for Dart Pub packages. Mainly used for testing.
func ClearPubCache() {
	pubVersionsCache.Range(func(key, _ interface{}) bool {
		pubVersionsCache.Delete(key)
		return true
	})
}

type PubBackend struct {
	client        *http.Client
	requestGroup  *singleflight.Group
	versionsCache *sync.Map
}

func NewPubBackend() *PubBackend {
	return &PubBackend{
		client:        pkgHttp.NewClientWithTimeout(15 * time.Second),
		requestGroup:  &pubFlight,
		versionsCache: &pubVersionsCache,
	}
}

func (b *PubBackend) Name() string {
	return "pub"
}

func (b *PubBackend) Dependencies() []string {
	return []string{"dart"}
}

type pubResponse struct {
	Name     string `json:"name"`
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

func (b *PubBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
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

	if b.client.Transport == nil {
		var versionStrs []string
		if readEcosystemMetadataDiskCache("pub", tool, &versionStrs, 10*time.Minute) {
			b.versionsCache.Store(tool, versionStrs)
			versions := make([]VersionInfo, len(versionStrs))
			for i, v := range versionStrs {
				versions[i] = VersionInfo{
					Version:  v,
					Platform: platform,
				}
			}
			return versions, nil
		}
	}

	result, err, _ := b.requestGroup.Do("versions:"+tool, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(tool); ok {
			return val, nil
		}

		url := fmt.Sprintf("https://pub.dev/api/packages/%s", tool)

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
			return nil, NewBackendError(b.Name(), tool, "package not found on pub.dev", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var data pubResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		var versionStrs []string
		for _, v := range data.Versions {
			versionStrs = append(versionStrs, v.Version)
		}

		// Reverse to get descending order
		for i, j := 0, len(versionStrs)-1; i < j; i, j = i+1, j-1 {
			versionStrs[i], versionStrs[j] = versionStrs[j], versionStrs[i]
		}

		b.versionsCache.Store(tool, versionStrs)
		if b.client.Transport == nil {
			writeEcosystemMetadataDiskCache("pub", tool, versionStrs)
		}
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

func (b *PubBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *PubBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *PubBackend) SupportsChecksum() bool  { return false }
func (b *PubBackend) SupportsGPG() bool       { return false }
func (b *PubBackend) AttestationType() string { return "" }
func (b *PubBackend) IsRecommended() bool     { return true }
func (b *PubBackend) IsScriptless() bool      { return true }
func (b *PubBackend) GetReach() string        { return "Medium" }
func (b *PubBackend) IsStable() bool          { return true }
func (b *PubBackend) SupportsOffline() bool   { return true }
