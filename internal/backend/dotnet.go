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

// DotnetBackend implements the Backend interface for .NET tools (NuGet).
type DotnetBackend struct {
	client        *http.Client
	requestGroup  singleflight.Group
	versionsCache sync.Map // map[string][]string (tool -> sorted version strings)
}

// NewDotnetBackend creates a new dotnet backend.
func NewDotnetBackend() *DotnetBackend {
	return &DotnetBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *DotnetBackend) Name() string {
	return "dotnet"
}

func (b *DotnetBackend) Dependencies() []string {
	return nil
}

type nugetVersionsResponse struct {
	Versions []string `json:"versions"`
}

func (b *DotnetBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	if ctx == nil {
		return nil, NewBackendError(b.Name(), tool, "create request", fmt.Errorf("net/http: nil Context"))
	}

	pkg := strings.ToLower(tool)

	if val, ok := b.versionsCache.Load(pkg); ok {
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

	result, err, _ := b.requestGroup.Do("dotnet:"+pkg, func() (interface{}, error) {
		if val, ok := b.versionsCache.Load(pkg); ok {
			return val, nil
		}

		url := fmt.Sprintf("https://api.nuget.org/v3-flatcontainer/%s/index.json", pkg)

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
			return nil, NewBackendError(b.Name(), tool, "package not found", nil)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
		}

		var registry nugetVersionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
			return nil, NewBackendError(b.Name(), tool, "decode response", err)
		}

		// NuGet returns versions in ascending order, we want descending
		var versionStrs []string
		for i := len(registry.Versions) - 1; i >= 0; i-- {
			versionStrs = append(versionStrs, registry.Versions[i])
		}

		b.versionsCache.Store(pkg, versionStrs)
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

func (b *DotnetBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *DotnetBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	// Provider handles actual downloading via dotnet cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *DotnetBackend) SupportsChecksum() bool {
	return true
}

func (b *DotnetBackend) SupportsGPG() bool {
	return false
}

func (b *DotnetBackend) AttestationType() string {
	return ""
}

func (b *DotnetBackend) IsRecommended() bool {
	return true
}

func (b *DotnetBackend) IsScriptless() bool {
	return true
}

func (b *DotnetBackend) GetReach() string {
	return "Large"
}

func (b *DotnetBackend) IsStable() bool {
	return true
}

func (b *DotnetBackend) SupportsOffline() bool {
	return true
}
