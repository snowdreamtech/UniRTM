package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type MavenBackend struct {
	client *http.Client
}

func NewMavenBackend() *MavenBackend {
	return &MavenBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	// tool is expected to be group:artifact
	parts := strings.Split(tool, ":")
	if len(parts) != 2 {
		return nil, NewBackendError(b.Name(), tool, "invalid tool name format, expected group:artifact", nil)
	}

	url := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=g:%s+AND+a:%s&rows=50&core=gav", parts[0], parts[1])

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var versions []VersionInfo
	for _, doc := range data.Response.Docs {
		versions = append(versions, VersionInfo{
			Version:  doc.Version,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *MavenBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
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

func (b *MavenBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
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
