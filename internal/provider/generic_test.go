// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGenericProvider_Name(t *testing.T) {
	p := NewGenericProvider()
	if p.Name() != "generic" {
		t.Errorf("expected generic, got %s", p.Name())
	}
}

func TestGenericProvider_CalculateExeScore(t *testing.T) {
	p := NewGenericProvider()

	tests := []struct {
		name     string
		toolName string
		minScore int
	}{
		{"tool", "tool", 80},
		{"tool.exe", "tool", 80},
		{"tool-cli", "tool", 30},
		{"other", "tool", 0},
		{"tool.md", "tool", -100},
		{"tool.txt", "tool", -100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := p.calculateExeScore(tc.name, tc.toolName)
			if score < tc.minScore {
				t.Errorf("expected score >= %d for %s, got %d", tc.minScore, tc.name, score)
			}
		})
	}
}

func TestGenericProvider_IsExecutable(t *testing.T) {
	p := NewGenericProvider()

	tmpDir := t.TempDir()

	exePath := filepath.Join(tmpDir, "testexe")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}

	f, err := os.Create(exePath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	if runtime.GOOS != "windows" {
		os.Chmod(exePath, 0755)
	}

	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatal(err)
	}

	if !p.isExecutable(info) {
		t.Errorf("expected %s to be recognized as executable", exePath)
	}

	txtPath := filepath.Join(tmpDir, "test.txt")
	f2, _ := os.Create(txtPath)
	f2.Close()
	info2, _ := os.Stat(txtPath)

	if p.isExecutable(info2) {
		t.Errorf("expected %s to NOT be recognized as executable", txtPath)
	}
}

func TestGenericProvider_ListExecutables_Empty(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	execs, err := p.ListExecutables("tool", tmpDir, "1.0")
	if err == nil {
		t.Error("expected error when no executables found")
	}
	if len(execs) != 0 {
		t.Errorf("expected 0 executables, got %d", len(execs))
	}
}

func TestGenericProvider_GetBinPaths(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "bin"), 0755)

	paths, err := p.GetBinPaths("tool", tmpDir, "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 paths (root and bin), got %d", len(paths))
	}
}

func TestGenericProvider_GetEnvVars(t *testing.T) {
	p := NewGenericProvider()
	env, err := p.GetEnvVars("tool", "/tmp", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(env))
	}
}

func TestGenericProvider_GenerateShims(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	exePath := filepath.Join(binDir, "mytool")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	f, _ := os.Create(exePath)
	f.Close()
	if runtime.GOOS != "windows" {
		os.Chmod(exePath, 0755)
	}

	shims, err := p.GenerateShims("mytool", tmpDir, "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	exeName := filepath.Join("bin", "mytool")
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	if content, ok := shims[exeName]; !ok {
		t.Errorf("expected shim for %s, got shims: %v", exeName, shims)
	} else if len(content) == 0 {
		t.Error("expected non-empty shim content")
	}
}

func TestGenericProvider_Uninstall(t *testing.T) {
	p := NewGenericProvider()
	err := p.Uninstall(context.Background(), "tool", "/tmp", "1.0")
	if err != nil {
		t.Errorf("expected no error from Uninstall, got %v", err)
	}
}
