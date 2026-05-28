// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellCommandStructure(t *testing.T) {
	assert.Equal(t, "shell <tool>@<version> [tool@version ...]", shellCmd.Use)
	assert.NotEmpty(t, shellCmd.Short)
	assert.NotNil(t, shellCmd.RunE)
}

func TestRunShell(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := shellCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Ensure shell resolves to something specific
	os.Setenv("SHELL", "/bin/bash")
	defer os.Unsetenv("SHELL")

	err := runShell(cmd, []string{"node@20.0.0"})
	assert.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	script := string(out)

	assert.Contains(t, script, "export UNIRTM_NODE_VERSION=\"20.0.0\"")
}

func TestRunShell_MultipleTools(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := shellCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	os.Setenv("SHELL", "/bin/bash")
	defer os.Unsetenv("SHELL")

	err := runShell(cmd, []string{"node@20.0.0", "python@3.11.0"})
	assert.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	script := string(out)

	assert.Contains(t, script, "export UNIRTM_NODE_VERSION=\"20.0.0\"")
	assert.Contains(t, script, "export UNIRTM_PYTHON_VERSION=\"3.11.0\"")
}

func TestRunShell_InvalidFormat(t *testing.T) {
	cmd := shellCmd
	err := runShell(cmd, []string{"node"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestRunShell_DryRun(t *testing.T) {
	cmd := shellCmd

	dryRun = true
	defer func() { dryRun = false }()

	err := runShell(cmd, []string{"node@20.0.0"})
	assert.NoError(t, err)
}
