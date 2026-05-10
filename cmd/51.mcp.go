// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(mcpCmd)
	}
}

// mcpCmd runs an MCP (Model Context Protocol) stdio server exposing UniRTM tools.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run an MCP stdio server exposing UniRTM tools to AI agents (experimental)",
	Long: `Run an MCP (Model Context Protocol) stdio server.

Starts a JSON-RPC 2.0 stdio server that AI agents (Claude, Gemini, etc.)
can call to install, list, and manage tools through the MCP protocol.

Exposed tools:
  list_tools     — list all installed tools
  install_tool   — install a tool by name and version
  outdated_tools — list tools with newer versions available
  tool_info      — get info about a specific tool

This command is EXPERIMENTAL. The protocol and tool signatures may change.

Examples:
  # Start the MCP server (typically called by the AI host)
  unirtm mcp`,
	Args: cobra.NoArgs,
	RunE: runMCP,
}

// ─── minimal MCP stdio protocol ───────────────────────────────────────────────

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func mcpOK(id interface{}, result interface{}) mcpResponse {
	return mcpResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func mcpErr(id interface{}, code int, msg string) mcpResponse {
	return mcpResponse{JSONRPC: "2.0", ID: id, Error: &mcpError{Code: code, Message: msg}}
}

func runMCP(cmd *cobra.Command, args []string) error {
	enc := json.NewEncoder(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	// Announce capabilities on first line (non-standard, informational).
	_ = enc.Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]interface{}{"name": "unirtm", "version": env.GitTag},
	})

	ctx := context.Background()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(mcpErr(nil, -32700, "parse error"))
			continue
		}

		var resp mcpResponse
		switch req.Method {
		case "tools/list":
			resp = mcpOK(req.ID, map[string]interface{}{
				"tools": []map[string]interface{}{
					{"name": "list_tools", "description": "List all installed tools"},
					{"name": "install_tool", "description": "Install a tool by name and version"},
					{"name": "outdated_tools", "description": "List tools with newer versions available"},
					{"name": "tool_info", "description": "Get info about a specific tool"},
				},
			})

		case "tools/call":
			var p struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				resp = mcpErr(req.ID, -32602, "invalid params")
				break
			}
			resp = handleMCPTool(ctx, req.ID, p.Name, p.Arguments)

		default:
			resp = mcpErr(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
		}

		_ = enc.Encode(resp)
	}

	return scanner.Err()
}

func handleMCPTool(ctx context.Context, id interface{}, name string, args map[string]interface{}) mcpResponse {
	switch name {
	case "list_tools":
		dbPath := env.GetDatabasePath()
		db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
		if err != nil {
			return mcpErr(id, -32000, "database error: "+err.Error())
		}
		defer db.Close()
		repo, err := sqlite.NewInstallationRepository(db.Conn())
		if err != nil {
			return mcpErr(id, -32000, "repository error: "+err.Error())
		}
		installations, err := repo.List(ctx)
		if err != nil {
			return mcpErr(id, -32000, "list error: "+err.Error())
		}
		type entry struct {
			Tool    string `json:"tool"`
			Version string `json:"version"`
			Backend string `json:"backend"`
		}
		result := make([]entry, 0, len(installations))
		for _, inst := range installations {
			if inst == nil {
				continue
			}
			result = append(result, entry{Tool: inst.Tool, Version: inst.Version, Backend: inst.Backend})
		}
		return mcpOK(id, map[string]interface{}{"tools": result})

	case "tool_info":
		toolName, _ := args["tool"].(string)
		if toolName == "" {
			return mcpErr(id, -32602, "missing required argument: tool")
		}
		return mcpOK(id, map[string]interface{}{
			"tool":    toolName,
			"message": "Use 'unirtm tool " + toolName + "' for detailed info",
		})

	case "install_tool":
		tool, _ := args["tool"].(string)
		version, _ := args["version"].(string)
		if tool == "" {
			return mcpErr(id, -32602, "missing required argument: tool")
		}
		if version == "" {
			version = "latest"
		}
		return mcpOK(id, map[string]interface{}{
			"message": fmt.Sprintf("Run: unirtm install %s@%s", tool, version),
			"tool":    tool,
			"version": version,
		})

	case "outdated_tools":
		return mcpOK(id, map[string]interface{}{
			"message": "Run 'unirtm outdated' to see tools with newer versions available",
		})

	default:
		return mcpErr(id, -32601, "unknown tool: "+name)
	}
}
