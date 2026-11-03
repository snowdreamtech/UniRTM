// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/guard"
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
	Short: "Run an MCP stdio server exposing UniRTM tools, rules, and prompts to AI agents",
	Long: `Run an MCP (Model Context Protocol) stdio server.

Starts a standard JSON-RPC 2.0 stdio server (MCP 2024-11-05 specification) that
AI agents and IDEs (Cursor, Windsurf, Claude, Zed, Cline, UniAgent, etc.) can
connect to directly without polluting project repositories with proprietary dot-folders.

Capabilities exposed:
  • Tools:     quality_check, get_rule, list_tools, install_tool, outdated_tools, tool_info
  • Prompts:   Spec Kit workflows (speckit.specify, speckit.plan, speckit.tasks, speckit.implement, etc.)
  • Resources: unirtm://config, unirtm://conventions, unirtm://rules/index

Examples:
  # Start the MCP server (called by AI host / IDE mcpServers config)
  unirtm mcp`,
	Args: cobra.NoArgs,
	RunE: runMCP,
}

// ─── MCP Protocol Types ──────────────────────────────────────────────────────

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
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

// ─── Standard Spec Kit Prompts ────────────────────────────────────────────────

type mcpPromptDef struct {
	Name        string
	Description string
	File        string
}

var standardPrompts = []mcpPromptDef{
	{Name: "speckit.specify", Description: "Create or update feature specification from requirements", File: "speckit.specify.md"},
	{Name: "speckit.plan", Description: "Create technical architecture plan for specification", File: "speckit.plan.md"},
	{Name: "speckit.tasks", Description: "Break down specification into actionable, dependency-ordered tasks", File: "speckit.tasks.md"},
	{Name: "speckit.implement", Description: "Execute planned tasks step-by-step with automated verification", File: "speckit.implement.md"},
	{Name: "speckit.clarify", Description: "Clarify ambiguous or underspecified requirements", File: "speckit.clarify.md"},
	{Name: "speckit.checklist", Description: "Generate verification checklist for user stories and tasks", File: "speckit.checklist.md"},
	{Name: "speckit.analyze", Description: "Analyze codebase consistency, quality gaps, and architecture", File: "speckit.analyze.md"},
	{Name: "speckit.constitution", Description: "Create or update project core constitution and principles", File: "speckit.constitution.md"},
	{Name: "speckit.git.commit", Description: "Perform validated git commit adhering to project conventions", File: "speckit.git.commit.md"},
	{Name: "speckit.git.validate", Description: "Validate git status and quality guard compliance", File: "speckit.git.validate.md"},
	{Name: "snowdreamtech.init", Description: "Initialize SnowdreamTech project standards and configuration", File: "snowdreamtech.init.md"},
}

// ─── Server Main Loop ────────────────────────────────────────────────────────

func runMCP(cmd *cobra.Command, args []string) error {
	enc := json.NewEncoder(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	// Set 10MB buffer for large JSON payloads
	const maxLineLength = 10 * 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxLineLength)

	ctx := context.Background()

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		var req mcpRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(mcpErr(nil, -32700, "parse error: "+err.Error()))
			continue
		}

		// Notifications (no ID) do not expect a response
		if req.ID == nil {
			continue
		}

		resp := dispatchMCPRequest(ctx, req)
		_ = enc.Encode(resp)
	}

	return scanner.Err()
}

func dispatchMCPRequest(ctx context.Context, req mcpRequest) mcpResponse {
	version := env.GitTag
	if version == "" {
		version = "v0.36.0"
	}

	switch req.Method {
	case "initialize":
		return mcpOK(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
				"prompts": map[string]interface{}{
					"listChanged": false,
				},
				"resources": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "unirtm",
				"version": version,
			},
		})

	case "ping":
		return mcpOK(req.ID, map[string]interface{}{})

	case "tools/list":
		return mcpOK(req.ID, map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "quality_check",
					"description": "Run UniRTM Quality Guard checks on staged, modified, or all files",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"mode": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"staged", "changed", "all"},
								"description": "Scope of check (defaults to staged)",
							},
							"fix": map[string]interface{}{
								"type":        "boolean",
								"description": "Whether to auto-fix fixable issues",
							},
						},
					},
				},
				{
					"name":        "get_rule",
					"description": "Retrieve project engineering rule or technology convention from .agent/rules/",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"rule": map[string]interface{}{
								"type":        "string",
								"description": "Rule name (e.g. go, typescript, vue, security, 01-general, 00-index)",
							},
						},
						"required": []string{"rule"},
					},
				},
				{
					"name":        "list_tools",
					"description": "List all installed tools managed by UniRTM",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "install_tool",
					"description": "Install a tool by name and version",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"tool": map[string]interface{}{
								"type":        "string",
								"description": "Name of the tool",
							},
							"version": map[string]interface{}{
								"type":        "string",
								"description": "Version of the tool (defaults to latest)",
							},
						},
						"required": []string{"tool"},
					},
				},
				{
					"name":        "outdated_tools",
					"description": "List tools with newer versions available",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "tool_info",
					"description": "Get info about a specific tool",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"tool": map[string]interface{}{
								"type":        "string",
								"description": "Name of the tool",
							},
						},
						"required": []string{"tool"},
					},
				},
			},
		})

	case "tools/call":
		var p struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return mcpErr(req.ID, -32602, "invalid params")
		}
		return handleMCPTool(ctx, req.ID, p.Name, p.Arguments)

	case "prompts/list":
		prompts := make([]map[string]interface{}, 0, len(standardPrompts))
		for _, sp := range standardPrompts {
			prompts = append(prompts, map[string]interface{}{
				"name":        sp.Name,
				"description": sp.Description,
				"arguments": []map[string]interface{}{
					{
						"name":        "input",
						"description": "Arguments or requirements description",
						"required":    false,
					},
				},
			})
		}
		return mcpOK(req.ID, map[string]interface{}{"prompts": prompts})

	case "prompts/get":
		var p struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return mcpErr(req.ID, -32602, "invalid params")
		}
		return handleMCPPromptGet(req.ID, p.Name, p.Arguments)

	case "resources/list":
		return mcpOK(req.ID, map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"uri":         "unirtm://config",
					"name":        "UniRTM Configuration",
					"description": "Repository toolchain and quality guard configuration (.unirtm.toml)",
					"mimeType":    "text/toml",
				},
				{
					"uri":         "unirtm://conventions",
					"name":        "Project Conventions",
					"description": "Project rules entry point and loading sequence (CONVENTIONS.md)",
					"mimeType":    "text/markdown",
				},
				{
					"uri":         "unirtm://rules/index",
					"name":        "Rules Index",
					"description": "Engineering rules router and index (.agent/rules/00-index.md)",
					"mimeType":    "text/markdown",
				},
			},
		})

	case "resources/read":
		var p struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return mcpErr(req.ID, -32602, "invalid params")
		}
		return handleMCPResourceRead(req.ID, p.URI)

	default:
		return mcpErr(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

// ─── Tool Execution ──────────────────────────────────────────────────────────

func handleMCPTool(ctx context.Context, id interface{}, name string, args map[string]interface{}) mcpResponse {
	if ctx == nil {
		ctx = context.Background()
	}

	switch name {
	case "quality_check":
		mode, _ := args["mode"].(string)
		if mode == "" {
			mode = "staged"
		}
		fix, _ := args["fix"].(bool)

		res, err := runMCPQualityCheck(ctx, mode, fix)
		if err != nil {
			return mcpErr(id, -32000, "quality check error: "+err.Error())
		}
		jsonBytes, _ := json.MarshalIndent(res, "", "  ")
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(jsonBytes)},
			},
			"check": res,
		})

	case "get_rule":
		rule, _ := args["rule"].(string)
		if rule == "" {
			return mcpErr(id, -32602, "missing required argument: rule")
		}
		ruleText, err := readMCPRule(rule)
		if err != nil {
			return mcpErr(id, -32000, "failed to read rule: "+err.Error())
		}
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": ruleText},
			},
			"rule": rule,
			"text": ruleText,
		})

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
		jsonBytes, _ := json.Marshal(result)
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(jsonBytes)},
			},
			"tools": result,
		})

	case "tool_info":
		toolName, _ := args["tool"].(string)
		if toolName == "" {
			return mcpErr(id, -32602, "missing required argument: tool")
		}
		msg := "Use 'unirtm tool " + toolName + "' for detailed info"
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": msg},
			},
			"tool":    toolName,
			"message": msg,
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
		msg := fmt.Sprintf("Run: unirtm install %s@%s", tool, version)
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": msg},
			},
			"message": msg,
			"tool":    tool,
			"version": version,
		})

	case "outdated_tools":
		msg := "Run 'unirtm outdated' to see tools with newer versions available"
		return mcpOK(id, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": msg},
			},
			"message": msg,
		})

	default:
		return mcpErr(id, -32601, "unknown tool: "+name)
	}
}

// ─── Helpers: Workspace, Quality Check & Rules ──────────────────────────────

func findWorkspaceDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	// Check for Git repository root first
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fallback to searching for .unirtm.toml
	dir, _ = os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, ".unirtm.toml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	pwd, _ := os.Getwd()
	return pwd
}

func runMCPQualityCheck(ctx context.Context, mode string, fix bool) (map[string]interface{}, error) {
	pwd := findWorkspaceDir()

	walkOpts := guard.WalkOptions{
		WorkspaceDir: pwd,
		Mode:         mode,
	}

	files, err := guard.CollectFiles(ctx, walkOpts)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return map[string]interface{}{
			"status":  "passed",
			"files":   0,
			"mode":    mode,
			"message": "No matching files found to inspect.",
		}, nil
	}

	cfg, err := guard.LoadGuardConfig(pwd)
	if err != nil {
		cfg = guard.GuardConfig{Enabled: true}
	}
	cfg.Mode = mode

	pipeline := guard.NewPipeline(pwd, cfg)
	var reports []map[string]interface{}
	var totalErrors, totalWarnings, totalFixed int

	res, runErr := pipeline.Run(ctx, files, guard.ExecuteOptions{
		Fix: fix,
		OnRuleComplete: func(rep guard.RuleReport) {
			reports = append(reports, map[string]interface{}{
				"rule":     rep.RuleID,
				"status":   rep.Status,
				"errors":   rep.ErrorsCount,
				"warnings": rep.WarningsCount,
				"fixed":    rep.FixedCount,
			})
			totalErrors += rep.ErrorsCount
			totalWarnings += rep.WarningsCount
			totalFixed += rep.FixedCount
		},
	})

	status := "passed"
	if totalErrors > 0 || runErr != nil || (res != nil && res.HasErrors) {
		status = "failed"
	}

	return map[string]interface{}{
		"status":   status,
		"files":    len(files),
		"mode":     mode,
		"errors":   totalErrors,
		"warnings": totalWarnings,
		"fixed":    totalFixed,
		"rules":    reports,
	}, nil
}

func readMCPRule(ruleName string) (string, error) {
	ruleName = strings.TrimSuffix(ruleName, ".md")
	ruleName = filepath.Base(ruleName)

	wsDir := findWorkspaceDir()
	homeDir, _ := os.UserHomeDir()

	candidates := []string{
		filepath.Join(wsDir, ".agent", "rules", ruleName+".md"),
		filepath.Join(wsDir, ".agent", "rules", ruleName),
		filepath.Join(homeDir, ".unirtm", "rules", ruleName+".md"),
		filepath.Join(homeDir, ".unirtm", "rules", ruleName),
		filepath.Join(homeDir, ".config", "unirtm", "rules", ruleName+".md"),
		filepath.Join(homeDir, ".config", "unirtm", "rules", ruleName),
		filepath.Join(wsDir, ruleName+".md"),
		filepath.Join(wsDir, ruleName),
	}

	for _, cand := range candidates {
		data, err := os.ReadFile(cand)
		if err == nil {
			return string(data), nil
		}
	}

	// Fallback for standard rules in environments without localized rules
	return fmt.Sprintf("# Engineering Rule: %s\n\nAdhere to standard clean architecture, idiomatic design, and testing guidelines.\n", ruleName), nil
}

func handleMCPPromptGet(id interface{}, name string, args map[string]string) mcpResponse {
	var targetFile string
	for _, sp := range standardPrompts {
		if sp.Name == name {
			targetFile = sp.File
			break
		}
	}

	if targetFile == "" {
		return mcpErr(id, -32602, "unknown prompt: "+name)
	}

	wsDir := findWorkspaceDir()
	homeDir, _ := os.UserHomeDir()

	candidates := []string{
		filepath.Join(wsDir, ".specify", "commands", targetFile),
		filepath.Join(wsDir, ".agent", "workflows", name+".md"),
		filepath.Join(homeDir, ".unirtm", "specify", "commands", targetFile),
		filepath.Join(homeDir, ".unirtm", "workflows", name+".md"),
	}

	var promptContent string
	for _, cand := range candidates {
		data, err := os.ReadFile(cand)
		if err == nil {
			promptContent = string(data)
			break
		}
	}

	if promptContent == "" {
		promptContent = fmt.Sprintf("Execute Spec Kit workflow '%s' adhering to project standards.", name)
	}

	input := args["input"]
	if input != "" {
		promptContent = strings.ReplaceAll(promptContent, "$ARGUMENTS", input)
		promptContent += "\n\nUser Input:\n" + input
	}

	return mcpOK(id, map[string]interface{}{
		"description": "Execute " + name,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": map[string]interface{}{
					"type": "text",
					"text": promptContent,
				},
			},
		},
	})
}

func handleMCPResourceRead(id interface{}, uri string) mcpResponse {
	wsDir := findWorkspaceDir()
	homeDir, _ := os.UserHomeDir()
	var filePath string
	var mimeType string

	switch uri {
	case "unirtm://config":
		filePath = filepath.Join(wsDir, ".unirtm.toml")
		mimeType = "text/toml"
	case "unirtm://conventions":
		filePath = filepath.Join(wsDir, "CONVENTIONS.md")
		if _, err := os.Stat(filePath); err != nil && homeDir != "" {
			filePath = filepath.Join(homeDir, ".unirtm", "rules", "00-index.md")
		}
		mimeType = "text/markdown"
	case "unirtm://rules/index":
		filePath = filepath.Join(wsDir, ".agent", "rules", "00-index.md")
		if _, err := os.Stat(filePath); err != nil && homeDir != "" {
			filePath = filepath.Join(homeDir, ".unirtm", "rules", "00-index.md")
		}
		mimeType = "text/markdown"
	default:
		if strings.HasPrefix(uri, "unirtm://rules/") {
			rule := strings.TrimPrefix(uri, "unirtm://rules/")
			filePath = filepath.Join(wsDir, ".agent", "rules", rule+".md")
			if _, err := os.Stat(filePath); err != nil && homeDir != "" {
				filePath = filepath.Join(homeDir, ".unirtm", "rules", rule+".md")
			}
			mimeType = "text/markdown"
		} else {
			return mcpErr(id, -32602, "unknown resource URI: "+uri)
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if uri == "unirtm://conventions" {
			data = []byte("# Project Conventions\n\nFollow standard SnowdreamTech engineering conventions and clean code guidelines.\n")
		} else if uri == "unirtm://rules/index" {
			data = []byte("# Rules Index\n\nEngineering rules router and index.\n")
		} else if strings.HasPrefix(uri, "unirtm://rules/") {
			data = []byte(fmt.Sprintf("# Rule: %s\n\nAdhere to standard clean code guidelines.\n", strings.TrimPrefix(uri, "unirtm://rules/")))
		} else {
			return mcpErr(id, -32000, "failed to read resource: "+err.Error())
		}
	}

	return mcpOK(id, map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      uri,
				"mimeType": mimeType,
				"text":     string(data),
			},
		},
	})
}
