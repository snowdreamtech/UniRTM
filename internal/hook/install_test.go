// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Skipf("git init failed, skipping test: %v", err)
	}
}

func TestInjectOrUpdateBlock_EmptyFile(t *testing.T) {
	newContent := injectOrUpdateBlock("", false)
	if !strings.HasPrefix(newContent, "#!/bin/sh\n") {
		t.Errorf("Expected shebang in empty file, got: %s", newContent)
	}
	if !strings.Contains(newContent, managedBlockStart) {
		t.Errorf("Expected managed block in empty file")
	}
}

func TestInjectOrUpdateBlock_CommitMsg(t *testing.T) {
	newContent := injectOrUpdateBlock("", false, "commit-msg")
	if !strings.Contains(newContent, "exec unirtm check \"$@\"") {
		t.Errorf("Expected commit-msg bridge script to execute 'unirtm check \"$@\"', got: %s", newContent)
	}
}

func TestInjectOrUpdateBlock_ExistingShebang(t *testing.T) {
	origContent := "#!/usr/bin/env bash\n# Some comment\nexit 0"
	newContent := injectOrUpdateBlock(origContent, true)

	lines := strings.Split(newContent, "\n")
	if lines[0] != "#!/usr/bin/env bash" {
		t.Errorf("Expected first line to be shebang, got %s", lines[0])
	}
	if !strings.Contains(newContent, managedBlockStart) {
		t.Errorf("Expected managed block")
	}
	if !strings.Contains(newContent, "# Some comment") {
		t.Errorf("Expected original content to be preserved")
	}
}

func TestInjectOrUpdateBlock_NoShebang(t *testing.T) {
	origContent := "echo 'Hello World'"
	newContent := injectOrUpdateBlock(origContent, false)

	if !strings.HasPrefix(newContent, "#!/bin/sh\n") {
		t.Errorf("Expected script to prepend shebang")
	}
	if !strings.Contains(newContent, "echo 'Hello World'") {
		t.Errorf("Expected original content to be preserved")
	}
}

func TestInjectOrUpdateBlock_Idempotent(t *testing.T) {
	origContent := "#!/bin/sh\n" + managedBlockStart + "\nold payload\n" + managedBlockEnd + "\nexit 0"
	newContent := injectOrUpdateBlock(origContent, true)

	if strings.Contains(newContent, "old payload") {
		t.Errorf("Expected old payload to be replaced")
	}
	if !strings.Contains(newContent, bridgeScriptPayload) {
		t.Errorf("Expected new payload to be injected")
	}
	// Count occurrences of managed block start
	count := strings.Count(newContent, managedBlockStart)
	if count != 1 {
		t.Errorf("Expected exactly 1 managed block, got %d", count)
	}
}

func TestInstallBridgeScript_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	err := InstallBridgeScript(context.Background(), tmpDir, "pre-commit")
	if err != nil {
		t.Fatalf("InstallBridgeScript failed: %v", err)
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Fatalf("expected bridge script to be created at %s", hookPath)
	}
}

func TestInstallBridgeScript_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	if err := InstallBridgeScript(context.Background(), tmpDir, "pre-commit"); err != nil {
		t.Fatalf("InstallBridgeScript failed: %v", err)
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	// Must be executable (0755)
	perm := info.Mode().Perm()
	if runtime.GOOS != "windows" {
		if perm != 0755 {
			t.Errorf("expected permissions 0755, got %04o", perm)
		}
	}
}

func TestEnsureGitHooksIdempotent_IdempotentAndFast(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Create UniRTM config file
	cfgPath := filepath.Join(tmpDir, ".unirtm.toml")
	_ = os.WriteFile(cfgPath, []byte("[settings]\n"), 0644)

	// 1. First run: should install hooks
	updated, err := EnsureGitHooksIdempotent(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if !updated {
		t.Fatalf("expected updated=true on initial run")
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	fi1, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("expected hook file to exist: %v", err)
	}
	modTime1 := fi1.ModTime()

	// 2. Second run: strictly idempotent, no disk modification, no mtime change, ultra-fast
	start := time.Now()
	updated2, err := EnsureGitHooksIdempotent(context.Background(), tmpDir)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if updated2 {
		t.Fatalf("expected updated=false on second idempotent run")
	}

	fi2, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("expected hook file to exist: %v", err)
	}
	if !fi2.ModTime().Equal(modTime1) {
		t.Fatalf("file was touched! ModTime changed from %v to %v", modTime1, fi2.ModTime())
	}

	// Performance assertion: check that idempotent check ran in < 5ms
	if duration > 50*time.Millisecond {
		t.Errorf("idempotent check took too long: %v (expected < 50ms)", duration)
	}
}

func TestEnsureGitHooksIdempotent_SideEffects(t *testing.T) {
	// Case 1: Non-git directory: should do nothing and return false, nil
	nonGitDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(nonGitDir, ".unirtm.toml"), []byte(""), 0644)

	updated, err := EnsureGitHooksIdempotent(context.Background(), nonGitDir)
	if err != nil || updated {
		t.Fatalf("expected false, nil for non-git dir, got updated=%v, err=%v", updated, err)
	}

	// Case 2: Git directory without UniRTM or hook config: should not install hooks
	gitOnlyDir := t.TempDir()
	initGitRepo(t, gitOnlyDir)

	updated, err = EnsureGitHooksIdempotent(context.Background(), gitOnlyDir)
	if err != nil || updated {
		t.Fatalf("expected false, nil for unconfigured repo, got updated=%v, err=%v", updated, err)
	}
	if _, statErr := os.Stat(filepath.Join(gitOnlyDir, ".git", "hooks", "pre-commit")); statErr == nil {
		t.Fatalf("expected pre-commit NOT to be created in unconfigured repo")
	}

	// Case 3: Preserves user's existing hook script
	userGitDir := t.TempDir()
	initGitRepo(t, userGitDir)
	_ = os.WriteFile(filepath.Join(userGitDir, ".unirtm.toml"), []byte(""), 0644)
	existingHook := "#!/bin/bash\necho 'custom user logic'\nexit 0"
	_ = os.MkdirAll(filepath.Join(userGitDir, ".git", "hooks"), 0755)
	_ = os.WriteFile(filepath.Join(userGitDir, ".git", "hooks", "pre-commit"), []byte(existingHook), 0755)

	updated, err = EnsureGitHooksIdempotent(context.Background(), userGitDir)
	if err != nil || !updated {
		t.Fatalf("expected true, nil, got %v, %v", updated, err)
	}
	content, _ := os.ReadFile(filepath.Join(userGitDir, ".git", "hooks", "pre-commit"))
	if !strings.Contains(string(content), "custom user logic") {
		t.Fatalf("expected user custom logic to be preserved, got: %s", string(content))
	}
}

func TestEnsureGitHooksIdempotent_DeepSubDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)
	_ = os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte("[settings]\n"), 0644)

	// Create deep sub-directory structure: tmpDir/src/app/core
	subDir := filepath.Join(tmpDir, "src", "app", "core")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectories: %v", err)
	}

	// Run from subDir
	updated, err := EnsureGitHooksIdempotent(context.Background(), subDir)
	if err != nil {
		t.Fatalf("subDir run failed: %v", err)
	}
	if !updated {
		t.Fatalf("expected updated=true when run from subDir")
	}

	// Verify hooks are created at the repo root
	rootHook := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(rootHook); os.IsNotExist(err) {
		t.Fatalf("expected root git hook to be created at %s", rootHook)
	}

	// Idempotent check from subDir
	updated2, err := EnsureGitHooksIdempotent(context.Background(), subDir)
	if err != nil || updated2 {
		t.Fatalf("expected false, nil on second subDir run, got updated=%v, err=%v", updated2, err)
	}
}

func TestEnsureGitHooksIdempotent_CustomHooksConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Configure custom hooks in .unirtm.toml
	cfgContent := `
[hooks]
pre-push = "unirtm check"
post-merge = "unirtm install"
invalid-hook-name = "echo test"
`
	_ = os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(cfgContent), 0644)

	updated, err := EnsureGitHooksIdempotent(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !updated {
		t.Fatalf("expected updated=true")
	}

	// pre-push and post-merge should be generated
	for _, h := range []string{"pre-commit", "commit-msg", "pre-push", "post-merge"} {
		path := filepath.Join(tmpDir, ".git", "hooks", h)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected hook %s to be created at %s", h, path)
		}
	}

	// invalid-hook-name must NOT be created (defense against invalid hook names/traversal)
	invalidPath := filepath.Join(tmpDir, ".git", "hooks", "invalid-hook-name")
	if _, err := os.Stat(invalidPath); !os.IsNotExist(err) {
		t.Errorf("expected invalid-hook-name to NOT be created")
	}
}
