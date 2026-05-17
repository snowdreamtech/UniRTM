package backend

import (
	"context"
	"net/http"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

type LuarocksBackend struct {
	client *http.Client
}

func NewLuarocksBackend() *LuarocksBackend {
	return &LuarocksBackend{
		client: pkgHttp.NewClientWithTimeout(15 * time.Second),
	}
}

func (b *LuarocksBackend) Name() string {
	return "luarocks"
}

func (b *LuarocksBackend) Dependencies() []string {
	return nil
}
func (b *LuarocksBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Luarocks doesn't have a very clean JSON API for just listing versions of one rock easily without parsing a huge manifest.
	// But we can check https://luarocks.org/modules/<user>/<tool>
	// For now, we return limited support as a placeholder with a meaningful note.
	return nil, NewBackendError(b.Name(), tool, "luarocks version listing is not yet implemented via REST", nil)
}

func (b *LuarocksBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *LuarocksBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *LuarocksBackend) SupportsChecksum() bool {
	return true
}

func (b *LuarocksBackend) SupportsGPG() bool {
	return false
}

func (b *LuarocksBackend) AttestationType() string {
	return ""
}

func (b *LuarocksBackend) IsRecommended() bool {
	return true
}

func (b *LuarocksBackend) IsScriptless() bool {
	return true
}

func (b *LuarocksBackend) GetReach() string {
	return "Medium"
}

func (b *LuarocksBackend) IsStable() bool {
	return true
}

func (b *LuarocksBackend) SupportsOffline() bool {
	return true
}
