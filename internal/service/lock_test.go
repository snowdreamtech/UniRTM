// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
)

func TestLockService_Lifecycle(t *testing.T) {
	t.Setenv("UNIRTM_LOCKED", "0")
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "unirtm.lock")

	opts := LockServiceOptions{
		LockfilePath: lockPath,
		StrictMode:   false,
	}

	ls, err := NewLockService(opts)
	if err != nil {
		t.Fatalf("failed to create LockService: %v", err)
	}

	if ls.Path() != lockPath {
		t.Errorf("expected path %q, got %q", lockPath, ls.Path())
	}
	if ls.IsStrictMode() {
		t.Error("expected strict mode to be false")
	}
	if !ls.IsEmpty() {
		t.Error("expected lockfile to be empty initially")
	}

	// Record Install
	info := &backend.VersionInfo{
		Version:     "1.0.0",
		DownloadURL: "https://example.com/tool-v1.0.0-linux-amd64.tar.gz",
		Checksum:    "sha256:12345",
		Platform: backend.Platform{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	// Make sure we pass backendName. Let's use "native".
	// Wait, we need to mock or ensure RecordInstall uses the platform we want?
	// RecordInstall uses `platform := backend.CurrentPlatform()` internally?
	// Actually we should just call it and it will use the current OS/Arch.
	err = ls.RecordInstall("core/tool", "native", info)
	if err != nil {
		t.Fatalf("failed to record install: %v", err)
	}

	if ls.IsEmpty() {
		t.Error("expected lockfile not to be empty after record")
	}

	// Resolve
	// Resolve uses the platform recorded in info.
	resolvedInfo, ok := ls.Resolve("core/tool", "1.0.0", info.Platform)
	if !ok {
		t.Error("expected to resolve tool from lockfile")
	} else {
		if resolvedInfo.DownloadURL != info.DownloadURL {
			t.Errorf("expected URL %q, got %q", info.DownloadURL, resolvedInfo.DownloadURL)
		}
		if resolvedInfo.Checksum != info.Checksum {
			t.Errorf("expected checksum %q, got %q", info.Checksum, resolvedInfo.Checksum)
		}
	}

	// Remove
	ls.RemoveTool("core/tool")
	_, ok = ls.Resolve("core/tool", "1.0.0", info.Platform)
	if ok {
		t.Error("expected to not resolve after removal")
	}

	// Verify persistence
	_, err = os.Stat(lockPath)
	if err != nil {
		t.Errorf("expected lockfile to be saved on disk, err: %v", err)
	}
}

func TestLockService_CheckStrict(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "unirtm.lock")

	// strictMode = true
	opts := LockServiceOptions{
		LockfilePath: lockPath,
		StrictMode:   true,
	}

	ls, _ := NewLockService(opts)

	plat := backend.Platform{
		OS:   "linux",
		Arch: "amd64",
	}

	// 1. Tool absent
	err := ls.CheckStrict("core/tool", "1.0.0", plat)
	if err == nil {
		t.Error("expected error for missing tool in strict mode")
	}

	// 2. Add tool but wrong platform
	pe := &lockfile.PlatformEntry{URL: "test"}
	// We need to add an entry for the tool first
	ls.lf.UpsertEntry("core/tool", &lockfile.ToolLockEntry{Version: "1.0.0"})
	ls.lf.UpsertPlatform("core/tool", "1.0.0", lockfile.PlatformKey("windows", "amd64", false), pe)
	err = ls.CheckStrict("core/tool", "1.0.0", plat)
	if err == nil {
		t.Error("expected error for missing platform in strict mode")
	}

	// 3. Match
	ls.lf.UpsertPlatform("core/tool", "1.0.0", lockfile.PlatformKey("linux", "amd64", false), pe)
	err = ls.CheckStrict("core/tool", "1.0.0", plat)
	if err != nil {
		t.Errorf("expected no error for matched platform, got %v", err)
	}
}

func TestDefaultLockFilePath(t *testing.T) {
	path := defaultLockFilePath()
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestLockService_backendForSpec(t *testing.T) {
	opts := LockServiceOptions{}
	ls, _ := NewLockService(opts)

	_, err := ls.backendForSpec("tool", "mock")
	if err == nil || err.Error() != "no backend registry configured" {
		t.Errorf("expected no backend registry configured error, got: %v", err)
	}

	registry := backend.NewRegistry()
	ls.SetBackendRegistry(registry)

	_, err = ls.backendForSpec("tool", "non-existent")
	if err == nil {
		t.Error("expected error for non-existent backend")
	}
}

func TestLockService_init(t *testing.T) {
	ls := &LockService{}
	ls.init()
}

func TestLockService_Generate_Empty(t *testing.T) {
	opts := LockServiceOptions{}
	ls, _ := NewLockService(opts)

	// MUST set registry, otherwise ls.backendRegistry.List() panics in Generate
	registry := backend.NewRegistry()
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	tools := map[string]ToolSpec{
		"go": {Name: "go", Version: "1.20", BackendName: "mock"},
	}

	// Backend not found in registry, should skip smoothly
	err := ls.Generate(ctx, tools, GenerateOptions{})
	if err != nil {
		t.Fatalf("expected no error for skipping unconfigured backend, got %v", err)
	}
}

type mockGenerateBackend struct{}

func (m *mockGenerateBackend) Name() string { return "mockGen" }
func (m *mockGenerateBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	return nil, nil
}
func (m *mockGenerateBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	return nil, nil
}
func (m *mockGenerateBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	if tool == "fail" {
		return nil, fmt.Errorf("simulated error")
	}
	return &backend.VersionInfo{
		Version:      "1.0",
		DownloadURL:  "http://example.com/url",
		Checksum:     "sha256:1234",
		GPGSignature: "sig",
	}, nil
}
func (m *mockGenerateBackend) SupportsChecksum() bool  { return true }
func (m *mockGenerateBackend) SupportsGPG() bool       { return true }
func (m *mockGenerateBackend) AttestationType() string { return "" }
func (m *mockGenerateBackend) IsRecommended() bool     { return false }
func (m *mockGenerateBackend) IsScriptless() bool      { return true }
func (m *mockGenerateBackend) GetReach() string        { return "Small" }
func (m *mockGenerateBackend) IsStable() bool          { return true }
func (m *mockGenerateBackend) SupportsOffline() bool   { return false }
func (m *mockGenerateBackend) Dependencies() []string  { return nil }

func TestLockService_Generate_SuccessAndErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")
	opts := LockServiceOptions{LockfilePath: lfPath}
	ls, _ := NewLockService(opts)

	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	tools := map[string]ToolSpec{
		"success": {Name: "success", Version: "1.0", BackendName: "mockGen"},
		"fail":    {Name: "fail", Version: "1.0", BackendName: "mockGen"},
	}

	err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms: []string{"linux-amd64", "windows-amd64"},
	})
	if err != nil {
		t.Fatalf("expected Generate to handle backend errors internally and return nil, got %v", err)
	}

	// Read lockfile to verify success tool was generated
	parsed, err := lockfile.Load(lfPath)
	if err != nil {
		t.Fatalf("failed to load lockfile: %v", err)
	}

	if parsed.GetEntry("success", "1.0") == nil {
		t.Error("expected 'success' tool in lockfile")
	}
}
