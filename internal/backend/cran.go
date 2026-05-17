package backend

import (
	"context"
	"net/http"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"time"
)

type CranBackend struct {
	client *http.Client
}

func NewCranBackend() *CranBackend {
	return &CranBackend{
		client: pkgHttp.NewClientWithTimeout(15 * time.Second),
	}
}

func (b *CranBackend) Name() string {
	return "cran"
}

func (b *CranBackend) Dependencies() []string {
	return nil
}
func (b *CranBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// CRAN versions are usually found at https://cran.r-project.org/package=<tool>
	// For simplicity, we return a limited implementation as a placeholder.
	return nil, NewBackendError(b.Name(), tool, "cran version listing is not yet implemented via REST", nil)
}

func (b *CranBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *CranBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *CranBackend) SupportsChecksum() bool {
	return true
}

func (b *CranBackend) SupportsGPG() bool {
	return false
}

func (b *CranBackend) AttestationType() string {
	return ""
}

func (b *CranBackend) IsRecommended() bool {
	return true
}

func (b *CranBackend) IsScriptless() bool {
	return true
}

func (b *CranBackend) GetReach() string {
	return "Medium"
}

func (b *CranBackend) IsStable() bool {
	return true
}

func (b *CranBackend) SupportsOffline() bool {
	return true
}
