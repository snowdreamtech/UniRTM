// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

func TestPipeline_RunAndFix(t *testing.T) {
	tmpDir := t.TempDir()

	// File 1: Has trailing whitespace & missing EOF newline (fixable)
	f1 := filepath.Join(tmpDir, "dirty.txt")
	_ = os.WriteFile(f1, []byte("hello world   "), 0644)

	// File 2: Has merge conflict marker (unfixable error)
	f2 := filepath.Join(tmpDir, "conflict.txt")
	_ = os.WriteFile(f2, []byte("<<<<<<< HEAD\nfoo\n=======\nbar\n>>>>>>> branch\n"), 0644)

	cfg := guard.GuardConfig{
		Enabled: true,
	}
	pipeline := guard.NewPipeline(tmpDir, cfg)
	ctx := context.Background()

	// 1. Run with Fix=false (dry-run check)
	res, err := pipeline.Run(ctx, []string{"dirty.txt", "conflict.txt"}, guard.ExecuteOptions{
		Fix: false,
	})
	if err != nil {
		t.Fatalf("Pipeline Run failed: %v", err)
	}

	if !res.HasErrors {
		t.Errorf("Expected HasErrors to be true due to merge conflict")
	}
	if res.Summary.ErrorsCount != 1 {
		t.Errorf("Expected 1 error, got %d", res.Summary.ErrorsCount)
	}
	if res.Summary.WarningsCount < 1 {
		t.Errorf("Expected at least 1 warning for dirty.txt, got %d", res.Summary.WarningsCount)
	}

	// 2. Run with Fix=true
	resFix, err := pipeline.Run(ctx, []string{"dirty.txt", "conflict.txt"}, guard.ExecuteOptions{
		Fix: true,
	})
	if err != nil {
		t.Fatalf("Pipeline Run with Fix failed: %v", err)
	}

	if resFix.Summary.FixedCount < 1 {
		t.Errorf("Expected fixed count >= 1, got %d", resFix.Summary.FixedCount)
	}

	// Verify disk content for dirty.txt
	content, _ := os.ReadFile(f1)
	if string(content) != "hello world\n" {
		t.Errorf("Expected cleaned content 'hello world\\n', got %q", string(content))
	}
}
