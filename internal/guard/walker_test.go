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

func TestWalker_AllModeAndSkipDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory tree
	// - src/main.go
	// - src/util.ts
	// - node_modules/bad.js (should be skipped)
	// - dist/bundle.js (should be skipped)
	// - .git/HEAD (should be skipped)
	// - README.md
	// - test.min.js (should be excluded by glob)
	mustWrite := func(rel string, content string) {
		p := filepath.Join(tmpDir, rel)
		_ = os.MkdirAll(filepath.Dir(p), 0755)
		_ = os.WriteFile(p, []byte(content), 0644)
	}

	mustWrite("src/main.go", "package main\n")
	mustWrite("src/util.ts", "export const x = 1;\n")
	mustWrite("node_modules/bad.js", "console.log('bad');\n")
	mustWrite("dist/bundle.js", "console.log('dist');\n")
	mustWrite(".git/HEAD", "ref: refs/heads/main\n")
	mustWrite("README.md", "# Hello\n")
	mustWrite("src/test.min.js", "minified\n")

	ctx := context.Background()

	// 1. Test CollectFiles all mode without extension filter
	files, err := guard.CollectFiles(ctx, guard.WalkOptions{
		WorkspaceDir: tmpDir,
		Mode:         "all",
		ExcludeFiles: []string{"*.min.js"},
	})
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	hasFile := func(target string) bool {
		target = filepath.Clean(target)
		for _, f := range files {
			if filepath.Clean(f) == target {
				return true
			}
		}
		return false
	}

	if !hasFile("src/main.go") {
		t.Errorf("Expected src/main.go in results, got %v", files)
	}
	if !hasFile("src/util.ts") {
		t.Errorf("Expected src/util.ts in results, got %v", files)
	}
	if !hasFile("README.md") {
		t.Errorf("Expected README.md in results, got %v", files)
	}
	if hasFile("node_modules/bad.js") {
		t.Errorf("node_modules should have been skipped!")
	}
	if hasFile("dist/bundle.js") {
		t.Errorf("dist should have been skipped!")
	}
	if hasFile("src/test.min.js") {
		t.Errorf("test.min.js should have been excluded by pattern!")
	}

	// 2. Test extension filtering: only .go
	goFiles, err := guard.CollectFiles(ctx, guard.WalkOptions{
		WorkspaceDir: tmpDir,
		Mode:         "all",
		Extensions:   []string{".go"},
	})
	if err != nil {
		t.Fatalf("CollectFiles for .go failed: %v", err)
	}
	if len(goFiles) != 1 || filepath.Clean(goFiles[0]) != filepath.Clean("src/main.go") {
		t.Errorf("Expected only src/main.go, got: %v", goFiles)
	}

	// 3. Test IncludeDirs filtering
	srcFiles, err := guard.CollectFiles(ctx, guard.WalkOptions{
		WorkspaceDir: tmpDir,
		Mode:         "all",
		IncludeDirs:  []string{"src"},
		ExcludeFiles: []string{"*.min.js"},
	})
	if err != nil {
		t.Fatalf("CollectFiles for src failed: %v", err)
	}
	for _, f := range srcFiles {
		if !hasPrefixOrEqual(f, "src") {
			t.Errorf("File %s is outside include_dirs 'src'", f)
		}
	}
}

func hasPrefixOrEqual(path, prefix string) bool {
	p := filepath.Clean(path)
	pref := filepath.Clean(prefix)
	return p == pref || len(p) > len(pref) && p[:len(pref)] == pref
}
