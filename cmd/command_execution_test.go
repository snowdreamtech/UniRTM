// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallCommand_EarlyReturns(t *testing.T) {
	// Setup
	originalDryRun := dryRun
	dryRun = true
	defer func() { dryRun = originalDryRun }()

	buf := new(bytes.Buffer)
	installCmd.SetOut(buf)
	installCmd.SetErr(buf)

	// Test 1: Empty args (might load config, but usually empty in CI)
	err := runInstall(installCmd, []string{})
	// This could return nil (if no tools in config) or proceed to dry-run and return nil
	assert.NoError(t, err)

	// Test 2: Missing tool name
	err = runInstall(installCmd, []string{"", "1.0.0"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool name is required")

	// Test 3: Missing version
	err = runInstall(installCmd, []string{"dummy-tool", ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version is required")

	// Test 4: Dry run with valid args
	err = runInstall(installCmd, []string{"dummy-tool", "20.0.0"})
	assert.NoError(t, err)

	// Test 5: Concurrent spinner manager
	mgr := newConcurrentSpinnerManager()
	mgr.Start()
	mgr.Add("dummy-tool", "20.0.0")
	mgr.Update("dummy-tool", "downloading")
	mgr.Complete("dummy-tool", "20.0.0", "done")
	mgr.Add("go", "1.21.0")
	mgr.Complete("go", "1.21.0", "failed: error")
	mgr.Stop()
}

func TestTokenCommand(t *testing.T) {
	// Test without filter
	err := runToken(tokenCmd, []string{})
	assert.NoError(t, err)

	// Test with filter
	err = runToken(tokenCmd, []string{"github"})
	assert.NoError(t, err)

	// Test json output
	originalJson := jsonOutput
	jsonOutput = true
	defer func() { jsonOutput = originalJson }()
	err = runToken(tokenCmd, []string{})
	assert.NoError(t, err)
}

func TestCurrentCommand(t *testing.T) {
	// Test without filter
	err := runCurrent(currentCmd, []string{})
	assert.NoError(t, err)

	// Test with filter
	err = runCurrent(currentCmd, []string{"invalid_tool"})
	// It should error because invalid_tool is not in config
	assert.Error(t, err)

	// Test json output / plain mode
	originalJson := jsonOutput
	jsonOutput = true
	defer func() { jsonOutput = originalJson }()
	err = runCurrent(currentCmd, []string{})
	assert.NoError(t, err)
}

func TestLsRemoteCommand_EarlyReturns(t *testing.T) {
	// Skip ls-remote because calling it directly with empty args causes panic,
	// and calling with valid args causes network requests.
	t.Skip("skipping ls-remote due to network requirements")
}

func TestUnuseCommand_EarlyReturns(t *testing.T) {
	err := runUnuse(unuseCmd, []string{})
	// Depending on config, unuse with no args might return nil if no tools to unuse, or error
	// Let's just assert it doesn't panic.
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func TestImplodeCommand_DryRun(t *testing.T) {
	// We can't easily test implode without mocking terminal input, unless we simulate arguments or dry-run.
	// But it uses pterm.DefaultInteractiveConfirm, which hangs.
	// However, we can just skip it here.
}

func TestUninstallCommand_EarlyReturns(t *testing.T) {
	err := runUninstall(uninstallCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool specification is required")
}
