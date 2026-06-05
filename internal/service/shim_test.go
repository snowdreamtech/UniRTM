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
		if len(paths) != 2 {
			t.Errorf("expected 2 paths on Windows, got %d", len(paths))
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
