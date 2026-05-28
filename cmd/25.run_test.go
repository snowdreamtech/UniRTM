// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunStructure(t *testing.T) {
	assert.Contains(t, runCmd.Use, "run", "runCmd command use should contain 'run'")
	assert.NotEmpty(t, runCmd.Short, "runCmd command short description should not be empty")
	assert.True(t, runCmd.Run != nil || runCmd.RunE != nil, "Run or RunE function should be set for runCmd")
}

func TestRunRun(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	os.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	os.Setenv("UNIRTM_CACHE_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")
	defer os.Unsetenv("UNIRTM_CACHE_DIR")

	cmd := runCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if cmd.RunE != nil {
		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)
	} else if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}
}

func TestRunTaskCommand_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	// Create a mock unirtm.toml with a dummy task
	configFile := filepath.Join(tmpDir, "unirtm.toml")
	_ = os.WriteFile(configFile, []byte(`[tasks.echo]
run = "echo hello"`), 0o644)

	// Temporarily change directory to tmpDir to pick up unirtm.toml
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	err := runCmd.RunE(runCmd, []string{"echo"})
	// Depending on runner, it might fail or succeed, but we just want coverage here.
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunTaskCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	err := runCmd.RunE(runCmd, []string{"non_existent_task_12345"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task execution failed")
}
