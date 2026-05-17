package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"sort"
	"time"
)

type ComposerBackend struct {
	client *http.Client
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
	return nil
}
type packagistResponse struct {
	Package struct {
		Versions map[string]interface{} `json:"versions"`
	} `json:"package"`
}

func (b *ComposerBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://packagist.org/packages/%s.json", tool)

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
		return nil, NewBackendError(b.Name(), tool, "package not found on packagist", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var data packagistResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for v := range data.Package.Versions {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	// Sort versions (roughly newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	return versions, nil
}

func (b *ComposerBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *ComposerBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
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
