package backend

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"sort"
	"strings"
	"time"
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

func (b *GoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Go proxy API: https://proxy.golang.org/<module>/@v/list
	// The module name needs to be escaped (lowercase and replace uppercase with !lowercase)
	// For simplicity, we assume the tool name is already a valid module path or we use it as is.
	url := fmt.Sprintf("https://proxy.golang.org/%s/@v/list", tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewBackendError(b.Name(), tool, "module not found on proxy.golang.org", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

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

	// Sort versions (roughly newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	return versions, nil
}

func (b *GoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *GoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
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
