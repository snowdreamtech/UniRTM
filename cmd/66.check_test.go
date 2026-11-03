// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCmd_CleanWorkspace(t *testing.T) {
	harness := NewE2EHarness(t)

	// Clean file in temp dir
	cleanFile := filepath.Join(harness.TmpDir, "clean.txt")
	_ = os.WriteFile(cleanFile, []byte("clean content\n"), 0644)

	stdout, stderr, err := harness.Run("check", cleanFile)
	if err != nil {
		t.Fatalf("Expected check to pass, got err: %v, stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Quality Guard passed") {
		t.Errorf("Expected 'Quality Guard passed' in stdout, got: %s", stdout)
	}
}

func TestCheckCmd_DirtyAndFix(t *testing.T) {
	harness := NewE2EHarness(t)

	// Dirty file with trailing whitespace
	dirtyFile := filepath.Join(harness.TmpDir, "dirty.txt")
	_ = os.WriteFile(dirtyFile, []byte("line with whitespace   \n"), 0644)

	// 1. Run check without fix -> should output warning
	stdout, _, _ := harness.Run("check", dirtyFile)
	if !strings.Contains(stdout, "trailing-whitespace") {
		t.Errorf("Expected trailing-whitespace report, got: %s", stdout)
	}

	// 2. Run fix -> should report fixed and update file on disk
	stdoutFix, _, err := harness.Run("fix", dirtyFile)
	if err != nil {
		t.Fatalf("Expected fix to pass, got: %v", err)
	}
	if !strings.Contains(stdoutFix, "FIXED") && !strings.Contains(stdoutFix, "fixed") {
		t.Errorf("Expected FIXED in output, got: %s", stdoutFix)
	}

	content, _ := os.ReadFile(dirtyFile)
	if string(content) != "line with whitespace\n" {
		t.Errorf("Expected cleaned content, got: %q", string(content))
	}
}

func TestCheckCmd_ConflictMarkerError(t *testing.T) {
	harness := NewE2EHarness(t)

	conflictFile := filepath.Join(harness.TmpDir, "conflict.txt")
	_ = os.WriteFile(conflictFile, []byte("<<<<<<< HEAD\nfoo\n=======\nbar\n>>>>>>> main\n"), 0644)

	_, _, err := harness.Run("check", conflictFile)
	if err == nil {
		t.Fatalf("Expected check to fail on merge conflict markers, but got nil error")
	}
}
