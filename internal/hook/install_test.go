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
