package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DenoBackend struct {
	client *http.Client
}

func NewDenoBackend() *DenoBackend {
	return &DenoBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *DenoBackend) Name() string {
	return "deno"
}

func (b *DenoBackend) Dependencies() []string {
	return nil
}
type denoVersionsResponse struct {
	Latest   string   `json:"latest"`
	Versions []string `json:"versions"`
}

func (b *DenoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://cdn.deno.land/%s/meta/versions.json", tool)

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
		return nil, NewBackendError(b.Name(), tool, "module not found on deno.land", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var data denoVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, v := range data.Versions {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *DenoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		url := fmt.Sprintf("https://cdn.deno.land/%s/meta/versions.json", tool)
		resp, err := b.client.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var data denoVersionsResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
				return &VersionInfo{
					Version:  data.Latest,
					Platform: platform,
				}, nil
			}
		}
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *DenoBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *DenoBackend) SupportsChecksum() bool {
	return true
}

func (b *DenoBackend) SupportsGPG() bool {
	return false
}

func (b *DenoBackend) AttestationType() string {
	return ""
}

func (b *DenoBackend) IsRecommended() bool {
	return true
}

func (b *DenoBackend) IsScriptless() bool {
	return true
}

func (b *DenoBackend) GetReach() string {
	return "Medium"
}

func (b *DenoBackend) IsStable() bool {
	return true
}

func (b *DenoBackend) SupportsOffline() bool {
	return true
}
