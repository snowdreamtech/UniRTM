// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNpmProvider_Install_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash-based mock test on windows")
	}
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	// Mock npm executable
	nodeInstallsDir := filepath.Join(tmpDir, "installs", "node", "18.0.0", "bin")
	err := os.MkdirAll(nodeInstallsDir, 0755)
	require.NoError(t, err)

	npmScript := filepath.Join(nodeInstallsDir, "npm")
	mockNpm := `#!/bin/sh
# Mock npm install
exit 0
`
	err = os.WriteFile(npmScript, []byte(mockNpm), 0755)
	require.NoError(t, err)

	p := provider.NewNpmProvider()
	installPath := filepath.Join(tmpDir, "npm_install", "test_pkg")

	err = p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.NoError(t, err)
}

func TestNpmProvider_Install_NpmNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	// Clear PATH
	t.Setenv("PATH", "")

	p := provider.NewNpmProvider()
	installPath := filepath.Join(tmpDir, "npm_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "npm is required")
}

func TestNpmProvider_ListExecutables(t *testing.T) {
	p := provider.NewNpmProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(binDir, "dummy1"), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, "dummy2"), []byte(""), 0644) // not executable

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	require.NoError(t, err)
	assert.Len(t, exes, 1)
	assert.Contains(t, exes, filepath.Join(binDir, "dummy1"))
}

// TestNpmProvider_RewriteCmdNodePath_npm7Format tests that the npm 7+ IF EXIST
// conditional block is rewritten to use the absolute node.exe path.
func TestNpmProvider_RewriteCmdNodePath_npm7Format(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only .cmd rewrite test")
	}
	p := provider.NewNpmProvider()
	tmpDir := t.TempDir()

	// npm 7+ generated .cmd content (CRLF line endings like Windows)
	cmdContent := "@ECHO off\r\nGOTO start\r\n:find_dp0\r\nSET dp0=%~dp0\r\nEXIT /b\r\n:start\r\nSETLOCAL\r\nCALL :find_dp0\r\nIF EXIST \"%dp0%\\node.exe\" (\r\n  SET \"_prog=%dp0%\\node.exe\"\r\n) ELSE (\r\n  SET \"_prog=node\"\r\n  SET PATHEXT=%PATHEXT:;.JS;=;%\r\n)\r\nendLocal & goto #_undefined_# 2>NUL || title %COMSPEC% & \"%_prog%\"  \"%dp0%\\..\\lib\\node_modules\\prettier\\bin\\prettier.cjs\" %*\r\n"
	cmdFile := filepath.Join(tmpDir, "prettier.cmd")
	require.NoError(t, os.WriteFile(cmdFile, []byte(cmdContent), 0644))

	// Create a fake node.exe so fixWindowsCmdWrappers can find it
	nodeDir := filepath.Join(tmpDir, "installs", "node", "26.1.0")
	require.NoError(t, os.MkdirAll(nodeDir, 0755))
	nodePath := filepath.Join(nodeDir, "node.exe")
	require.NoError(t, os.WriteFile(nodePath, []byte(""), 0755))

	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	require.NoError(t, p.PostInstall(context.Background(), "prettier", tmpDir, "3.8.3"))

	result, err := os.ReadFile(cmdFile)
	require.NoError(t, err)
	content := string(result)

	// The IF EXIST block should be replaced with a direct SET
	assert.Contains(t, content, `SET "_prog=`+nodePath+`"`)
	assert.NotContains(t, content, `IF EXIST "%dp0%\node.exe"`)
	assert.NotContains(t, content, `SET "_prog=node"`)
	// The script path portion must remain intact
	assert.Contains(t, content, "prettier.cjs")
}

// TestNpmProvider_RewriteCmdNodePath_LegacyFormat tests the older npm one-liner
// format ("%~dp0\node.exe") is rewritten correctly.
func TestNpmProvider_RewriteCmdNodePath_LegacyFormat(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only .cmd rewrite test")
	}
	p := provider.NewNpmProvider()
	tmpDir := t.TempDir()

	cmdContent := "@ECHO OFF\r\n\"%~dp0\\node.exe\"  \"%~dp0\\..\\node_modules\\prettier\\bin\\prettier.js\" %*\r\n"
	cmdFile := filepath.Join(tmpDir, "prettier.cmd")
	require.NoError(t, os.WriteFile(cmdFile, []byte(cmdContent), 0644))

	nodeDir := filepath.Join(tmpDir, "installs", "node", "26.1.0")
	require.NoError(t, os.MkdirAll(nodeDir, 0755))
	nodePath := filepath.Join(nodeDir, "node.exe")
	require.NoError(t, os.WriteFile(nodePath, []byte(""), 0755))

	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	require.NoError(t, p.PostInstall(context.Background(), "prettier", tmpDir, "3.8.3"))

	result, err := os.ReadFile(cmdFile)
	require.NoError(t, err)
	content := string(result)

	assert.Contains(t, content, `"`+nodePath+`"`)
	assert.NotContains(t, content, `%~dp0\node.exe`)
}

// TestNpmProvider_RewriteCmdNodePath_NoNodePattern tests that .cmd files
// without any node.exe pattern are left untouched.
func TestNpmProvider_RewriteCmdNodePath_NoNodePattern(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only .cmd rewrite test")
	}
	p := provider.NewNpmProvider()
	tmpDir := t.TempDir()

	original := "@ECHO OFF\r\nsome-other-tool.exe %*\r\n"
	cmdFile := filepath.Join(tmpDir, "tool.cmd")
	require.NoError(t, os.WriteFile(cmdFile, []byte(original), 0644))

	nodeDir := filepath.Join(tmpDir, "installs", "node", "26.1.0")
	require.NoError(t, os.MkdirAll(nodeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nodeDir, "node.exe"), []byte(""), 0755))

	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	require.NoError(t, p.PostInstall(context.Background(), "tool", tmpDir, "1.0.0"))

	result, err := os.ReadFile(cmdFile)
	require.NoError(t, err)
	assert.Equal(t, original, string(result))
}
