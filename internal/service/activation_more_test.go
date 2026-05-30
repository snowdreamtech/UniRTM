// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/provider"
)

func TestActivationManager_GenerateProjectActivation_More(t *testing.T) {
	registry := provider.NewRegistry()
	mockProv := &mockProvider{name: "go-pkg"}
	registry.Register("go-pkg-tool", mockProv)

	m := NewActivationManager("/shims", "/data", registry)
	ctx := context.Background()

	// Use nil envVars and special go: prefix to hit the interception logic
	toolVersions := map[string]string{
		"go:tool": "1.0.0",
	}

	script, err := m.GenerateProjectActivation(ctx, ShellBash, "/project", toolVersions, nil)
	require.NoError(t, err)
	assert.NotNil(t, script)

	// Test Fish and PowerShell scripts
	config := ActivationConfig{
		Shell:         ShellFish,
		Scope:         ScopeProject,
		ShimsDir:      "/shims",
		ProjectDir:    "/project",
		ToolVersions:  map[string]string{"node": "20.0.0"},
		EnvVars:       map[string]string{"FOO": "bar"},
		InjectedPaths: []string{"/data/installs/node/20.0.0/bin"},
		Sources:       []string{"/path/to/source.fish"},
		UseShims:      false,
	}

	fishScript, err := m.GenerateActivationScript(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, fishScript.Content, "set -gx UNIRTM_PATH \"/data/installs/node/20.0.0/bin\"")
	assert.Contains(t, fishScript.Content, "set -gx UNIRTM_NODE_VERSION \"20.0.0\"")
	assert.Contains(t, fishScript.Content, "set -gx FOO \"bar\"")
	assert.Contains(t, fishScript.Content, "source \"/path/to/source.fish\"")

	config.Shell = ShellPowerShell
	psScript, err := m.GenerateActivationScript(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, psScript.Content, fmt.Sprintf("$env:UNIRTM_PATH = $unirtmPaths -join '%c'", os.PathListSeparator))
	assert.Contains(t, psScript.Content, "$env:UNIRTM_NODE_VERSION = \"20.0.0\"")
	assert.Contains(t, psScript.Content, "$env:FOO = \"bar\"")
	assert.Contains(t, psScript.Content, ". \"/path/to/source.fish\"")

	// Test UseShims for Fish and PowerShell
	config.UseShims = true

	config.Shell = ShellFish
	fishScript, err = m.GenerateActivationScript(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, fishScript.Content, "set -gx PATH \"/shims\" $PATH")

	config.Shell = ShellPowerShell
	psScript, err = m.GenerateActivationScript(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, psScript.Content, "$shimsDir = ")
}
