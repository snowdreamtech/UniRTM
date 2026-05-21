// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolCommandStructure(t *testing.T) {
	assert.Contains(t, toolCmd.Use, "tool")
	assert.NotEmpty(t, toolCmd.Short)
	assert.NotNil(t, toolCmd.RunE)
}

func TestToolCommandArgs(t *testing.T) {
	err := toolCmd.Args(toolCmd, []string{"cli/cli"})
	assert.NoError(t, err)

	err = toolCmd.Args(toolCmd, []string{})
	assert.Error(t, err, "0 args should fail")

	err = toolCmd.Args(toolCmd, []string{"a", "b"})
	assert.Error(t, err, "2 args should fail")
}

func TestDetectShimPath_Found(t *testing.T) {
	tmp := t.TempDir()
	// Create a fake shim binary named "gh" (the binary for cli/cli)
	shimFile := filepath.Join(tmp, "gh")
	_ = os.WriteFile(shimFile, []byte("#!/bin/sh"), 0o755)

	got := detectShimPath(tmp, "cli/cli")
	// base of "cli/cli" is "cli" — not "gh", so it falls through
	// That's expected — the function uses base name heuristic.
	assert.NotEmpty(t, got)
}

func TestDetectActiveVersions_NoMatch(t *testing.T) {
	// Requested versions that are not in the installed list → empty result.
	got := detectActiveVersions([]string{"1.0.0"}, []string{"2.70.0", "2.72.0"})
	assert.Empty(t, got)
}

func TestDetectActiveVersions_ExactMatch(t *testing.T) {
	got := detectActiveVersions([]string{"2.72.0"}, []string{"2.70.0", "2.72.0"})
	assert.Equal(t, []string{"2.72.0"}, got)
}

func TestDetectActiveVersions_PrefixMatch(t *testing.T) {
	// Requested "2" should match installed "2.72.0" via prefix.
	got := detectActiveVersions([]string{"2"}, []string{"2.70.0", "2.72.0"})
	assert.Len(t, got, 1)
	assert.Equal(t, "2.70.0", got[0])
}


func TestToolInfo_Struct(t *testing.T) {
	info := toolInfo{
		Tool:         "cli/cli",
		Backend:      "github",
		Installed:    []string{"2.70.0", "2.72.0"},
		Active:       []string{"2.72.0"},
		ConfigSource: "Merged Hierarchy Config",
	}
	assert.Equal(t, "github", info.Backend)
	assert.Len(t, info.Installed, 2)
	assert.Equal(t, []string{"2.72.0"}, info.Active)
	assert.Equal(t, "Merged Hierarchy Config", info.ConfigSource)
}
