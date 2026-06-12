// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
)

type LuaBackend struct{}

func NewLuaBackend() *LuaBackend {
	return &LuaBackend{}
}

func (b *LuaBackend) Name() string {
	return "lua"
}

func (b *LuaBackend) Dependencies() []string {
	return nil
}

func (b *LuaBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// Lua version resolution is handled internally by LuaProvider.
	return nil, NewBackendError(b.Name(), tool, "lua version listing is not yet implemented via REST", nil)
}

func (b *LuaBackend) ResolveVersion(ctx context.Context, tool, versionRequest string, platform Platform) (*VersionInfo, error) {
	versionRequest = NormalizeVersionPrefix(versionRequest, false)
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *LuaBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform Platform) (*VersionInfo, error) {
	version = NormalizeVersionPrefix(version, false)
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *LuaBackend) SupportsChecksum() bool {
	return false
}

func (b *LuaBackend) SupportsGPG() bool {
	return false
}

func (b *LuaBackend) AttestationType() string {
	return ""
}

func (b *LuaBackend) IsRecommended() bool {
	return true
}

func (b *LuaBackend) IsScriptless() bool {
	return true
}

func (b *LuaBackend) GetReach() string {
	return "Medium"
}

func (b *LuaBackend) IsStable() bool {
	return true
}

func (b *LuaBackend) SupportsOffline() bool {
	return true
}
