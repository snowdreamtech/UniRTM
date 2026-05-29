// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCondaProvider_Interface(t *testing.T) {
	var _ Provider = (*CondaProvider)(nil)
}

func TestCondaProvider_FindConda(t *testing.T) {
	p := NewCondaProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	scriptPath := filepath.Join(binDir, "conda")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho conda"), 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	conda, err := p.findConda()
	if err != nil {
		t.Fatalf("expected to find conda, but got error: %v", err)
	}
	if conda != scriptPath {
		t.Errorf("expected %s, got %s", scriptPath, conda)
	}
}

func TestCondaProvider_Install(t *testing.T) {
	p := NewCondaProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	scriptPath := filepath.Join(binDir, "conda")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho installing..."), 0755)

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "python", installPath, "", "3.9.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}

func TestCondaProvider_Install_CondaNotFound(t *testing.T) {
	t.Setenv("PATH", "")

	p := NewCondaProvider()
	installPath := filepath.Join(t.TempDir(), "conda_install", "test_pkg")

	err := p.Install(context.Background(), "test_pkg", installPath, "", "1.0.0")
	if err == nil {
		t.Fatalf("expected error when conda is not found")
	}
}

func TestCondaProvider_ListExecutables(t *testing.T) {
	p := NewCondaProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	os.WriteFile(filepath.Join(binDir, "dummy1"), []byte(""), 0755)
	os.WriteFile(filepath.Join(binDir, "dummy2"), []byte(""), 0644)

	exes, err := p.ListExecutables("test_pkg", tmpDir, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exes) != 1 {
		t.Errorf("expected 1 executable, got %d", len(exes))
	}
}
