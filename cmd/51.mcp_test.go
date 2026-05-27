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
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	
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

	// Test install_tool
	resp = handleMCPTool(ctx, 4, "install_tool", map[string]interface{}{"tool": "node", "version": "20"})
	assert.Nil(t, resp.Error)

	resp = handleMCPTool(ctx, 5, "install_tool", map[string]interface{}{})
	assert.NotNil(t, resp.Error)

	// Test outdated_tools
	resp = handleMCPTool(ctx, 6, "outdated_tools", nil)
	assert.Nil(t, resp.Error)

	// Test unknown
	resp = handleMCPTool(ctx, 7, "unknown_tool", nil)
	assert.NotNil(t, resp.Error)
}
