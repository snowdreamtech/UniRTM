// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/snowdreamtech/unirtm/internal/backend"
	uplugin "github.com/snowdreamtech/unirtm/internal/plugin"
)

// ExampleBackend implements the backend.Backend interface
type ExampleBackend struct{}

func (b *ExampleBackend) Name() string {
	return "example"
}

func (b *ExampleBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	return []backend.VersionInfo{
		{
			Version:     "1.0.0",
			DownloadURL: "https://example.com/download/1.0.0",
			Checksum:    "abcdef123456",
			Platform:    platform,
		},
	}, nil
}

func (b *ExampleBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	return &backend.VersionInfo{
		Version:     "1.0.0",
		DownloadURL: "https://example.com/download/1.0.0",
		Checksum:    "abcdef123456",
		Platform:    platform,
	}, nil
}

func (b *ExampleBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	return &backend.VersionInfo{
		Version:     version,
		DownloadURL: "https://example.com/download/" + version,
		Checksum:    "abcdef123456",
		Platform:    platform,
	}, nil
}

func (b *ExampleBackend) SupportsChecksum() bool {
	return true
}

func (b *ExampleBackend) SupportsGPG() bool {
	return false
}

func main() {
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"backend": &uplugin.BackendPlugin{Impl: &ExampleBackend{}},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: uplugin.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
