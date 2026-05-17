package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"time"
)

type CabalBackend struct {
	client *http.Client
}

func NewCabalBackend() *CabalBackend {
	return &CabalBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *CabalBackend) Name() string {
	return "cabal"
}

func (b *CabalBackend) Dependencies() []string {
	return nil
}
type hackageResponse []struct {
	Version string `json:"version"`
}

func (b *CabalBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://hackage.haskell.org/package/%s.json", tool)

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
		return nil, NewBackendError(b.Name(), tool, "package not found on hackage", nil)
	}

	var data hackageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, v := range data {
		versions = append(versions, VersionInfo{
			Version:  v.Version,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *CabalBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		return &versions[len(versions)-1], nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CabalBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CabalBackend) SupportsChecksum() bool {
	return true
}

func (b *CabalBackend) SupportsGPG() bool {
	return false
}

func (b *CabalBackend) AttestationType() string {
	return ""
}

func (b *CabalBackend) IsRecommended() bool {
	return true
}

func (b *CabalBackend) IsScriptless() bool {
	return true
}

func (b *CabalBackend) GetReach() string {
	return "Large"
}

func (b *CabalBackend) IsStable() bool {
	return true
}

func (b *CabalBackend) SupportsOffline() bool {
	return true
}
