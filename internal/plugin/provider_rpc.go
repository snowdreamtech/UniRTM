// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package plugin

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

// ProviderRPCClient is an implementation of provider.Provider that talks over RPC.
type ProviderRPCClient struct {
	client *rpc.Client
}

func (m *ProviderRPCClient) Name() string {
	var resp string
	err := m.client.Call("Plugin.Name", new(interface{}), &resp)
	if err != nil {
		return ""
	}
	return resp
}

type InstallArgs struct {
	InstallPath  string
	ArtifactPath string
	Version      string
}

func (m *ProviderRPCClient) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	var resp struct{}
	args := InstallArgs{
		InstallPath:  installPath,
		ArtifactPath: artifactPath,
		Version:      version,
	}
	return m.client.Call("Plugin.Install", args, &resp)
}

type PostInstallArgs struct {
	InstallPath string
	Version     string
}

func (m *ProviderRPCClient) PostInstall(ctx context.Context, installPath string, version string) error {
	var resp struct{}
	args := PostInstallArgs{
		InstallPath: installPath,
		Version:     version,
	}
	return m.client.Call("Plugin.PostInstall", args, &resp)
}

type GenerateShimsArgs struct {
	InstallPath string
	Version     string
}

func (m *ProviderRPCClient) GenerateShims(installPath string, version string) (map[string]string, error) {
	var resp map[string]string
	args := GenerateShimsArgs{
		InstallPath: installPath,
		Version:     version,
	}
	err := m.client.Call("Plugin.GenerateShims", args, &resp)
	return resp, err
}

type DetectVersionArgs struct {
	InstallPath string
}

func (m *ProviderRPCClient) DetectVersion(ctx context.Context, installPath string) (string, error) {
	var resp string
	args := DetectVersionArgs{
		InstallPath: installPath,
	}
	err := m.client.Call("Plugin.DetectVersion", args, &resp)
	return resp, err
}

type ListExecutablesArgs struct {
	InstallPath string
	Version     string
}

func (m *ProviderRPCClient) ListExecutables(installPath string, version string) ([]string, error) {
	var resp []string
	args := ListExecutablesArgs{
		InstallPath: installPath,
		Version:     version,
	}
	err := m.client.Call("Plugin.ListExecutables", args, &resp)
	return resp, err
}

type UninstallArgs struct {
	InstallPath string
	Version     string
}

func (m *ProviderRPCClient) Uninstall(ctx context.Context, installPath string, version string) error {
	var resp struct{}
	args := UninstallArgs{
		InstallPath: installPath,
		Version:     version,
	}
	return m.client.Call("Plugin.Uninstall", args, &resp)
}

// ProviderRPCServer is the RPC server that ProviderRPCClient talks to, conforming to
// the requirements of net/rpc.
type ProviderRPCServer struct {
	Impl provider.Provider
}

func (s *ProviderRPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *ProviderRPCServer) Install(args InstallArgs, resp *struct{}) error {
	return s.Impl.Install(context.Background(), args.InstallPath, args.ArtifactPath, args.Version)
}

func (s *ProviderRPCServer) PostInstall(args PostInstallArgs, resp *struct{}) error {
	return s.Impl.PostInstall(context.Background(), args.InstallPath, args.Version)
}

func (s *ProviderRPCServer) GenerateShims(args GenerateShimsArgs, resp *map[string]string) error {
	res, err := s.Impl.GenerateShims(args.InstallPath, args.Version)
	*resp = res
	return err
}

func (s *ProviderRPCServer) DetectVersion(args DetectVersionArgs, resp *string) error {
	res, err := s.Impl.DetectVersion(context.Background(), args.InstallPath)
	*resp = res
	return err
}

func (s *ProviderRPCServer) ListExecutables(args ListExecutablesArgs, resp *[]string) error {
	res, err := s.Impl.ListExecutables(args.InstallPath, args.Version)
	*resp = res
	return err
}

func (s *ProviderRPCServer) Uninstall(args UninstallArgs, resp *struct{}) error {
	return s.Impl.Uninstall(context.Background(), args.InstallPath, args.Version)
}

// ProviderPlugin is the implementation of plugin.Plugin so we can serve/consume this
type ProviderPlugin struct {
	Impl provider.Provider
}

func (p *ProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProviderRPCServer{Impl: p.Impl}, nil
}

func (ProviderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProviderRPCClient{client: c}, nil
}
