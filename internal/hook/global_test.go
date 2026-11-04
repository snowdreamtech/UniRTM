// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetGlobalHooksDir(t *testing.T) {
	dir := GetGlobalHooksDir()
	if dir == "" {
		t.Fatal("expected non-empty global hooks dir")
	}
	if !strings.HasSuffix(dir, filepath.Join("unirtm", "hooks")) && !strings.HasSuffix(dir, "hooks") {
		t.Errorf("unexpected global hooks dir structure: %s", dir)
	}
}

func TestGenerateGlobalHookScript(t *testing.T) {
	script := GenerateGlobalHookScript("pre-commit")
	if !strings.HasPrefix(script, "#!/bin/sh\n") {
		t.Errorf("expected shebang in global hook script")
	}
	if !strings.Contains(script, managedBlockStart) {
		t.Errorf("expected managed block start")
	}
	if !strings.Contains(script, "UNIRTM_BIN") {
		t.Errorf("expected UNIRTM_BIN in script")
	}
	if !strings.Contains(script, "is_unirtm_project") {
		t.Errorf("expected is_unirtm_project helper function")
	}
	if !strings.Contains(script, "exec unirtm check --staged") {
		t.Errorf("expected staged check for pre-commit")
	}
	if !strings.Contains(script, "exec \"$GIT_DIR/hooks/$HOOK_NAME\"") {
		t.Errorf("expected delegation to local git hook for non-unirtm repos")
	}
}

func TestInstallGlobalBridgeScript(t *testing.T) {
	tmpConfig := t.TempDir()
	t.Setenv("UNIRTM_CONFIG_DIR", tmpConfig)

	err := InstallGlobalBridgeScript(context.Background(), "pre-commit")
	if err != nil {
		t.Fatalf("InstallGlobalBridgeScript failed: %v", err)
	}

	hookPath := filepath.Join(tmpConfig, "hooks", "pre-commit")
	fi, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("expected global hook to be created at %s", hookPath)
	}
	if runtime.GOOS != "windows" {
		if fi.Mode().Perm() != 0755 {
			t.Errorf("expected 0755 perm, got %04o", fi.Mode().Perm())
		}
	}

	// Test invalid hook name
	err = InstallGlobalBridgeScript(context.Background(), "../evil")
	if err == nil {
		t.Fatal("expected error for path traversal hook name")
	}
}
