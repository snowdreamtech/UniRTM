// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDotnetProvider_Interface(t *testing.T) {
	var _ Provider = (*DotnetProvider)(nil)
}

func TestDotnetProvider_FindDotnet(t *testing.T) {
	p := NewDotnetProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	dotnetName := "dotnet"
	scriptContent := []byte("#!/bin/sh\necho dotnet")
	if runtime.GOOS == "windows" {
		dotnetName = "dotnet.bat"
		scriptContent = []byte("@echo dotnet")
	}
	scriptPath := filepath.Join(binDir, dotnetName)
	os.WriteFile(scriptPath, scriptContent, 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	dotnet, err := p.findDotnet()
	if err != nil {
		t.Fatalf("expected to find dotnet, but got error: %v", err)
	}
	if dotnet != scriptPath {
		t.Errorf("expected %s, got %s", scriptPath, dotnet)
	}
}

func TestDotnetProvider_Install(t *testing.T) {
	p := NewDotnetProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	dotnetName := "dotnet"
	scriptContent := []byte("#!/bin/sh\necho installing...")
	if runtime.GOOS == "windows" {
		dotnetName = "dotnet.bat"
		scriptContent = []byte("@echo installing...")
	}
	scriptPath := filepath.Join(binDir, dotnetName)
	os.WriteFile(scriptPath, scriptContent, 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "tool", installPath, "", "1.0.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}

func TestDotnetProvider_findDotnet(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)

	p := NewDotnetProvider()

	// Create fake dotnet installation
	dotnetDir := filepath.Join(tmpData, "installs", "dotnet", "7.0.0")
	os.MkdirAll(dotnetDir, 0755)
	dotnetPath := filepath.Join(dotnetDir, "dotnet")
	os.WriteFile(dotnetPath, []byte("fake binary"), 0755)

	found, err := p.findDotnet()
	require.NoError(t, err)
	require.Equal(t, dotnetPath, found)
}

func TestDotnetProvider_Install_DotnetNotFound(t *testing.T) {
	t.Setenv("PATH", "")

	p := NewDotnetProvider()
	installPath := filepath.Join(t.TempDir(), "dotnet_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	if err == nil {
		t.Fatalf("expected error when dotnet is not found")
	}
}

func TestDotnetProvider_ListExecutables(t *testing.T) {
	p := NewDotnetProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin") // Wait, dotnet usually returns bin or root dir?
	os.MkdirAll(binDir, 0755)

	dummy1Name := "dummy1"
	dummy2Name := "dummy2"
	if runtime.GOOS == "windows" {
		dummy1Name = "dummy1.exe"
		dummy2Name = "dummy2.exe"
	}
	os.WriteFile(filepath.Join(binDir, dummy1Name), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, dummy2Name), []byte(""), 0644)

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exes) != 1 {
		t.Errorf("expected 1 executable, got %d", len(exes))
	}
}
