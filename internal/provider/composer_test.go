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

func TestComposerProvider_Interface(t *testing.T) {
	var _ Provider = (*ComposerProvider)(nil)
}

func TestComposerProvider_Install(t *testing.T) {
	p := NewComposerProvider()
	tmpDir := t.TempDir()

	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	binDir := filepath.Join(tmpDir, "installs", "native-php", "1.0", "bin")
	os.MkdirAll(binDir, 0755)
	phpName := "php"
	scriptContent := []byte("#!/bin/sh\necho installing...")
	if env.RuntimeGOOS == "windows" {
		phpName = "php.exe"
	}
	scriptPath := filepath.Join(binDir, phpName)
	os.WriteFile(scriptPath, scriptContent, 0755)

	// mock composer.phar download by creating it
	composerPhar := filepath.Join(binDir, "composer.phar")
	os.WriteFile(composerPhar, []byte("fake composer"), 0755)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "tool", installPath, "", "1.0.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}

func TestComposerProvider_Install_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	p := NewComposerProvider()
	installPath := filepath.Join(t.TempDir(), "install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to find PHP and Composer")
}

func TestComposerProvider_ListExecutables(t *testing.T) {
	p := NewComposerProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "vendor", "bin")
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
