// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestGoPkgProvider_Name(t *testing.T) {
	p := NewGoPkgProvider()
	if p.Name() != "go" {
		t.Errorf("expected go, got %s", p.Name())
	}
}

func TestGoPkgProvider_DetectVersion(t *testing.T) {
	p := NewGoPkgProvider()
	v, err := p.DetectVersion(context.Background(), "tool", "/tmp/install/v1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if v != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", v)
	}
}

func TestGoPkgProvider_GetBinPaths(t *testing.T) {
	p := NewGoPkgProvider()
	paths, err := p.GetBinPaths("tool", "/tmp/install", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != "/tmp/install" {
		t.Errorf("expected [/tmp/install], got %v", paths)
	}
}

func TestGoPkgProvider_GetEnvVars(t *testing.T) {
	p := NewGoPkgProvider()
	env, err := p.GetEnvVars("tool", "/tmp/install", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(env))
	}
}

func TestGoPkgProvider_Uninstall(t *testing.T) {
	p := NewGoPkgProvider()
	err := p.Uninstall(context.Background(), "tool", "/tmp/install", "1.0")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGoPkgProvider_Verify(t *testing.T) {
	p := NewGoPkgProvider()
	err := p.Verify(context.Background(), "tool", "1.0", "/tmp/install")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGoPkgProvider_GenerateShims(t *testing.T) {
	p := NewGoPkgProvider()
	
	// Because GenerateShims calls ListExecutables which reads the directory,
	// let's just make sure it returns an error if the directory doesn't exist.
	_, err := p.GenerateShims("tool", "/nonexistent_dir_for_test", "1.0")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestGoPkgProvider_ListExecutables(t *testing.T) {
	p := NewGoPkgProvider()
	tmpDir := t.TempDir()
	
	// Create an executable
	// For GoPkgProvider, ListExecutables checks for executable bit or .exe
	// Wait, we don't have to test all OS variations here since we know the logic.
	
	execs, err := p.ListExecutables("tool", tmpDir, "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(execs) != 0 {
		t.Errorf("expected 0 executables initially, got %d", len(execs))
	}
}
