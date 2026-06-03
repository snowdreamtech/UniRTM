// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNpmProvider_WindowsWrappers(t *testing.T) {
	p := NewNpmProvider()

	dir := t.TempDir()

	// mock node install
	dataDir := filepath.Join(dir, "data")
	nodeDir := filepath.Join(dataDir, "installs", "node", "20.0.0")
	os.MkdirAll(nodeDir, 0755)
	nodeExe := filepath.Join(nodeDir, "node.exe")
	os.WriteFile(nodeExe, []byte(""), 0755)

	// mock UNIRTM_DATA_DIR
	t.Setenv("UNIRTM_DATA_DIR", dataDir)

	installPath := filepath.Join(dir, "npm_install")
	os.MkdirAll(installPath, 0755)

	cmdFile := filepath.Join(installPath, "test.cmd")
	os.WriteFile(cmdFile, []byte(`@ECHO off
GOTO start
:find_dp0
SET dp0=%~dp0
EXIT /b
:start
SETLOCAL
CALL :find_dp0

IF EXIST "%dp0%\node.exe" (
  SET "_prog=%dp0%\node.exe"
) ELSE (
  SET "_prog=node"
  SET PATHEXT=%PATHEXT:;.JS;=;%
)

endLocal & goto #_undefined_# 2>NUL || title %COMSPEC% & "%_prog%"  "%dp0%\node_modules\npm\bin\npm-cli.js" %*`), 0755)

	p.fixWindowsCmdWrappers(installPath)
	p.rewriteCmdNodePath(cmdFile, "C:\\mock\\node.exe")
	p.findNodeExe()
	p.findNodeBinDir()
}
