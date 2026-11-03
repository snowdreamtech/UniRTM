// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package plugin

import (
	"net"
	"net/rpc"
	"testing"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendPlugin_ServerClient(t *testing.T) {
	mb := &mockBackend{}
	bp := &BackendPlugin{Impl: mb}

	// Test Server()
	srv, err := bp.Server(nil)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	_, ok := srv.(*BackendRPCServer)
	assert.True(t, ok, "expected *BackendRPCServer")

	// Test Client()
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	rpcClient := rpc.NewClient(c1)
	defer rpcClient.Close()

	cli, err := BackendPlugin{}.Client(&goplugin.MuxBroker{}, rpcClient)
	require.NoError(t, err)
	assert.NotNil(t, cli)
	_, ok = cli.(*BackendRPCClient)
	assert.True(t, ok, "expected *BackendRPCClient")
}

func TestProviderPlugin_ServerClient(t *testing.T) {
	mp := &mockProvider{}
	pp := &ProviderPlugin{Impl: mp}

	// Test Server()
	srv, err := pp.Server(nil)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	_, ok := srv.(*ProviderRPCServer)
	assert.True(t, ok, "expected *ProviderRPCServer")

	// Test Client()
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	rpcClient := rpc.NewClient(c1)
	defer rpcClient.Close()

	cli, err := ProviderPlugin{}.Client(&goplugin.MuxBroker{}, rpcClient)
	require.NoError(t, err)
	assert.NotNil(t, cli)
	_, ok = cli.(*ProviderRPCClient)
	assert.True(t, ok, "expected *ProviderRPCClient")
}
