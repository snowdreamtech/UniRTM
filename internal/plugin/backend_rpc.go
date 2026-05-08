// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package plugin

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/snowdreamtech/unirtm/internal/backend"
)

// ListVersionsArgs represents arguments for ListVersions.
type ListVersionsArgs struct {
	Tool     string
	Platform backend.Platform
}

// ResolveVersionArgs represents arguments for ResolveVersion.
type ResolveVersionArgs struct {
	Tool           string
	VersionRequest string
	Platform       backend.Platform
}

// GetDownloadInfoArgs represents arguments for GetDownloadInfo.
type GetDownloadInfoArgs struct {
	Tool     string
	Version  string
	Platform backend.Platform
}

// BackendRPCClient is an implementation of backend.Backend that talks over RPC.
type BackendRPCClient struct {
	client *rpc.Client
}

func (m *BackendRPCClient) Name() string {
	var resp string
	err := m.client.Call("Plugin.Name", new(interface{}), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (m *BackendRPCClient) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	var resp []backend.VersionInfo
	args := ListVersionsArgs{
		Tool:     tool,
		Platform: platform,
	}
	err := m.client.Call("Plugin.ListVersions", args, &resp)
	return resp, err
}

func (m *BackendRPCClient) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	var resp backend.VersionInfo
	args := ResolveVersionArgs{
		Tool:           tool,
		VersionRequest: versionRequest,
		Platform:       platform,
	}
	err := m.client.Call("Plugin.ResolveVersion", args, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *BackendRPCClient) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	var resp backend.VersionInfo
	args := GetDownloadInfoArgs{
		Tool:     tool,
		Version:  version,
		Platform: platform,
	}
	err := m.client.Call("Plugin.GetDownloadInfo", args, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *BackendRPCClient) SupportsChecksum() bool {
	var resp bool
	err := m.client.Call("Plugin.SupportsChecksum", new(interface{}), &resp)
	if err != nil {
		return false
	}
	return resp
}

func (m *BackendRPCClient) SupportsGPG() bool {
	var resp bool
	err := m.client.Call("Plugin.SupportsGPG", new(interface{}), &resp)
	if err != nil {
		return false
	}
	return resp
}

// BackendRPCServer is the RPC server that BackendRPCClient talks to, conforming to
// the requirements of net/rpc.
type BackendRPCServer struct {
	// This is the real implementation
	Impl backend.Backend
}

func (s *BackendRPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *BackendRPCServer) ListVersions(args ListVersionsArgs, resp *[]backend.VersionInfo) error {
	res, err := s.Impl.ListVersions(context.Background(), args.Tool, args.Platform)
	*resp = res
	return err
}

func (s *BackendRPCServer) ResolveVersion(args ResolveVersionArgs, resp *backend.VersionInfo) error {
	res, err := s.Impl.ResolveVersion(context.Background(), args.Tool, args.VersionRequest, args.Platform)
	if err == nil && res != nil {
		*resp = *res
	}
	return err
}

func (s *BackendRPCServer) GetDownloadInfo(args GetDownloadInfoArgs, resp *backend.VersionInfo) error {
	res, err := s.Impl.GetDownloadInfo(context.Background(), args.Tool, args.Version, args.Platform)
	if err == nil && res != nil {
		*resp = *res
	}
	return err
}

func (s *BackendRPCServer) SupportsChecksum(args interface{}, resp *bool) error {
	*resp = s.Impl.SupportsChecksum()
	return nil
}

func (s *BackendRPCServer) SupportsGPG(args interface{}, resp *bool) error {
	*resp = s.Impl.SupportsGPG()
	return nil
}

// BackendPlugin is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a BackendRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return BackendRPCClient for this.
type BackendPlugin struct {
	Impl backend.Backend
}

func (p *BackendPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &BackendRPCServer{Impl: p.Impl}, nil
}

func (BackendPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &BackendRPCClient{client: c}, nil
}
