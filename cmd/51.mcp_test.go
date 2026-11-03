// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
)

func TestHandleMCPTool(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	os.MkdirAll(filepath.Dir(env.GetDatabasePath()), 0755)

	ctx := context.Background()

	// Test list_tools
	resp := handleMCPTool(ctx, 1, "list_tools", nil)
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Nil(t, resp.Error)

	// Test tool_info
	resp = handleMCPTool(ctx, 2, "tool_info", map[string]interface{}{"tool": "go"})
	assert.Nil(t, resp.Error)

	resp = handleMCPTool(ctx, 3, "tool_info", map[string]interface{}{})
	assert.NotNil(t, resp.Error)

	// Test install_tool (use a dummy tool to avoid actually downloading large binaries during tests)
	resp = handleMCPTool(ctx, 4, "install_tool", map[string]interface{}{"tool": "dummy-mcp-tool", "version": "1.0.0"})
	assert.Nil(t, resp.Error)

	resp = handleMCPTool(ctx, 5, "install_tool", map[string]interface{}{})
	assert.NotNil(t, resp.Error)

	// Test outdated_tools
	resp = handleMCPTool(ctx, 6, "outdated_tools", nil)
	assert.Nil(t, resp.Error)

	// Test get_rule
	resp = handleMCPTool(ctx, 7, "get_rule", map[string]interface{}{"rule": "00-index"})
	assert.Nil(t, resp.Error)

	resp = handleMCPTool(ctx, 8, "get_rule", map[string]interface{}{})
	assert.NotNil(t, resp.Error)

	// Test quality_check
	resp = handleMCPTool(ctx, 9, "quality_check", map[string]interface{}{"mode": "staged"})
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)

	// Test unknown
	resp = handleMCPTool(ctx, 10, "unknown_tool", nil)
	assert.NotNil(t, resp.Error)
}

func TestDispatchMCPRequest(t *testing.T) {
	ctx := context.Background()

	// 1. Initialize
	resp := dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	})
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Nil(t, resp.Error)
	resMap, ok := resp.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "2024-11-05", resMap["protocolVersion"])

	// 2. Ping
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "ping",
	})
	assert.Nil(t, resp.Error)

	// 3. Tools list
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/list",
	})
	assert.Nil(t, resp.Error)
	resMap = resp.Result.(map[string]interface{})
	tools := resMap["tools"].([]map[string]interface{})
	assert.NotEmpty(t, tools)

	// 4. Prompts list
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "prompts/list",
	})
	assert.Nil(t, resp.Error)
	resMap = resp.Result.(map[string]interface{})
	prompts := resMap["prompts"].([]map[string]interface{})
	assert.NotEmpty(t, prompts)

	// 5. Prompts get
	paramBytes, _ := json.Marshal(map[string]interface{}{
		"name": "speckit.specify",
		"arguments": map[string]string{
			"input": "test feature",
		},
	})
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "prompts/get",
		Params:  paramBytes,
	})
	assert.Nil(t, resp.Error)

	// 6. Resources list
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "resources/list",
	})
	assert.Nil(t, resp.Error)
	resMap = resp.Result.(map[string]interface{})
	resources := resMap["resources"].([]map[string]interface{})
	assert.NotEmpty(t, resources)

	// 7. Resources read
	paramBytes, _ = json.Marshal(map[string]interface{}{
		"uri": "unirtm://conventions",
	})
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "resources/read",
		Params:  paramBytes,
	})
	assert.Nil(t, resp.Error)

	// 8. Unknown method
	resp = dispatchMCPRequest(ctx, mcpRequest{
		JSONRPC: "2.0",
		ID:      8,
		Method:  "unknown/method",
	})
	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
}
