// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/require"
)

func TestPubProvider_Interface(t *testing.T) {
	var _ Provider = (*PubProvider)(nil)
}

func TestPubProvider_Install(t *testing.T) {
	p := NewPubProvider()
	tmpDir := t.TempDir()

	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	binDir := filepath.Join(tmpDir, "installs", "native-dart", "1.0", "bin")
	os.MkdirAll(binDir, 0755)
	dartName := "dart"
	scriptContent := []byte("#!/bin/sh\necho installing...")
	if env.RuntimeGOOS == "windows" {
		dartName = "dart.exe"
	}
	scriptPath := filepath.Join(binDir, dartName)
	os.WriteFile(scriptPath, scriptContent, 0755)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "tool", installPath, "", "1.0.0")
	if err != nil {
		if env.RuntimeGOOS == "windows" {
			t.Logf("Ignoring expected execution error on Windows for mock binary: %v", err)
		} else {
			t.Fatalf("install failed: %v", err)
		}
	}
}

func TestPubProvider_Install_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	p := NewPubProvider()
	installPath := filepath.Join(t.TempDir(), "install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to find native dart")
}

func TestPubProvider_ListExecutables(t *testing.T) {
	p := NewPubProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	dummy1Name := "dummy1"
	dummy2Name := "dummy2"
	if env.RuntimeGOOS == "windows" {
		dummy1Name = "dummy1.exe"
		dummy2Name = "dummy2.exe"
	}
	os.WriteFile(filepath.Join(binDir, dummy1Name), []byte(""), 0755)
	os.Chmod(filepath.Join(binDir, dummy1Name), 0755)
	os.WriteFile(filepath.Join(binDir, dummy2Name), []byte(""), 0644)

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	require.NoError(t, err)
	if env.RuntimeGOOS == "windows" {
		require.Len(t, exes, 2)
		require.Contains(t, exes, filepath.Join(binDir, dummy1Name))
		require.Contains(t, exes, filepath.Join(binDir, dummy2Name))
	} else {
		require.Len(t, exes, 1)
		require.Contains(t, exes, filepath.Join(binDir, dummy1Name))
	}
}
