// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

type mockInstaller struct {
	mu           sync.Mutex
	installed    map[string]bool
	installError error
	called       int
}

func newMockInstaller() *mockInstaller {
	return &mockInstaller{
		installed: make(map[string]bool),
	}
}

func (m *mockInstaller) IsInstalled(ctx context.Context, tool, version, backend string) (bool, *repository.Installation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.installed[tool], nil
}

func (m *mockInstaller) Install(ctx context.Context, toolKey, tool, version, backend string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called++
	if m.installError != nil {
		return m.installError
	}
	m.installed[tool] = true
	return nil
}

func TestConcurrentManager_InstallAll(t *testing.T) {
	installer := newMockInstaller()
	cm := NewConcurrentManager(installer, ConcurrentManagerConfig{MaxConcurrency: 2})

	reqs := []ToolInstallRequest{
		{ToolKey: "go", Tool: "go", Version: "1.20", DependsOn: []string{}},
		{ToolKey: "hugo", Tool: "hugo", Version: "latest", DependsOn: []string{"go"}},
		{ToolKey: "node", Tool: "node", Version: "18", DependsOn: []string{}},
	}

	results, err := cm.InstallAll(context.Background(), reqs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, res := range results {
		if !res.Success {
			t.Errorf("expected tool %s to succeed, got error: %s", res.Tool, res.Error)
		}
	}

	if installer.called != 3 {
		t.Errorf("expected 3 install calls, got %d", installer.called)
	}
}

func TestConcurrentManager_DependencyFailure(t *testing.T) {
	installer := newMockInstaller()
	installer.installError = errors.New("install failed")
	cm := NewConcurrentManager(installer, ConcurrentManagerConfig{MaxConcurrency: 2})

	reqs := []ToolInstallRequest{
		{ToolKey: "go", Tool: "go", Version: "1.20", DependsOn: []string{}},
		{ToolKey: "hugo", Tool: "hugo", Version: "latest", DependsOn: []string{"go"}},
	}

	results, err := cm.InstallAll(context.Background(), reqs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 'go' should fail
	// 'hugo' should skip due to dependency failure
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	goRes, hugoRes := results[0], results[1]
	if goRes.Tool != "go" {
		goRes, hugoRes = results[1], results[0]
	}

	if goRes.Success {
		t.Error("expected go to fail")
	}
	if hugoRes.Success {
		t.Error("expected hugo to skip/fail")
	}
	if hugoRes.Error != "skipped: a dependency failed" {
		t.Errorf("expected hugo error to be 'skipped: a dependency failed', got %q", hugoRes.Error)
	}
}

func TestConcurrentManager_AlreadyInstalled(t *testing.T) {
	installer := newMockInstaller()
	installer.installed["go"] = true
	cm := NewConcurrentManager(installer, ConcurrentManagerConfig{MaxConcurrency: 2})

	reqs := []ToolInstallRequest{
		{ToolKey: "go", Tool: "go", Version: "1.20", DependsOn: []string{}},
	}

	results, err := cm.InstallAll(context.Background(), reqs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].Success || results[0].Error != ErrAlreadyInstalled.Error() {
		t.Errorf("expected already installed status, got %v", results[0])
	}

	if installer.called != 0 {
		t.Errorf("expected 0 install calls, got %d", installer.called)
	}
}

func TestTopoSort_CircularDependency(t *testing.T) {
	reqs := []ToolInstallRequest{
		{Tool: "a", DependsOn: []string{"b"}},
		{Tool: "b", DependsOn: []string{"a"}},
	}

	_, err := topoSort(reqs)
	if err == nil {
		t.Error("expected circular dependency error")
	}
}
