// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/stretchr/testify/require"
)

func TestLockService_Generate(t *testing.T) {
	opts := LockServiceOptions{
		LockfilePath: t.TempDir() + "/unirtm.lock",
		StrictMode:   false,
	}
	ls, err := NewLockService(opts)
	require.NoError(t, err)

	ls.init()

	// Create a minimal backend registry
	reg := backend.NewRegistry()
	ls.SetBackendRegistry(reg)

	ctx := context.Background()
	// Test generate empty
	err = ls.Generate(ctx, nil, GenerateOptions{})
	require.NoError(t, err)

	// Test buildSubset
	specs := map[string]ToolSpec{
		"go":   {Name: "go", Version: "1.20"},
		"node": {Name: "node", Version: "18"},
	}
	subset := buildSubset(specs, []string{"go"})
	require.Len(t, subset, 1)
	require.Contains(t, subset, "go")

	subsetAll := buildSubset(specs, nil)
	require.Len(t, subsetAll, 2)
}

func TestLockService_Validate(t *testing.T) {
	opts := LockServiceOptions{
		LockfilePath: t.TempDir() + "/unirtm.lock",
		StrictMode:   false,
	}
	ls, err := NewLockService(opts)
	require.NoError(t, err)

	err = ls.Validate()
	require.NoError(t, err)
}

func TestLockService_Save(t *testing.T) {
	opts := LockServiceOptions{
		LockfilePath: t.TempDir() + "/unirtm.lock",
		StrictMode:   false,
	}
	ls, err := NewLockService(opts)
	require.NoError(t, err)

	err = ls.save()
	require.NoError(t, err)
}

func TestLockService_BackendForSpec(t *testing.T) {
	opts := LockServiceOptions{
		LockfilePath: t.TempDir() + "/unirtm.lock",
		StrictMode:   false,
	}
	ls, err := NewLockService(opts)
	require.NoError(t, err)

	reg := backend.NewRegistry()
	ls.SetBackendRegistry(reg)

	_, err = ls.backendForSpec("missing", "missing")
	require.Error(t, err)
}
