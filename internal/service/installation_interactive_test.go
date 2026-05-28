// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestInstallationManager_SelectVersionInteractive(t *testing.T) {
	backendReg := backend.NewRegistry()
	providerReg := provider.NewRegistry()
	im := &InstallationManager{
		backendRegistry:  backendReg,
		providerRegistry: providerReg,
	}
	_, err := im.SelectVersionInteractive(context.Background(), "node", "")
	require.Error(t, err) // Expect an error since providerRegistry is empty
}

func TestInstallationManager_SortTools(t *testing.T) {
	backendReg := backend.NewRegistry()
	providerReg := provider.NewRegistry()
	im := &InstallationManager{
		backendRegistry:  backendReg,
		providerRegistry: providerReg,
	}
	tools := map[string]config.ToolConfig{
		"node": {Version: "18.0.0"},
	}
	sorted := im.SortTools(tools)
	require.Len(t, sorted, 1)
	require.Equal(t, "node", sorted[0].OriginalName)
	require.Equal(t, "node", sorted[0].ToolName)
	require.Equal(t, "18.0.0", sorted[0].Version)
}
