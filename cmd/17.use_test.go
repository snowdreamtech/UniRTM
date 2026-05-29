// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUseCommandStructure(t *testing.T) {
	assert.Equal(t, "use <tool>@<version>", useCmd.Use)
	assert.NotEmpty(t, useCmd.Short)
	assert.NotNil(t, useCmd.RunE)
}

func TestRunUse(t *testing.T) {
	tmpDir := t.TempDir()

	// Switch to tmpDir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := useCmd
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runUse(cmd, []string{"dummy@1.0.0"})
	// It's expected to return error because dummy tool can't be installed automatically
	if err != nil && !strings.Contains(err.Error(), "backend not found") {
		assert.NoError(t, err)
	}

	// Check if unirtm.toml was created
	configFile := filepath.Join(tmpDir, "unirtm.toml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "dummy = \"1.0.0\"")
}

func TestRunUse_Multiple(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := useCmd
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runUse(cmd, []string{"dummy1@1.0.0", "dummy2@2.0.0"})
	if err != nil && !strings.Contains(err.Error(), "backend not found") {
		assert.NoError(t, err)
	}

	configFile := filepath.Join(tmpDir, "unirtm.toml")
	content, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "dummy1 = \"1.0.0\"")
	assert.Contains(t, string(content), "dummy2 = \"2.0.0\"")
}

func TestRunUse_SpecificPath(t *testing.T) {
	tmpDir := t.TempDir()

	targetDir := filepath.Join(tmpDir, "target")

	cmd := useCmd
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	usePath = targetDir
	defer func() { usePath = "" }()

	err := runUse(cmd, []string{"dummy@1.0.0"})
	if err != nil && !strings.Contains(err.Error(), "backend not found") {
		assert.NoError(t, err)
	}

	configFile := filepath.Join(targetDir, "unirtm.toml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)
}

func TestRunUse_Global(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := useCmd
	cmd.SetContext(context.Background())
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	useGlobal = true
	defer func() { useGlobal = false }()

	err := runUse(cmd, []string{"dummy@1.0.0"})
	if err != nil && !strings.Contains(err.Error(), "backend not found") {
		assert.NoError(t, err)
	}

	configFile := filepath.Join(tmpDir, ".config", "unirtm", "unirtm.toml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)
}

func TestRunUse_DryRun(t *testing.T) {
	cmd := useCmd
	cmd.SetContext(context.Background())

	dryRun = true
	defer func() { dryRun = false }()

	err := runUse(cmd, []string{"dummy@1.0.0"})
	assert.NoError(t, err)
}

func TestRunUse_InvalidFormat(t *testing.T) {
	cmd := useCmd
	cmd.SetContext(context.Background())
	err := runUse(cmd, []string{"dummy"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}
