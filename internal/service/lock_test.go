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
	_, err := ls.Generate(ctx, tools, GenerateOptions{})
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

	_, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms:       []string{"linux-amd64", "windows-amd64"},
		AllowIncomplete: true,
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

func TestLockService_Generate_RetainsExistingOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	// Pre-create lockfile containing valid entry for python@3.14.7
	lf := lockfile.New(lfPath)
	lf.UpsertEntry("python", &lockfile.ToolLockEntry{
		Version:   "3.14.7",
		Backend:   "mockGen",
		Platforms: make(map[string]*lockfile.PlatformEntry),
	})
	lf.UpsertPlatform("python", "3.14.7", "linux-amd64", &lockfile.PlatformEntry{
		Checksum: "sha256:python123",
		URL:      "http://example.com/python.tar.gz",
	})
	if err := lf.Save(); err != nil {
		t.Fatalf("failed to seed lockfile: %v", err)
	}

	opts := LockServiceOptions{LockfilePath: lfPath}
	ls, err := NewLockService(opts)
	if err != nil {
		t.Fatalf("failed to create lock service: %v", err)
	}

	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	tools := map[string]ToolSpec{
		"python": {Name: "python", Version: "3.14.7", BackendName: "mockGen"},
		"fail":   {Name: "fail", Version: "1.0", BackendName: "mockGen"},
	}

	// Requesting "fail" (fails) and "python" (where "python" will fail if tool=="fail" or if mock returns error)
	// Here mock returns error for "fail". For python, let's test when mock succeeds or fails.
	report, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms: []string{"linux-amd64"},
	})
	if err != nil {
		t.Fatalf("expected Generate to complete without error, got %v", err)
	}

	// Read lockfile to verify 'python' entry was NOT deleted despite 'fail' error
	parsed, err := lockfile.Load(lfPath)
	if err != nil {
		t.Fatalf("failed to load lockfile: %v", err)
	}

	pyEntry := parsed.GetPlatform("python", "3.14.7", "linux-amd64")
	if pyEntry == nil || pyEntry.Checksum != "sha256:python123" {
		t.Errorf("expected python lock entry to be retained in lockfile, got %v", pyEntry)
	}

	if report.IsComplete() {
		t.Errorf("expected report to indicate missing platforms for 'fail'")
	}
}

func TestLockService_Generate_ForceBypassesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	// Seed lockfile with fail entry
	lf := lockfile.New(lfPath)
	lf.UpsertEntry("fail", &lockfile.ToolLockEntry{
		Version:   "1.0",
		Backend:   "mockGen",
		Platforms: make(map[string]*lockfile.PlatformEntry),
	})
	lf.UpsertPlatform("fail", "1.0", "linux-amd64", &lockfile.PlatformEntry{
		Checksum: "sha256:oldfail",
		URL:      "http://example.com/oldfail",
	})
	_ = lf.Save()

	ls, _ := NewLockService(LockServiceOptions{LockfilePath: lfPath})
	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	tools := map[string]ToolSpec{
		"fail": {Name: "fail", Version: "1.0", BackendName: "mockGen"},
	}

	// With Force: true, fail resolution failure should NOT retain old entry and should report missing
	report, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms: []string{"linux-amd64"},
		Force:     true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.IsComplete() {
		t.Error("expected missing platform report when Force=true and resolution fails")
	}
}

func TestLockService_Generate_OrphanCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	// Seed lockfile with orphan tool "removed_tool" and kept tool "success"
	lf := lockfile.New(lfPath)
	lf.UpsertEntry("removed_tool", &lockfile.ToolLockEntry{
		Version:   "1.0",
		Backend:   "mockGen",
		Platforms: make(map[string]*lockfile.PlatformEntry),
	})
	lf.UpsertPlatform("removed_tool", "1.0", "linux-amd64", &lockfile.PlatformEntry{
		Checksum: "sha256:orphan",
		URL:      "http://example.com/orphan",
	})
	_ = lf.Save()

	ls, _ := NewLockService(LockServiceOptions{LockfilePath: lfPath})
	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	// Config tools now only has "success", "removed_tool" was deleted from config
	tools := map[string]ToolSpec{
		"success": {Name: "success", Version: "1.0", BackendName: "mockGen"},
	}

	_, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms: []string{"linux-amd64"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := lockfile.Load(lfPath)
	if err != nil {
		t.Fatalf("failed to load lockfile: %v", err)
	}

	if parsed.GetEntry("removed_tool", "1.0") != nil {
		t.Error("expected orphan tool 'removed_tool' to be removed from lockfile during full lock generation")
	}
	if parsed.GetEntry("success", "1.0") == nil {
		t.Error("expected 'success' tool to be in lockfile")
	}
}

func TestLockService_Generate_AtomicWriteFirewall(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	ls, _ := NewLockService(LockServiceOptions{LockfilePath: lfPath})
	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	ctx := context.Background()
	tools := map[string]ToolSpec{
		"success": {Name: "success", Version: "1.0", BackendName: "mockGen"},
		"fail":    {Name: "fail", Version: "1.0", BackendName: "mockGen"},
	}

	// Generating incomplete lockfile without AllowIncomplete=true MUST NOT write to disk
	report, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms:       []string{"linux-amd64"},
		AllowIncomplete: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.IsComplete() {
		t.Error("expected report to be incomplete")
	}

	if _, err := os.Stat(lfPath); !os.IsNotExist(err) {
		t.Errorf("expected lockfile %s NOT to be created on disk when lock is incomplete and AllowIncomplete is false", lfPath)
	}

	// Generating with AllowIncomplete=true MUST write to disk
	_, _ = ls.Generate(ctx, tools, GenerateOptions{
		Platforms:       []string{"linux-amd64"},
		AllowIncomplete: true,
	})

	if _, err := os.Stat(lfPath); os.IsNotExist(err) {
		t.Errorf("expected lockfile %s to be created when AllowIncomplete is true", lfPath)
	}
}

func TestLockService_LegacyKeyMigration(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	ls, _ := NewLockService(LockServiceOptions{LockfilePath: lfPath})
	registry := backend.NewRegistry()
	registry.Register(&mockGenerateBackend{})
	ls.SetBackendRegistry(registry)

	// Pre-populate lockfile with a legacy unprefixed key "legacy/tool" that has a locked platform "windows-amd64"
	ls.lf.UpsertEntry("legacy/tool", &lockfile.ToolLockEntry{
		Version: "1.0",
		Backend: "mockGen",
	})
	ls.lf.UpsertPlatform("legacy/tool", "1.0", "windows-amd64", &lockfile.PlatformEntry{
		URL:      "http://example.com/legacy-win.zip",
		Checksum: "sha256:123456",
	})

	// Config has canonical prefixed key "github:legacy/tool"
	tools := map[string]ToolSpec{
		"github:legacy/tool": {
			Name:        "legacy/tool",
			Version:     "1.0",
			BackendName: "mockGen",
		},
	}

	ctx := context.Background()
	_, err := ls.Generate(ctx, tools, GenerateOptions{
		Platforms:       []string{"linux-amd64"},
		AllowIncomplete: true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify legacy unprefixed key was removed
	if ls.lf.GetEntry("legacy/tool", "1.0") != nil {
		t.Error("expected legacy unprefixed key 'legacy/tool' to be removed from lockfile")
	}

	// Verify canonical prefixed key has the preserved "windows-amd64" platform entry
	winPlat := ls.lf.GetPlatform("github:legacy/tool", "1.0", "windows-amd64")
	if winPlat == nil || winPlat.URL != "http://example.com/legacy-win.zip" {
		t.Errorf("expected migrated windows-amd64 platform URL, got %v", winPlat)
	}
}




