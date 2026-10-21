// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSlotRule_MetadataAndMatching(t *testing.T) {
	slot := ToolSlot{
		ID:                "test.formatter",
		Description:       "Test Formatter",
		Category:          CategoryFormat,
		Phase:             PhaseFormatter,
		DefaultToolName:   "nonexistent-tool-xyz-12345",
		DefaultCommand:    "nonexistent-tool-xyz-12345 {{files}}",
		DefaultExtensions: []string{".go"},
	}

	rule := NewSlotRule(slot)

	if rule.ID() != "slot:test.formatter" {
		t.Fatalf("expected ID slot:test.formatter, got %s", rule.ID())
	}
	if rule.Category() != CategoryFormat {
		t.Fatalf("expected category %s, got %s", CategoryFormat, rule.Category())
	}
	if rule.Phase() != PhaseFormatter {
		t.Fatalf("expected phase %s, got %s", PhaseFormatter, rule.Phase())
	}

	files := []string{"main.go", "app.py", "script.sh", "sub/pkg.go"}
	matched := rule.FilterMatchingFiles(files)
	if len(matched) != 2 || matched[0] != "main.go" || matched[1] != "sub/pkg.go" {
		t.Fatalf("unexpected matching files: %v", matched)
	}

	// Should not run because executable does not exist
	if rule.ShouldRun(files) {
		t.Fatalf("expected ShouldRun to return false for nonexistent tool")
	}

	// Non-matching file list
	if rule.ShouldRun([]string{"app.py", "script.sh"}) {
		t.Fatalf("expected ShouldRun to return false for non-matching files")
	}
}

func TestSlotRule_GofmtExecution(t *testing.T) {
	tmpDir := t.TempDir()
	goodFile := filepath.Join(tmpDir, "good.go")
	if err := os.WriteFile(goodFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	slot := ToolSlot{
		ID:                "go.formatter",
		Description:       "Go canonical code formatter",
		Category:          CategoryFormat,
		Phase:             PhaseFormatter,
		DefaultToolName:   "gofmt",
		DefaultCommand:    "gofmt -l {{files}}",
		FixCommand:        "gofmt -w {{files}}",
		DefaultExtensions: []string{".go"},
	}

	rule := NewSlotRule(slot)
	if !rule.ShouldRun([]string{goodFile}) {
		t.Skip("gofmt is not available in test environment")
	}

	// Check clean file
	issues, err := rule.Execute(context.Background(), []string{goodFile}, ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues on clean formatted file, got %d: %v", len(issues), issues)
	}

	// Create unformatted file
	badFile := filepath.Join(tmpDir, "bad.go")
	if err := os.WriteFile(badFile, []byte("package main\n\nfunc main(   ) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	issues, err = rule.Execute(context.Background(), []string{badFile}, ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if len(issues) == 0 {
		t.Fatalf("expected formatting issue detected on unformatted file")
	}
	if !issues[0].Fixable {
		t.Fatalf("expected issue to be marked as fixable")
	}

	// Now execute with Fix: true
	issues, err = rule.Execute(context.Background(), []string{badFile}, ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatalf("unexpected fix execution error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues after fix, got %d", len(issues))
	}

	// Verify file is now formatted
	issues, err = rule.Execute(context.Background(), []string{badFile}, ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected check execution error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues on fixed file, got %d", len(issues))
	}
}
