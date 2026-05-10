// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── self-update ─────────────────────────────────────────────────────────────

func TestSelfUpdateCommandStructure(t *testing.T) {
	assert.Contains(t, selfUpdateCmd.Use, "self-update")
	assert.Contains(t, selfUpdateCmd.Aliases, "upgrade")
	assert.NotNil(t, selfUpdateCmd.RunE)
	assert.NotNil(t, selfUpdateCmd.Flags().Lookup("version"))
}

// ─── implode ──────────────────────────────────────────────────────────────────

func TestImplodeCommandStructure(t *testing.T) {
	assert.Contains(t, implodeCmd.Use, "implode")
	assert.NotNil(t, implodeCmd.RunE)
	assert.NotNil(t, implodeCmd.Flags().Lookup("yes"))
}

// ─── generate ─────────────────────────────────────────────────────────────────

func TestGenerateCommandStructure(t *testing.T) {
	assert.Contains(t, generateCmd.Use, "generate")
	assert.Contains(t, generateCmd.Aliases, "gen")
}

func TestGenerateSubcommands(t *testing.T) {
	assert.NotNil(t, generateGithubActionCmd.RunE)
	assert.NotNil(t, generatePreCommitCmd.RunE)
	assert.NotNil(t, generateShellAliasCmd.RunE)
}

func TestGenerateGithubAction_Output(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = generateGithubActionCmd.RunE(generateGithubActionCmd, []string{})
	})
	assert.Contains(t, out, "unirtm install")
}

func TestGeneratePreCommit_Output(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = generatePreCommitCmd.RunE(generatePreCommitCmd, []string{})
	})
	assert.Contains(t, out, "pre-commit")
}

func TestGenerateShellAlias_Output(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = generateShellAliasCmd.RunE(generateShellAliasCmd, []string{})
	})
	assert.Contains(t, out, "alias u=")
}

// ─── en ───────────────────────────────────────────────────────────────────────

func TestEnCmdStructure(t *testing.T) {
	assert.Contains(t, enCmd.Use, "en")
	assert.NotNil(t, enCmd.RunE)
}

// ─── install-into ─────────────────────────────────────────────────────────────

func TestInstallIntoCmdStructure(t *testing.T) {
	assert.Contains(t, installIntoCmd.Use, "install-into")
	assert.NotNil(t, installIntoCmd.RunE)
	assert.NotNil(t, installIntoCmd.Flags().Lookup("backend"))
}

func TestInstallIntoCmdArgs(t *testing.T) {
	err := installIntoCmd.Args(installIntoCmd, []string{"node@22.14.0", "/tmp/node"})
	assert.NoError(t, err)
	err = installIntoCmd.Args(installIntoCmd, []string{"node@22.14.0"})
	assert.Error(t, err)
}

func TestParseToolVersion(t *testing.T) {
	tool, ver := parseToolVersion("cli/cli@2.70.0")
	assert.Equal(t, "cli/cli", tool)
	assert.Equal(t, "2.70.0", ver)

	tool, ver = parseToolVersion("node")
	assert.Equal(t, "node", tool)
	assert.Equal(t, "", ver)

	tool, ver = parseToolVersion("golang/go@1.22.0")
	assert.Equal(t, "golang/go", tool)
	assert.Equal(t, "1.22.0", ver)
}

// ─── shell-alias ──────────────────────────────────────────────────────────────

func TestShellAliasCmdStructure(t *testing.T) {
	assert.Contains(t, shellAliasCmd.Use, "shell-alias")
	assert.Contains(t, shellAliasCmd.Aliases, "alias")
	assert.NotNil(t, shellAliasCmd.RunE)
}

func TestShellAliasSubcommands(t *testing.T) {
	assert.NotNil(t, shellAliasListCmd.RunE)
	assert.NotNil(t, shellAliasAddCmd.RunE)
	assert.NotNil(t, shellAliasRemoveCmd.RunE)
}

func TestShellAliasAddArgs(t *testing.T) {
	err := shellAliasAddCmd.Args(shellAliasAddCmd, []string{"node", "lts", "22.0.0"})
	assert.NoError(t, err)
	err = shellAliasAddCmd.Args(shellAliasAddCmd, []string{"node", "lts"})
	assert.Error(t, err)
}

// ─── edit ─────────────────────────────────────────────────────────────────────

func TestEditCmdStructure(t *testing.T) {
	assert.Contains(t, editCmd.Use, "edit")
	assert.NotNil(t, editCmd.RunE)
	assert.NotNil(t, editCmd.Flags().Lookup("global"))
}

// ─── token ────────────────────────────────────────────────────────────────────

func TestTokenCmdStructure(t *testing.T) {
	assert.Contains(t, tokenCmd.Use, "token")
	assert.NotNil(t, tokenCmd.RunE)
}

func TestTokenCmdArgs(t *testing.T) {
	err := tokenCmd.Args(tokenCmd, []string{})
	assert.NoError(t, err)
	err = tokenCmd.Args(tokenCmd, []string{"github"})
	assert.NoError(t, err)
	err = tokenCmd.Args(tokenCmd, []string{"a", "b"})
	assert.Error(t, err, "2 args should fail")
}

func TestTokenMasking(t *testing.T) {
	// Verify that knownTokens are defined correctly.
	assert.NotEmpty(t, knownTokens)
	found := false
	for _, t := range knownTokens {
		if t.EnvVar == "GITHUB_TOKEN" {
			found = true
		}
	}
	assert.True(t, found, "GITHUB_TOKEN should be in knownTokens")
}

// ─── mcp ──────────────────────────────────────────────────────────────────────

func TestMCPCmdStructure(t *testing.T) {
	assert.Contains(t, mcpCmd.Use, "mcp")
	assert.NotNil(t, mcpCmd.RunE)
}

func TestMCPHandleTool_ListTools(t *testing.T) {
	// DB may not exist in test env — handleMCPTool may panic if database.Open panics.
	// We just verify the function signature and non-unknown-tool path via install_tool.
	resp := handleMCPTool(context.Background(), "req-1", "install_tool",
		map[string]interface{}{"tool": "node", "version": "22"})
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, "req-1", resp.ID)
}

func TestMCPHandleTool_Unknown(t *testing.T) {
	resp := handleMCPTool(nil, 1, "unknown_tool", nil)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
}

func TestMCPHandleTool_InstallTool_MissingArg(t *testing.T) {
	resp := handleMCPTool(nil, 2, "install_tool", map[string]interface{}{})
	assert.NotNil(t, resp.Error)
}

func TestMCPHandleTool_InstallTool_OK(t *testing.T) {
	resp := handleMCPTool(nil, 3, "install_tool", map[string]interface{}{"tool": "cli/cli", "version": "2.72.0"})
	assert.Nil(t, resp.Error)
}

// ─── ls-remote ─────────────────────────────────────────────────────────────

func TestLsRemoteCmdStructure(t *testing.T) {
	assert.Contains(t, lsRemoteCmd.Use, "ls-remote")
	assert.NotNil(t, lsRemoteCmd.RunE)
}

func TestLsRemoteCmdArgs(t *testing.T) {
	err := lsRemoteCmd.Args(lsRemoteCmd, []string{"node"})
	assert.NoError(t, err)
	err = lsRemoteCmd.Args(lsRemoteCmd, []string{"node@20"})
	assert.NoError(t, err)
}

// ─── prepare ──────────────────────────────────────────────────────────────────

func TestPrepareCmdStructure(t *testing.T) {
	assert.Contains(t, prepareCmd.Use, "prepare")
	assert.Contains(t, prepareCmd.Aliases, "prep")
	assert.NotNil(t, prepareCmd.RunE)
}

// ─── sync ─────────────────────────────────────────────────────────────────────

func TestSyncCmdStructure(t *testing.T) {
	assert.Contains(t, syncCmd.Use, "sync")
	assert.NotNil(t, syncCmd.RunE)
}

// ─── test-tool ───────────────────────────────────────────────────────────────

func TestTestToolCmdStructure(t *testing.T) {
	assert.Contains(t, testToolCmd.Use, "test-tool")
	assert.NotNil(t, testToolCmd.RunE)
}

// ─── tool-stub ───────────────────────────────────────────────────────────────

func TestToolStubCmdStructure(t *testing.T) {
	assert.Contains(t, toolStubCmd.Use, "tool-stub")
	assert.NotNil(t, toolStubCmd.RunE)
}
