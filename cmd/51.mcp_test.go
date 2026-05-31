// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
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

	// Test install_tool (use a dummy tool to avoid actually downloading large binaries like Node.js during tests)
	resp = handleMCPTool(ctx, 4, "install_tool", map[string]interface{}{"tool": "dummy-mcp-tool", "version": "1.0.0"})
	assert.NotNil(t, resp.Error) // Will fail because dummy-mcp-tool doesn't exist, but tests arg parsing

	resp = handleMCPTool(ctx, 5, "install_tool", map[string]interface{}{})
	assert.NotNil(t, resp.Error)

	// Test outdated_tools
	resp = handleMCPTool(ctx, 6, "outdated_tools", nil)
	assert.Nil(t, resp.Error)

	// Test unknown
	resp = handleMCPTool(ctx, 7, "unknown_tool", nil)
	assert.NotNil(t, resp.Error)
}
