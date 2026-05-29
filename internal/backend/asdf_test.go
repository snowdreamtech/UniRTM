// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAsdfBackend_Name(t *testing.T) {
	b := NewAsdfBackend()
	if b.Name() != "asdf" {
		t.Errorf("expected 'asdf', got '%s'", b.Name())
	}
}

func TestAsdfBackend_Interface(t *testing.T) {
	var _ Backend = (*AsdfBackend)(nil)
}

func TestAsdfBackend_Properties(t *testing.T) {
	b := NewAsdfBackend()
	if b.Dependencies() != nil {
		t.Errorf("expected nil dependencies")
	}
	if b.SupportsChecksum() || b.SupportsGPG() || b.AttestationType() != "" || b.IsRecommended() || b.IsScriptless() || b.GetReach() != "Huge" || b.IsStable() || b.SupportsOffline() {
		t.Errorf("properties not returning expected values")
	}
}

func TestAsdfBackend_ListVersions(t *testing.T) {
	b := NewAsdfBackend()
	tmpDir := t.TempDir()
	b.pluginsPath = tmpDir
	b.registryPath = filepath.Join(tmpDir, "registry")

	toolDir := filepath.Join(tmpDir, "nodejs")
	os.MkdirAll(filepath.Join(toolDir, "bin"), 0755)

	scriptContent := "#!/bin/sh\necho '18.0.0\n20.0.0'"
	scriptPath := filepath.Join(toolDir, "bin", "list-all")
	os.WriteFile(scriptPath, []byte(scriptContent), 0755)

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	versions, err := b.ListVersions(ctx, "nodejs", platform)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
}

func TestAsdfBackend_ResolveVersion(t *testing.T) {
	b := NewAsdfBackend()
	tmpDir := t.TempDir()
	b.pluginsPath = tmpDir
	b.registryPath = filepath.Join(tmpDir, "registry")

	toolDir := filepath.Join(tmpDir, "nodejs")
	os.MkdirAll(filepath.Join(toolDir, "bin"), 0755)

	scriptContent := "#!/bin/sh\necho '20.0.0\n21.0.0'"
	scriptPath := filepath.Join(toolDir, "bin", "list-all")
	os.WriteFile(scriptPath, []byte(scriptContent), 0755)

	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "nodejs", "20.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "20.0.0" {
		t.Errorf("expected version 20.0.0, got %s", info.Version)
	}

	infoLatest, err := b.ResolveVersion(ctx, "nodejs", "latest", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// "21.0.0" comes last in "lines" which means it's pushed to [0] because ListVersions loops backwards
	if infoLatest.Version != "21.0.0" {
		t.Errorf("expected latest version 21.0.0, got %s", infoLatest.Version)
	}
}

func TestAsdfBackend_GetDownloadInfo(t *testing.T) {
	b := NewAsdfBackend()
	tmpDir := t.TempDir()
	b.pluginsPath = tmpDir
	b.registryPath = filepath.Join(tmpDir, "registry")

	toolDir := filepath.Join(tmpDir, "nodejs")
	os.MkdirAll(toolDir, 0755)

	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.GetDownloadInfo(ctx, "nodejs", "20.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "20.0.0" {
		t.Errorf("expected 20.0.0, got %s", info.Version)
	}
}

func TestAsdfBackend_EnsurePlugin_UpdateRegistryAndClone(t *testing.T) {
	gitMockDir := t.TempDir()
	gitMockPath := filepath.Join(gitMockDir, "git")

	gitMockScript := `#!/bin/sh
if [ "$1" = "clone" ]; then
  if echo "$@" | grep "asdf-plugins" > /dev/null; then
    mkdir -p "$3/.git"
    mkdir -p "$3/plugins"
    echo "repository = https://github.com/fake/fake-tool.git" > "$3/plugins/fake-tool"
    exit 0
  fi
  # for plugin clone
  if echo "$@" | grep "fake-tool" > /dev/null; then
    mkdir -p "$3"
    exit 0
  fi
  exit 1
fi
exit 0
`
	if runtime.GOOS == "windows" {
		gitMockPath += ".cmd"
		gitMockScript = `@echo off
if "%~1"=="clone" (
	echo %* | findstr "asdf-plugins" >nul
	if not errorlevel 1 (
		mkdir "%~3\.git"
		mkdir "%~3\plugins"
		echo repository = https://github.com/fake/fake-tool.git > "%~3\plugins\fake-tool"
		exit /b 0
	)
	echo %* | findstr "fake-tool" >nul
	if not errorlevel 1 (
		mkdir "%~3"
		exit /b 0
	)
	exit /b 1
)
exit /b 0
`
	}

	err := os.WriteFile(gitMockPath, []byte(gitMockScript), 0755)
	if err != nil {
		t.Fatalf("failed to create mock git: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", gitMockDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	b := NewAsdfBackend()
	tmpDir := t.TempDir()
	b.pluginsPath = tmpDir
	b.registryPath = filepath.Join(tmpDir, "registry")

	ctx := context.Background()

	// 1. Should update registry and clone the plugin successfully
	pluginDir, err := b.ensurePlugin(ctx, "fake-tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pluginDir != filepath.Join(tmpDir, "fake-tool") {
		t.Errorf("expected %s, got %s", filepath.Join(tmpDir, "fake-tool"), pluginDir)
	}

	// 2. Try again to hit the "already cloned" path
	pluginDir2, err := b.ensurePlugin(ctx, "fake-tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pluginDir2 != pluginDir {
		t.Errorf("expected same dir")
	}

	// 3. Try with an unknown tool which will fail the fallback git clone
	_, err = b.ensurePlugin(ctx, "unknown-tool")
	if err == nil {
		t.Errorf("expected error for unknown tool clone")
	}
}
