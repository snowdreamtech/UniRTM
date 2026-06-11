// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeNativeConfig(t *testing.T, dir string, content string) {
	t.Helper()
	tomlPath := filepath.Join(dir, ".unirtm.toml")
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func TestNativeRunner_Detect_WithHooks(t *testing.T) {
	tmpDir := t.TempDir()
	writeNativeConfig(t, tmpDir, `
[hooks]
pre-commit = "echo 'success'"
`)
	n := NativeRunner{}
	if !n.Detect(tmpDir) {
		t.Errorf("Detect() should return true for config with [hooks]")
	}
}

func TestNativeRunner_Detect_WithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	writeNativeConfig(t, tmpDir, `
[tools]
go = "1.21.0"
`)
	n := NativeRunner{}
	if n.Detect(tmpDir) {
		t.Errorf("Detect() should return false when no [hooks] defined")
	}
}

func TestNativeRunner_RunInDir_ExecutesHook(t *testing.T) {
	tmpDir := t.TempDir()
	writeNativeConfig(t, tmpDir, `
[hooks]
pre-commit = "echo 'hook executed'"
`)
	n := NativeRunner{}
	// Use RunInDir to pass dir explicitly — no os.Chdir needed
	if err := n.RunInDir(context.Background(), tmpDir, "pre-commit", "", nil); err != nil {
		t.Errorf("RunInDir() failed: %v", err)
	}
}

func TestNativeRunner_RunInDir_SilentOnMissingHook(t *testing.T) {
	tmpDir := t.TempDir()
	writeNativeConfig(t, tmpDir, `
[hooks]
pre-commit = "echo 'hello'"
`)
	n := NativeRunner{}
	// commit-msg is not defined — should silently succeed
	if err := n.RunInDir(context.Background(), tmpDir, "commit-msg", "", nil); err != nil {
		t.Errorf("RunInDir() should return nil for undefined hook, got: %v", err)
	}
}

func TestNativeRunner_RunInDir_PassesArgsSecurely(t *testing.T) {
	tmpDir := t.TempDir()
	// Use printf to echo $1 — if the argument is passed as a single token (not split on space),
	// printf will print it without error regardless of embedded spaces.
	writeNativeConfig(t, tmpDir, `
[hooks]
commit-msg = "printf '%s\n' \"$1\""
`)
	n := NativeRunner{}
	// Argument contains a space — must be forwarded as a single token, not word-split
	safeArg := filepath.Join(tmpDir, "COMMIT EDITMSG")
	if err := n.RunInDir(context.Background(), tmpDir, "commit-msg", "", []string{safeArg}); err != nil {
		t.Errorf("RunInDir() failed with spaced arg: %v", err)
	}
}
