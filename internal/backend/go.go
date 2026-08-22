// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

type GoBackend struct {
	client *http.Client
}

func NewGoBackend() *GoBackend {
	return &GoBackend{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}
}

func (b *GoBackend) Name() string {
	return "go"
}

func (b *GoBackend) Dependencies() []string {
	return []string{"go"}
}

func getGoProxyBase() string {
	goproxy := env.Get("GOPROXY")
	if goproxy == "" {
		return "https://proxy.golang.org"
	}
	for _, proxy := range strings.Split(goproxy, ",") {
		proxy = strings.TrimSpace(proxy)
		if proxy == "direct" || proxy == "off" || proxy == "" {
			continue
		}
		return strings.TrimSuffix(proxy, "/")
	}
	return "https://proxy.golang.org"
}

func (b *GoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Go proxy API: https://proxy.golang.org/<module>/@v/list
	// If a tool specifies a subpackage (e.g. golang.org/x/vuln/cmd/govulncheck),
	// fetching list on the subpackage returns 404. We iteratively fallback to parent directories
	// to locate the actual Go module root.
	modulePath := tool
	for {
		url := fmt.Sprintf("%s/%s/@v/list", getGoProxyBase(), modulePath)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "create request", err)
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, NewBackendError(b.Name(), tool, "execute request", err)
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var versions []VersionInfo
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				v := strings.TrimSpace(scanner.Text())
				if v != "" {
					versions = append(versions, VersionInfo{
						Version:  v,
						Platform: platform,
					})
				}
			}

			// Sort versions (newest first)
			sort.Slice(versions, func(i, j int) bool {
				return versions[i].Version > versions[j].Version
			})

			return versions, nil
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			lastSlash := strings.LastIndex(modulePath, "/")
			if lastSlash > 0 {
				modulePath = modulePath[:lastSlash]
				continue
			}
			return nil, NewBackendError(b.Name(), tool, "module not found on proxy.golang.org", nil)
		}

		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}
}

func (b *GoBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, true)
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

func (b *GoBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, true)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *GoBackend) SupportsChecksum() bool {
	return true
}

func (b *GoBackend) SupportsGPG() bool {
	return false
}

func (b *GoBackend) AttestationType() string {
	return ""
}

func (b *GoBackend) IsRecommended() bool {
	return true
}

func (b *GoBackend) IsScriptless() bool {
	return true
}

func (b *GoBackend) GetReach() string {
	return "Huge"
}

func (b *GoBackend) IsStable() bool {
	return true
}

func (b *GoBackend) SupportsOffline() bool {
	return true
}
