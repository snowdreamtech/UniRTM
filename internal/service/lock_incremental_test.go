// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/lockfile"
)

type mockFailingBackend struct{}

func (m *mockFailingBackend) Name() string { return "failing" }
func (m *mockFailingBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	return nil, nil
}
func (m *mockFailingBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	return nil, nil
}
func (m *mockFailingBackend) GetDownloadInfo(ctx context.Context, tool, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	return nil, fmt.Errorf("simulated network failure")
}
func (m *mockFailingBackend) SupportsChecksum() bool  { return true }
func (m *mockFailingBackend) SupportsGPG() bool       { return true }
func (m *mockFailingBackend) AttestationType() string { return "" }
func (m *mockFailingBackend) IsRecommended() bool     { return false }
func (m *mockFailingBackend) IsScriptless() bool      { return true }
func (m *mockFailingBackend) GetReach() string        { return "Small" }
func (m *mockFailingBackend) IsStable() bool          { return true }
func (m *mockFailingBackend) SupportsOffline() bool   { return false }
func (m *mockFailingBackend) Dependencies() []string  { return nil }

func TestLockService_IncrementalPreservationAndFallback(t *testing.T) {
	tmpDir := t.TempDir()
	lfPath := filepath.Join(tmpDir, "unirtm.lock")

	// 1. Create an initial lockfile with two tools: pipx:yamllint and pipx:pre-commit (no checksums)
	lf := lockfile.New(lfPath)
	lf.UpsertEntry("pipx:yamllint", &lockfile.ToolLockEntry{
		Version: "1.38.0",
		Backend: "pipx",
	})
	lf.UpsertPlatform("pipx:yamllint", "1.38.0", "linux-amd64", &lockfile.PlatformEntry{
		URL: "https://pipx.local/yamllint",
	})

	lf.UpsertEntry("pipx:pre-commit", &lockfile.ToolLockEntry{
		Version: "4.6.2",
		Backend: "pipx",
	})
	lf.UpsertPlatform("pipx:pre-commit", "4.6.2", "linux-amd64", &lockfile.PlatformEntry{
		URL: "https://pipx.local/pre-commit",
	})

	if err := lf.Save(); err != nil {
		t.Fatalf("failed to seed lockfile: %v", err)
	}

	// 2. Initialize LockService with seeded lockfile
	ls, err := NewLockService(LockServiceOptions{LockfilePath: lfPath})
	if err != nil {
		t.Fatalf("failed to create LockService: %v", err)
	}

	reg := backend.NewRegistry()
	reg.Register(&mockFailingBackend{})
	ls.SetBackendRegistry(reg)

	// 3. Run Generate for pipx:yamllint with failing backend (simulating network error)
	tools := map[string]ToolSpec{
		"pipx:yamllint": {Name: "yamllint", Version: "1.38.0", BackendName: "failing"},
	}

	report, err := ls.Generate(context.Background(), tools, GenerateOptions{
		Tools:     []string{"pipx:yamllint"},
		Platforms: []string{"linux-amd64"},
		Force:     false,
	})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	if !report.IsComplete() {
		t.Errorf("expected report to be complete due to fallback retention, got missing: %+v", report.Missing)
	}

	// Save lockfile
	if err := ls.save(); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	// 4. Reload lockfile and verify BOTH yamllint and pre-commit exist
	reloaded, err := lockfile.Load(lfPath)
	if err != nil {
		t.Fatalf("failed to reload lockfile: %v", err)
	}

	yamllintPlat := reloaded.GetPlatform("pipx:yamllint", "1.38.0", "linux-amd64")
	if yamllintPlat == nil || yamllintPlat.URL != "https://pipx.local/yamllint" {
		t.Errorf("expected yamllint lock entry to be retained, got: %v", yamllintPlat)
	}

	precommitPlat := reloaded.GetPlatform("pipx:pre-commit", "4.6.2", "linux-amd64")
	if precommitPlat == nil || precommitPlat.URL != "https://pipx.local/pre-commit" {
		t.Errorf("expected pre-commit lock entry to be preserved across incremental run, got: %v", precommitPlat)
	}
}
