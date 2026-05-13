package backend

import (
	"context"
	"net/http"
	"time"
)

type GoBackend struct {
	client *http.Client
}

func NewGoBackend() *GoBackend {
	return &GoBackend{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *GoBackend) Name() string {
	return "go"
}

func (b *GoBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// In the future, we can integrate with proxy.golang.org to list versions
	return nil, nil
}

func (b *GoBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	// For now, we return the version as is. 
	// In the future, we can integrate with proxy.golang.org to resolve "latest"
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
