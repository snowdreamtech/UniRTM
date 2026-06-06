// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
	if perm != 0755 {
		t.Errorf("expected permissions 0755, got %04o", perm)
	}
}

func TestInstallBridgeScript_FileContent(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	if err := InstallBridgeScript(context.Background(), tmpDir, "commit-msg"); err != nil {
		t.Fatalf("InstallBridgeScript failed: %v", err)
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "commit-msg")
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	contentStr := string(content)

	// Must be a POSIX sh script
	if !strings.HasPrefix(contentStr, "#!/bin/sh") {
		t.Errorf("expected shebang '#!/bin/sh', got: %q", contentStr[:20])
	}
	// Must not use 'source' (bash extension)
	if strings.Contains(contentStr, " source ") {
		t.Errorf("bridge script must not use 'source' (bash extension); use '.' instead")
	}
	// Must pass arguments correctly
	if !strings.Contains(contentStr, `"$@"`) {
		t.Errorf(`bridge script must contain "$@" to forward git hook args`)
	}
	// Must call unirtm hook run
	if !strings.Contains(contentStr, "unirtm hook run") {
		t.Errorf("bridge script must call 'unirtm hook run'")
	}
}

func TestInstallBridgeScript_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	// Call twice — must not fail on second call
	for i := 0; i < 2; i++ {
		if err := InstallBridgeScript(context.Background(), tmpDir, "pre-commit"); err != nil {
			t.Fatalf("call %d: InstallBridgeScript failed: %v", i+1, err)
		}
	}
}
