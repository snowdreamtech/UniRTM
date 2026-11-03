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

func TestTrustStructure(t *testing.T) {
	assert.Contains(t, trustCmd.Use, "trust", "trustCmd command use should contain 'trust'")
	assert.NotEmpty(t, trustCmd.Short, "trustCmd command short description should not be empty")
	assert.True(t, trustCmd.Run != nil || trustCmd.RunE != nil, "Run or RunE function should be set for trustCmd")

	// Verify flags exist
	assert.NotNil(t, trustCmd.Flags().Lookup("list"), "trustCmd should have --list flag")
	assert.NotNil(t, trustCmd.Flags().Lookup("all"), "trustCmd should have --all flag")
}

// TestRunTrust: trust (no flags) trusts ALL project config files in current dir.
func TestRunTrust(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0755)

	// Create multiple project config files
	for _, f := range []string{"unirtm.toml", ".unirtm.toml", "unirtm.dev.toml"} {
		os.WriteFile(filepath.Join(projectDir, f), []byte(""), 0644)
	}

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	trustList = false
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}
}

// TestRunTrustAll: trust -a trusts ALL project config files + global config.
func TestRunTrustAll(t *testing.T) {
	tmpDir := t.TempDir()
	globalDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", globalDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0755)

	for _, f := range []string{"unirtm.toml", ".unirtm.toml", "unirtm.dev.toml", "unirtm.prod.toml"} {
		os.WriteFile(filepath.Join(projectDir, f), []byte(""), 0644)
	}

	// Create global config file
	os.WriteFile(filepath.Join(globalDir, "unirtm.toml"), []byte(""), 0644)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	trustList = false
	trustAll = true
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}

	// Reset
	trustList = false
	trustAll = false
}

// TestRunTrustAllNoGlobal: trust -a works even if global config doesn't exist.
func TestRunTrustAllNoGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	emptyGlobal := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", emptyGlobal) // no unirtm.toml inside
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "unirtm.toml"), []byte(""), 0644)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	trustList = false
	trustAll = true
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}

	// Reset
	trustList = false
	trustAll = false
}

// TestRunTrustNoFiles: trust in a dir with no config files shows info message.
func TestRunTrustNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	emptyDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(emptyDir, 0755)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	trustList = false
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{emptyDir})
	}
}

// TestRunTrustList: trust -l shows only the current project's trusted config files.
func TestRunTrustList(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "unirtm.toml"), []byte(""), 0644)
	os.WriteFile(filepath.Join(projectDir, "unirtm.dev.toml"), []byte(""), 0644)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Trust first
	trustList = false
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}

	// Then -l: list current project
	trustList = true
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}

	// Reset
	trustList = false
	trustAll = false
}

// TestRunTrustListAll: trust -la shows all globally trusted files.
func TestRunTrustListAll(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "unirtm.toml"), []byte(""), 0644)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Trust first
	trustList = false
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{projectDir})
	}

	// Then -la: list all global
	trustList = true
	trustAll = true
	if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}

	// Reset
	trustList = false
	trustAll = false
}

// TestRunTrustListEmpty: trust -l with no trusted files shows info message.
func TestRunTrustListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)
	t.Setenv("UNIRTM_CACHE_DIR", tmpDir)

	cmd := trustCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	trustList = true
	trustAll = false
	if cmd.Run != nil {
		cmd.Run(cmd, []string{tmpDir})
	}

	trustList = true
	trustAll = true
	if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	}

	// Reset
	trustList = false
	trustAll = false
}

// TestFindAllProjectConfigFiles verifies correct file discovery.
func TestFindAllProjectConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	expected := []string{
		".unirtm.toml", "unirtm.toml",
		"unirtm.dev.toml", ".unirtm.prod.toml",
		"unirtm.yaml", ".unirtm.staging.yaml",
	}
	unexpected := []string{
		"other.toml", "config.yaml", "unirtm_notes.txt",
	}

	for _, f := range append(expected, unexpected...) {
		os.WriteFile(filepath.Join(tmpDir, f), []byte(""), 0644)
	}

	found := findAllProjectConfigFiles(tmpDir)
	foundMap := map[string]bool{}
	for _, p := range found {
		foundMap[filepath.Base(p)] = true
	}

	for _, f := range expected {
		assert.True(t, foundMap[f], "expected config file %q to be found", f)
	}
	for _, f := range unexpected {
		assert.False(t, foundMap[f], "unexpected file %q should not be found", f)
	}
}
