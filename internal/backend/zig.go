package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

type ZigBackend struct {
	client *http.Client
}

func NewZigBackend() *ZigBackend {
	return &ZigBackend{
		client: pkgHttp.NewClientWithTimeout(10 * time.Second),
	}
}

func (b *ZigBackend) Name() string {
	return "zig"
}

func (b *ZigBackend) Dependencies() []string {
	return nil
}

type zigDownloadResponse map[string]interface{}

func (b *ZigBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Zig compiler versions are listed at https://ziglang.org/download/index.json
	// For zig packages, it's often github releases.
	// For now we implement the compiler/core discovery.
	url := "https://ziglang.org/download/index.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	var data zigDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for v := range data {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *ZigBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		return &VersionInfo{
			Version:  "master", // Zig calls latest 'master' or we pick from index
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *ZigBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *ZigBackend) SupportsChecksum() bool {
	return true
}

func (b *ZigBackend) SupportsGPG() bool {
	return false
}

func (b *ZigBackend) AttestationType() string {
	return ""
}

func (b *ZigBackend) IsRecommended() bool {
	return true
}

func (b *ZigBackend) IsScriptless() bool {
	return true
}

func (b *ZigBackend) GetReach() string {
	return "Medium"
}

func (b *ZigBackend) IsStable() bool {
	return true
}

func (b *ZigBackend) SupportsOffline() bool {
	return true
}
