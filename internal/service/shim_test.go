// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

func TestGenerator_GenerateShim(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)

	ctx := context.Background()
	err := g.GenerateShim(ctx, "go")
	if err != nil {
		t.Fatalf("failed to generate shim: %v", err)
	}

	if !g.ShimExists("go") {
		t.Error("expected shim to exist")
	}

	shims, err := g.ListShims()
	if err != nil {
		t.Fatalf("failed to list shims: %v", err)
	}

	found := false
	for _, shim := range shims {
		if shim == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'go' in listed shims")
	}

	err = g.RemoveShim(ctx, "go")
	if err != nil {
		t.Fatalf("failed to remove shim: %v", err)
	}

	if g.ShimExists("go") {
		t.Error("expected shim to be removed")
	}
}

func TestToolVersionEnvVar(t *testing.T) {
	tests := []struct {
		tool string
		want string
	}{
		{"node", "UNIRTM_NODE_VERSION"},
		{"go", "UNIRTM_GO_VERSION"},
		{"rust-analyzer", "UNIRTM_RUST_ANALYZER_VERSION"},
	}

	for _, tt := range tests {
		got := toolVersionEnvVar(tt.tool)
		if got != tt.want {
			t.Errorf("toolVersionEnvVar(%q) = %q, want %q", tt.tool, got, tt.want)
		}
	}
}

func TestGenerator_GenerateShim_MultipleExecutables(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)

	ctx := context.Background()
	err := g.GenerateShim(ctx, "node", "node", "npm", "npx")
	if err != nil {
		t.Fatalf("failed to generate shims: %v", err)
	}

	if !g.ShimExists("npm") {
		t.Error("expected npm shim to exist")
	}
	if !g.ShimExists("npx") {
		t.Error("expected npx shim to exist")
	}
	if !g.ShimExists("nodejs") {
		t.Error("expected nodejs alias shim to exist")
	}
}

func TestExecuteBinary_NotExists(t *testing.T) {
	origGOOS := env.RuntimeGOOS
	defer func() { env.RuntimeGOOS = origGOOS }()

	env.RuntimeGOOS = "linux"
	err := ExecuteBinary(filepath.Join(t.TempDir(), "non-existent-binary"), []string{"binary"})
	if err == nil {
		t.Error("expected error when executing non-existent binary on linux")
	}

	env.RuntimeGOOS = "windows"
	err = ExecuteBinary(filepath.Join(t.TempDir(), "non-existent-binary.exe"), []string{"binary.exe"})
	if err == nil {
		t.Error("expected error when executing non-existent binary on windows")
	}
}

func TestGenerator_generateUnixShim_Fallback(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims_is_a_file")
	g := NewGenerator(shimsDir, "/installs")

	// Create a file instead of a directory to guarantee write failure cross-platform
	f, err := os.Create(shimsDir)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}
	f.Close()

	err = g.generateUnixShim("go", "go")
	if err == nil {
		t.Error("expected error when creating shim in invalid/read-only directory")
	}
}
func TestShimPaths(t *testing.T) {
	g := NewGenerator("/shims", "/installs")
	paths := g.shimPaths("go")
	if len(paths) == 0 {
		t.Error("expected at least one path")
	}
	if env.RuntimeGOOS == "windows" {
		if len(paths) != 3 {
			t.Errorf("expected 3 paths on Windows, got %d", len(paths))
		}
	} else {
		if len(paths) != 1 {
			t.Errorf("expected 1 path on Unix, got %d", len(paths))
		}
	}
}

func TestGenerator_GenerateUnixShim_Direct(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)
	err := g.generateUnixShim("go", "go")
	if err != nil {
		t.Fatalf("generateUnixShim failed: %v", err)
	}
}

func TestGenerator_GenerateWindowsShim_Direct(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)
	err := g.generateWindowsShim("go", "go")
	if err != nil {
		t.Fatalf("generateWindowsShim failed: %v", err)
	}
}

func TestGenerator_GenerateShim_SelfReferentialPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)

	ctx := context.Background()
	err := g.GenerateShim(ctx, "unirtm")
	if err == nil {
		t.Error("expected error when attempting to generate shim for unirtm itself")
	}
}

func TestIsSelfReferential(t *testing.T) {
	unirtmExe, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to get os.Executable: %v", err)
	}

	if !isSelfReferential(unirtmExe, unirtmExe, "node", "node") {
		t.Error("expected isSelfReferential to be true when shimPath == unirtmPath")
	}

	for _, blocked := range []string{"unirtm", "mise", "rtx", "asdf"} {
		if !isSelfReferential("/tmp/shims/"+blocked, "/usr/bin/unirtm", blocked, blocked) {
			t.Errorf("expected isSelfReferential to be true for tool %s", blocked)
		}
	}

	if isSelfReferential("/tmp/shims/node", "/usr/bin/unirtm", "node", "node") {
		t.Error("expected isSelfReferential to be false for distinct tool and paths")
	}
}

func TestGenerator_GenerateShim_SymlinkEvaluation(t *testing.T) {
	tmpDir := t.TempDir()
	realBin := filepath.Join(tmpDir, "real_unirtm")
	if err := os.WriteFile(realBin, []byte("binary"), 0755); err != nil {
		t.Fatalf("failed to write dummy binary: %v", err)
	}
	symlinkBin := filepath.Join(tmpDir, "symlink_unirtm")
	if err := os.Symlink(realBin, symlinkBin); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	if !isSelfReferential(symlinkBin, realBin, "node", "node") {
		t.Error("expected isSelfReferential to recognize symlink pointing to real binary")
	}
}

func TestGenerator_GenerateWindowsShim_AlreadyWithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	shimsDir := filepath.Join(tmpDir, "shims")
	installsDir := filepath.Join(tmpDir, "installs")

	g := NewGenerator(shimsDir, installsDir)
	_ = os.MkdirAll(shimsDir, 0755)
	// Simulate legacy buggy files already present
	_ = os.WriteFile(filepath.Join(shimsDir, "node.exe.exe"), []byte("buggy"), 0755)
	_ = os.WriteFile(filepath.Join(shimsDir, "node.exe.cmd"), []byte("buggy"), 0755)

	// Calling with node.exe should create node.cmd/node.exe, NOT node.exe.exe or node.exe.cmd
	err := g.generateWindowsShim("node", "node.exe")
	if err != nil {
		t.Fatalf("generateWindowsShim failed: %v", err)
	}

	// Verify legacy files were cleaned up and no double extension exists
	if _, err := os.Stat(filepath.Join(shimsDir, "node.exe.exe")); err == nil {
		t.Error("node.exe.exe should have been cleaned up")
	}
	if _, err := os.Stat(filepath.Join(shimsDir, "node.exe.cmd")); err == nil {
		t.Error("node.exe.cmd should have been cleaned up")
	}

	err = g.generateWindowsShim("node", "npm.cmd")
	if err != nil {
		t.Fatalf("generateWindowsShim for npm.cmd failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(shimsDir, "npm.cmd.exe")); err == nil {
		t.Error("npm.cmd.exe should not have been created")
	}
	if _, err := os.Stat(filepath.Join(shimsDir, "npm.cmd.cmd")); err == nil {
		t.Error("npm.cmd.cmd should not have been created")
	}
}
