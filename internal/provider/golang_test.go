// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGolangProvider_Name(t *testing.T) {
	p := NewGolangProvider()
	if p.Name() != "go" {
		t.Errorf("expected go, got %s", p.Name())
	}
}

func TestGolangProvider_PostInstall(t *testing.T) {
	p := NewGolangProvider()
	tmpDir := t.TempDir()
	
	err := p.PostInstall(context.Background(), "go", tmpDir, "1.22.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check if GOPATH directories were created
	gopath := filepath.Join(tmpDir, "gopath")
	dirs := []string{"bin", "pkg", "src"}
	for _, dir := range dirs {
		if _, err := filepath.Abs(filepath.Join(gopath, dir)); err != nil {
			t.Errorf("expected directory %s to exist", dir)
		}
	}
}

func TestGolangProvider_ListExecutables(t *testing.T) {
	p := NewGolangProvider()
	execs, err := p.ListExecutables("go", "/tmp", "1.22.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(execs) != 2 {
		t.Errorf("expected 2 executables, got %d", len(execs))
	}
}

func TestGolangProvider_GetBinPaths(t *testing.T) {
	p := NewGolangProvider()
	paths, err := p.GetBinPaths("go", "/tmp/go1.22", "1.22.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(paths))
	}
	expected := filepath.Join("/tmp/go1.22", "bin")
	if paths[0] != expected {
		t.Errorf("expected %s, got %s", expected, paths[0])
	}
}

func TestGolangProvider_GetEnvVars(t *testing.T) {
	p := NewGolangProvider()
	
	vars, err := p.GetEnvVars("go", "/tmp/go1.22", "1.22.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// By default GO_SET_GOROOT is not "0", so GOROOT should be set
	if val, ok := vars["GOROOT"]; !ok || val != "/tmp/go1.22" {
		t.Errorf("expected GOROOT=/tmp/go1.22, got %s", val)
	}
}

func TestGolangProvider_GenerateShims(t *testing.T) {
	p := NewGolangProvider()
	shims, err := p.GenerateShims("go", "/tmp/go1.22", "1.22.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shims) != 2 {
		t.Errorf("expected 2 shims, got %d", len(shims))
	}
}

func TestGolangProvider_Uninstall(t *testing.T) {
	p := NewGolangProvider()
	tmpDir := t.TempDir()
	
	err := p.PostInstall(context.Background(), "go", tmpDir, "1.22.0")
	if err != nil {
		t.Fatalf("failed to run PostInstall: %v", err)
	}
	
	err = p.Uninstall(context.Background(), "go", tmpDir, "1.22.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
